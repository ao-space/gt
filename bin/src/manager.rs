use std::collections::{BTreeMap, HashMap};
use std::ffi::OsStr;
use std::future::Future;
use std::path::PathBuf;
use std::process::Stdio;
use std::sync::Arc;
use std::time::{Duration, Instant};
use std::{env, fs, io, process};

use anyhow::{anyhow, Context, Error, Result};
use clap::ValueEnum;
use log::{error, info, warn};
use notify::{ErrorKind, Event, PollWatcher, RecommendedWatcher, RecursiveMode, Watcher};
use serde::{Deserialize, Serialize};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::process::{ChildStdin, ChildStdout, Command};
use tokio::sync::{mpsc, oneshot, Mutex};
use tokio::time::timeout;

use crate::cs::{ClientArgs, ServerArgs};

#[derive(Debug)]
pub struct ManagerArgs {
    pub config: Option<PathBuf>,
    pub server_args: Option<ServerArgs>,
    pub client_args: Option<ClientArgs>,
}

#[derive(Debug, Copy, Clone, PartialEq, Eq, PartialOrd, Ord, ValueEnum)]
pub enum Signal {
    /// Send reload signal
    Reload,
    /// Send restart signal
    Restart,
    /// Send stop signal
    Stop,
}

enum WatcherEnum {
    Poll(PollWatcher),
    Recommended(RecommendedWatcher),
}

#[derive(Debug, PartialEq, Eq, Hash, Clone)]
enum ProcessConfigEnum {
    Config(PathBuf),
    Server(ServerArgs),
    Client(ClientArgs),
}

pub struct Manager {
    program: String,
    args: ManagerArgs,
    cmds: Arc<Mutex<HashMap<ProcessConfigEnum, Cmd>>>,
    configs: Arc<Mutex<Option<Vec<ProcessConfigEnum>>>>,
}

impl Manager {
    pub fn new(program: String, args: ManagerArgs) -> Self {
        Self {
            program,
            args,
            cmds: Arc::new(Mutex::new(HashMap::new())),
            configs: Arc::new(Mutex::new(None)),
        }
    }

    async fn collect_configs(
        &self,
    ) -> Result<(Vec<ProcessConfigEnum>, Option<Vec<ProcessConfigEnum>>)> {
        let configs;
        if let Some(args) = &self.args.server_args {
            configs = vec![ProcessConfigEnum::Server(args.clone())];
        } else if let Some(args) = &self.args.client_args {
            configs = vec![ProcessConfigEnum::Client(args.clone())];
        } else {
            let config = match &self.args.config {
                None => env::current_dir()?,
                Some(path) => path.into(),
            };
            configs = collect_files(config.clone())?;
        }
        if configs.is_empty() {
            return Err(anyhow!("no target found"));
        }
        Ok((configs.clone(), self.configs.lock().await.replace(configs)))
    }

    async fn process_signal<F>(&self, sender: F) -> Result<()>
    where
        F: for<'a> SendShutdownCallback<'a> + Copy + Send + 'static,
    {
        let (configs, old_configs) = self.collect_configs().await?;
        if let Some(old_configs) = old_configs {
            let mut shutdown = vec![];
            for p in old_configs {
                if !configs.contains(&p) {
                    shutdown.push(p);
                }
            }

            self.run_configs(configs, sender)
                .await
                .context("run_configs failed")?;

            let mut guard = self.cmds.lock().await;
            for p in shutdown {
                let cmd = guard.remove(&p);
                process_shutdown(p, cmd, sender).await;
            }
        }
        Ok(())
    }

    fn watcher<F>(&self, handler: F) -> Result<WatcherEnum>
    where
        F: FnMut(notify::Result<Event>) + Send + 'static + Clone,
    {
        let watcher = match notify::recommended_watcher(handler.clone()) {
            Ok(watcher) => WatcherEnum::Recommended(watcher),
            Err(e) => match &e.kind {
                ErrorKind::Io(io_error) => {
                    if io_error.raw_os_error() == Some(38) {
                        let watcher = PollWatcher::new(
                            handler,
                            notify::Config::default()
                                .with_poll_interval(Duration::from_secs(2))
                                .with_compare_contents(false),
                        )?;
                        WatcherEnum::Poll(watcher)
                    } else {
                        return Err(e.into());
                    }
                }
                _ => {
                    return Err(e.into());
                }
            },
        };
        Ok(watcher)
    }

    async fn run<F>(
        &self,
        configs: Vec<ProcessConfigEnum>,
        sub_cmd: &'static str,
        sender: F,
    ) -> Result<()>
    where
        F: for<'a> SendShutdownCallback<'a> + Copy + Send + 'static,
    {
        macro_rules! cmd_config {
            ($cmd:expr, $config:expr) => {
                match &$config {
                    ProcessConfigEnum::Config(path) => {
                        $cmd.arg("-c").arg(path.clone());
                    }
                    ProcessConfigEnum::Server(args) => {
                        if let Some(path) = &args.config {
                            $cmd.arg("-c").arg(path.clone());
                        }
                    }
                    ProcessConfigEnum::Client(args) => {
                        if let Some(path) = &args.config {
                            $cmd.arg("-c").arg(path.clone());
                        }
                    }
                }
                $cmd.stdin(Stdio::piped());
                $cmd.stdout(Stdio::piped());
            };
        }
        let mut cmds = configs
            .into_iter()
            .map(|config| {
                info!("run {sub_cmd} config: {:?}", config);
                let mut cmd = Command::new(self.program.clone());
                cmd.arg(sub_cmd);
                cmd_config!(cmd, config);
                cmd.spawn()
                    .context(format!("failed to start {sub_cmd} : {:?}", config))
                    .map(|c| (c, config))
            })
            .collect::<Result<Vec<_>, Error>>()?;

        for (c, config) in &mut cmds {
            let res = timeout(
                Duration::from_secs(3),
                wait_for_ready(config, &mut c.stdout),
            )
            .await;
            match res {
                Err(e) => {
                    error!("{config:?} wait for ready timeout: {e:?}");
                }
                Ok(Err(e)) => {
                    error!("{config:?} wait for ready error: {e:?}");
                }
                Ok(_) => info!("{config:?} ready"),
            }
        }

        for (mut c, config) in cmds {
            let name = self.program.clone();
            let cmds = self.cmds.clone();
            tokio::spawn(async move {
                let mut start_time = Instant::now();
                loop {
                    let stdin = c.stdin.take();
                    let stdout = c.stdout.take();
                    let (tx, rx) = oneshot::channel();
                    process_shutdown(
                        config.clone(),
                        cmds.lock().await.insert(
                            config.clone(),
                            Cmd {
                                stdin,
                                stdout,
                                tx: Some(tx),
                            },
                        ),
                        sender,
                    )
                    .await;
                    tokio::select! {
                        ref res = c.wait() => {
                            match res {
                                Ok(s) => {
                                    info!("{sub_cmd} ({config:?}) exited: {:?}", s);
                                }
                                Err(e) => {
                                    error!("{sub_cmd} ({config:?}) exited with error: {:?}", e);
                                }
                            }
                        },
                        res = rx => {
                            match res {
                                Ok(_) => {
                                    match c.kill().await {
                                        Ok(_) => info!("{sub_cmd} ({config:?}) killed"),
                                        Err(e) => error!("{sub_cmd} ({config:?}) failed to kill: {:?}", e),
                                    }
                                }
                                Err(_) => {
                                    info!("{sub_cmd} ({config:?}) stopped");
                                    match c.wait().await {
                                        Ok(s) => {
                                            info!("{sub_cmd} ({config:?}) exited: {:?}", s);
                                        }
                                        Err(e) => {
                                            error!("{sub_cmd} ({config:?}) exited with error: {:?}", e);
                                        }
                                    }
                                }
                            }
                            return;
                        }
                    }
                    let wait_time = if start_time.elapsed() < Duration::from_secs(60) {
                        warn!("{sub_cmd} ({config:?}) exited too quick");
                        Duration::from_secs(60)
                    } else {
                        Duration::from_secs(3)
                    };
                    loop {
                        info!("restarting {sub_cmd} ({config:?}) in {wait_time:?}");
                        tokio::time::sleep(wait_time).await;
                        let mut cmd = Command::new(&name);
                        cmd.arg(sub_cmd);
                        cmd_config!(cmd, config);
                        match cmd.spawn() {
                            Ok(child) => {
                                start_time = Instant::now();
                                c = child;
                                info!("restarted {sub_cmd} ({config:?})");
                                break;
                            }
                            Err(e) => {
                                error!("failed to restart {sub_cmd} ({config:?}): {:?}", e);
                            }
                        }
                    }
                }
            });
        }
        Ok(())
    }

    async fn run_configs<F>(&self, configs: Vec<ProcessConfigEnum>, sender: F) -> Result<()>
    where
        F: for<'a> SendShutdownCallback<'a> + Copy + Send + 'static,
    {
        let mut server_config = vec![];
        let mut client_config = vec![];
        for config in configs {
            match &config {
                ProcessConfigEnum::Config(path) => {
                    if is_client_config_path(path).context("is_client_config_path failed")? {
                        client_config.push(config);
                    } else {
                        server_config.push(config);
                    }
                }
                ProcessConfigEnum::Server(_) => server_config.push(config),
                ProcessConfigEnum::Client(_) => client_config.push(config),
            }
        }
        if !server_config.is_empty() {
            self.run(server_config, "sub-server", sender)
                .await
                .context("run_server failed")?;
        }

        if !client_config.is_empty() {
            self.run(client_config, "sub-client", sender)
                .await
                .context("run_client failed")?;
        }
        Ok(())
    }

    pub fn run_manager(self) -> Result<()> {
        macro_rules! process_event {
            ($path:ident, $tx:ident, $pst:ident, {$($file:literal:$sig:expr),+}) => {
                match $path.file_name().and_then(OsStr::to_str) {
                $(
                Some($file) => {
                    let now = Instant::now();
                    if let Some(process_signal_time) = $pst {
                        if now - process_signal_time < Duration::from_secs(3) {
                            info!("{} too frequently, ignored", $file);
                            return;
                        }
                    }
                    $pst = Some(now);
                    if let Err(e) = $tx.blocking_send($sig) {
                        error!("send {} event failed: {}", $file, e);
                    }
                }
                )+
                None | Some(_) => {}
                }
            };
        }
        let rp = create_signal_files()?;
        let rt = tokio::runtime::Builder::new_current_thread()
            .enable_all()
            .build()
            .unwrap();
        rt.block_on(async {
            let (configs, _) = self
                .collect_configs()
                .await
                .context("collect_files failed")?;
            self.run_configs(configs, send_nop)
                .await
                .context("run_configs failed")?;
            let (tx, mut rx) = mpsc::channel(1);
            let mut process_signal_time = None;
            let mut watcher = self
                .watcher(move |res| match res {
                    Ok(event) => {
                        info!("watch event: {:?}", event);
                        if let Some(path_buf) = event.paths.first() {
                            process_event!(path_buf, tx, process_signal_time, {"reload": Signal::Reload, "restart": Signal::Restart, "stop": Signal::Stop});
                        }
                    }
                    Err(e) => error!("watch error: {:?}", e),
                })
                .context("watch failed")?;
            match watcher {
                WatcherEnum::Recommended(ref mut watcher) => {
                    info!("recommended watcher {:?}", rp);
                    watcher.watch(&rp, RecursiveMode::NonRecursive)?;
                }
                WatcherEnum::Poll(ref mut watcher) => {
                    info!("poll watcher watching {:?}", rp);
                    watcher.watch(&rp, RecursiveMode::NonRecursive)?;
                }
            }
            loop {
                tokio::select! {
                    _ = tokio::signal::ctrl_c() => {
                        info!("ctrl_c received!");
                        break;
                    }
                    res = rx.recv() => {
                        match &res {
                            Some(Signal::Reload) => {
                                match self.process_signal(send_graceful_shutdown).await {
                                    Ok(_) => info!("reload scheduled"),
                                    Err(e) => {
                                        error!("reload error: {:?}", e);
                                    }
                                }
                            }
                            Some(Signal::Restart) => {
                                match self.process_signal(send_shutdown).await {
                                    Ok(_) => info!("restart scheduled"),
                                    Err(e) => {
                                        error!("restart error: {:?}", e);
                                    }
                                }
                            }
                            None | Some(Signal::Stop) => {
                                for (p, cmd) in self.cmds.lock().await.drain() {
                                    process_shutdown(p, Some(cmd), send_shutdown).await;
                                }
                                info!("stop scheduled");
                                break;
                            }
                        }
                    }
                }
            }
            Ok::<(), Error>(())
        })?;
        info!("run_manager done");
        rt.shutdown_timeout(Duration::from_millis(100));
        Ok(())
    }
}

async fn process_shutdown<F>(config: ProcessConfigEnum, cmd: Option<Cmd>, sender: F)
where
    F: for<'a> SendShutdownCallback<'a> + Copy + Send + 'static,
{
    match cmd {
        Some(mut cmd) => {
            let res = timeout(Duration::from_secs(10), sender(&config, &mut cmd)).await;
            let mut kill = || {
                if let Some(tx) = cmd.tx.take() {
                    info!("{config:?} being killed");
                    if let Err(e) = tx.send(()) {
                        error!("{config:?} kill channel send error: {e:?}");
                    } else {
                        info!("{config:?} kill channel sent");
                    }
                }
            };
            match res {
                Err(e) => {
                    error!("{config:?} shutdown timeout: {e:?}");
                    kill();
                }
                Ok(Err(e)) => {
                    error!("{config:?} shutdown error: {e:?}");
                    kill();
                }
                Ok(_) => info!("{config:?} shutdown"),
            }
        }
        None => info!("{config:?} no command child exists"),
    }
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase", tag = "op")]
pub enum OP {
    Ready,
    GracefulShutdown,
    GracefulShutdownDone,
    Shutdown,
    ShutdownDone,
}

async fn read_json(reader: &mut ChildStdout) -> Result<OP> {
    const MAX_JSON_LENGTH: u32 = 8 * 1024;
    let mut buffer = [0; 4];
    reader
        .read_exact(&mut buffer)
        .await
        .context("failed to receive header")?;
    let length = u32::from_be_bytes(buffer);
    if length > MAX_JSON_LENGTH {
        return Err(anyhow!("json too large: {}", length));
    }
    let mut buffer = vec![0; length as usize];
    reader
        .read_exact(&mut buffer)
        .await
        .context("failed to receive json")?;
    serde_json::from_slice::<OP>(&buffer).context("failed to parse op json")
}

async fn write_json(writer: &mut ChildStdin, op: &OP) -> Result<()> {
    let json = serde_json::to_string(op)?;
    let l = json.len() as u32;
    writer
        .write_all(&l.to_be_bytes())
        .await
        .context("write answer len")?;
    writer
        .write_all(json.as_bytes())
        .await
        .context("write answer")?;
    writer.flush().await.context("flush answer")?;
    Ok(())
}

async fn wait_for_ready(path: &ProcessConfigEnum, stdout: &mut Option<ChildStdout>) -> Result<()> {
    info!(
        "wait for cmd ready {path:?} recv: {:?}",
        read_json(stdout.as_mut().ok_or(anyhow!("no stdout"))?).await?
    );
    Ok(())
}

trait SendShutdownCallback<'a>: Fn(&'a ProcessConfigEnum, &'a mut Cmd) -> Self::Fut {
    type Fut: Future<Output = Result<()>> + Send;
}

impl<
        'a,
        Out: Future<Output = Result<()>> + Send,
        F: Fn(&'a ProcessConfigEnum, &'a mut Cmd) -> Out + Copy + Send + 'static,
    > SendShutdownCallback<'a> for F
{
    type Fut = Out;
}

async fn send_graceful_shutdown(path: &ProcessConfigEnum, c: &mut Cmd) -> Result<()> {
    send_op(path, c, OP::GracefulShutdown).await
}

async fn send_shutdown(path: &ProcessConfigEnum, c: &mut Cmd) -> Result<()> {
    send_op(path, c, OP::Shutdown).await
}

async fn send_nop(_: &ProcessConfigEnum, _: &mut Cmd) -> Result<()> {
    panic!("send_nop should not be called")
}

#[inline]
async fn send_op(path: &ProcessConfigEnum, c: &mut Cmd, op: OP) -> Result<()> {
    write_json(c.stdin.as_mut().ok_or(anyhow!("no stdin"))?, &op).await?;
    info!("signal({op:?}) {path:?} sent");
    info!(
        "signal({op:?}) {path:?} recv: {:?}",
        read_json(c.stdout.as_mut().ok_or(anyhow!("no stdout"))?).await?
    );
    Ok(())
}

struct Cmd {
    stdin: Option<ChildStdin>,
    stdout: Option<ChildStdout>,
    tx: Option<oneshot::Sender<()>>,
}

fn create_signal_files() -> Result<PathBuf> {
    let mut gt = env::temp_dir();
    gt.push("gt-manager");
    fs::create_dir_all(&gt).with_context(|| format!("failed to create {gt:?}"))?;
    gt.push("pid");
    fs::write(&gt, format!("{}\n", process::id()))
        .with_context(|| format!("failed to write {gt:?}"))?;
    gt.pop();
    Ok(gt)
}

pub fn send_signal(signal: Signal) -> Result<()> {
    let file_name = match signal {
        Signal::Reload => "reload",
        Signal::Restart => "restart",
        Signal::Stop => "stop",
    };
    let mut gt = env::temp_dir();
    gt.push("gt-manager");
    gt.push(file_name);
    let _ =
        fs::File::create(&gt).with_context(|| format!("failed to send {signal:?} to {gt:?}"))?;
    Ok(())
}

fn collect_files(path: PathBuf) -> io::Result<Vec<ProcessConfigEnum>> {
    let mut files = vec![];
    if path.is_dir() {
        for entry in fs::read_dir(path)? {
            let entry = entry?;
            let path = entry.path();
            if path.is_dir() {
                collect_files(path)?;
            } else {
                let fm = entry.metadata()?.len();
                if fm > 10 * 1024 * 1024 {
                    info!("ignored file {} is too large", path.display());
                    continue;
                }
                match path.extension().and_then(OsStr::to_str) {
                    Some("yaml") | Some("yml") => {}
                    None | Some(_) => {
                        info!(
                            "ignored file {} is not end with yml or yaml",
                            path.display()
                        );
                        continue;
                    }
                }
                info!("collected file {}", path.display());
                files.push(ProcessConfigEnum::Config(path));
            }
        }
    } else {
        info!("collected file {}", path.display());
        files.push(ProcessConfigEnum::Config(path));
    }
    Ok(files)
}

#[derive(Serialize, Deserialize, Debug)]
struct Config {
    #[serde(rename = "type")]
    typ: Option<String>,
    services: Option<Vec<BTreeMap<String, String>>>,
}

fn is_client_config_path(path: &PathBuf) -> Result<bool> {
    let yaml = fs::read_to_string(path)?;
    is_client_config(&yaml)
}

fn is_client_config(yaml: &str) -> Result<bool> {
    let c = serde_yaml::from_str::<Config>(yaml)?;
    if c.services.is_some() {
        return Ok(true);
    }
    if let Some(typ) = c.typ {
        return match typ.as_str() {
            "client" => Ok(true),
            "server" => Ok(false),
            t => Err(anyhow!("invalid config type {}", t)),
        };
    }
    Ok(false)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn is_client_config_works() {
        let cc = "
version: 1.0
services:
  # http 转发
  - local: http://127.0.0.1:80
    hostPrefix: blog
  # http 转发
  - local: http://127.0.0.1:80
    hostPrefix: web
  # https sni 转发
  - local: https://www.baidu.com
    hostPrefix: www
  # server 10022 tcp 端口转发流量到 client 本地 22 tcp 端口
  - local: tcp://127.0.0.1:22
    # 服务器端口
    remoteTCPPort: 10022
    # 如果 10022 端口被占用，则使用服务器随机端口
    remoteTCPRandom: true
options:
  id: id-should-be-overwritten
  secret: secret-should-be-overwritten
  # 服务器地址
  # remote: tls://1.1.1.1:4443
  remote: tcp://1.1.1.1:80
  # 连接池并发连接数
  remoteConnections: 5
  logLevel: info
        ";
        assert!(is_client_config(cc).unwrap());
        let cc = "
version: 1.0
options:
  addr: 80
  sniAddr: 443
  #  tlsAddr: 4443
  #  certFile: /opt/crt/tls.crt
  #  keyFile: /opt/crt/tls.key
  logLevel: info
  timeout: 90s
  stunAddr: 3478
users:
  id-should-be-overwritten:
    secret: secret-should-be-overwritten
    tcp:
      - range: 10000-15000
      - range: 20000-25000
        ";
        assert!(!is_client_config(cc).unwrap());
        let cc = "
type: client
        ";
        assert!(is_client_config(cc).unwrap());
        let cc = "
type: server
        ";
        assert!(!is_client_config(cc).unwrap());
    }

    #[test]
    fn is_op_works() {
        let json = serde_json::to_string(&OP::GracefulShutdown);
        println!("op {:?}", json);
    }
}
