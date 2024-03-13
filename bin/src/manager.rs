/*
 * Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

use std::collections::{BTreeMap, HashMap};
use std::ffi::OsStr;
use std::future::Future;
use std::io::Cursor;
#[cfg(unix)]
use std::os::unix::process::CommandExt;
#[cfg(windows)]
use std::os::windows::process::CommandExt;
use std::path::PathBuf;
use std::process::Stdio;
use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::Arc;
use std::time::{Duration, Instant};
use std::{env, fs, future, io, process};

use anyhow::{anyhow, Context, Error, Result};
use clap::ValueEnum;
use futures::future::{BoxFuture, FutureExt};
use log::{error, info, warn};
use notify::{ErrorKind, Event, PollWatcher, RecommendedWatcher, RecursiveMode, Watcher};
use serde::{de, ser, Deserialize, Serialize};
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use tokio::process::{Child, ChildStdin, ChildStdout, Command};
use tokio::sync::oneshot::{Receiver, Sender};
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

#[derive(Debug, PartialEq, Eq, Hash, Clone, Serialize, Deserialize)]
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
        F: for<'a> SendShutdownCallback<'a> + Copy + Send + Sync + 'static,
    {
        let manager = spawn_manager().await?;
        let mut stdout = manager.stdout.ok_or(anyhow!("no stdout"))?;
        let handle = async {
            loop {
                let op: MangerOP = read_hex_len_json(&mut stdout)
                    .await
                    .context("read manager ready op")?;
                info!("read manager op: {:?}", op);
                match op {
                    MangerOP::Ready(config) => match self.cmds.lock().await.remove(&config) {
                        Some(cmd) => {
                            process_shutdown(config, cmd, sender).await;
                        }
                        None => {
                            info!("no cmd found for config: {:?}", config);
                        }
                    },
                    MangerOP::ReadyDone => {
                        break;
                    }
                }
            }
            Result::<()>::Ok(())
        };
        match timeout(Duration::from_secs(120), handle).await {
            Err(e) => {
                info!("read manager op timeout: {e:?}");
            }
            Ok(Err(e)) => {
                info!("read manager op error: {e:?}");
            }
            Ok(_) => info!("read manager op done"),
        }

        let mut guard = self.cmds.lock().await;
        for (config, cmd) in guard.drain() {
            process_shutdown(config, cmd, sender).await;
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

    fn sync_run(
        program: String,
        cmd_map: Arc<Mutex<HashMap<ProcessConfigEnum, Cmd>>>,
        configs: Vec<ProcessConfigEnum>,
        sub_cmd: &'static str,
    ) -> BoxFuture<'static, Result<()>> {
        async move { Self::run(program, cmd_map, configs, sub_cmd).await }.boxed()
    }

    async fn handle_stdout(
        mut stdout: Option<ChildStdout>,
        config: ProcessConfigEnum,
        sub_cmd: &'static str,
        mut shutdown_tx: Option<Sender<()>>,
        mut reconnect: Option<impl Future<Output = ()> + Sized>,
        done_counter: Arc<AtomicUsize>,
        exited: &mut bool,
    ) {
        loop {
            let res = match stdout {
                Some(ref mut o) => read_json(o).await,
                None => future::pending().await,
            };
            match res {
                Ok(op) => match op {
                    OP::Ready => {
                        info!("{sub_cmd} ({config:?}) ready op received");
                        if 1 == done_counter.fetch_sub(1, Ordering::Relaxed) {
                            if let Err(e) =
                                write_hex_len_json(&mut tokio::io::stdout(), &MangerOP::ReadyDone)
                                    .await
                            {
                                error!("{sub_cmd} ({config:?}) write ready done op failed: {e}");
                            }
                        } else if let Err(e) = write_hex_len_json(
                            &mut tokio::io::stdout(),
                            &MangerOP::Ready(config.clone()),
                        )
                        .await
                        {
                            error!("{sub_cmd} ({config:?}) write ready op failed: {e}");
                        }
                    }
                    OP::ShutdownDone | OP::GracefulShutdownDone => {
                        if let Some(tx) = shutdown_tx.take() {
                            info!("{sub_cmd} ({config:?}) shutdown done op received");
                            match tx.send(()) {
                                Ok(_) => {
                                    info!("{sub_cmd} ({config:?}) shutdown_tx send");
                                }
                                Err(_) => {
                                    error!("{sub_cmd} ({config:?}) failed to shutdown_tx send");
                                }
                            }
                        } else {
                            error!("{sub_cmd} ({config:?}) shutdown done op received again");
                        }
                        *exited = true;
                        future::pending::<()>().await;
                    }
                    OP::Reconnect => {
                        info!("{sub_cmd} ({config:?}) reconnect op received");
                        match reconnect.take() {
                            None => error!("{sub_cmd} ({config:?}) reconnect op received again"),
                            Some(reconnect) => reconnect.await,
                        }
                        *exited = true;
                        future::pending::<()>().await;
                    }
                    _ => {
                        error!("{sub_cmd} ({config:?}) unexpected op received: {:?}", op);
                    }
                },
                Err(e) => {
                    error!("{sub_cmd} ({config:?}) read_json failed: {:?}", e);
                    future::pending::<()>().await;
                }
            }
        }
    }

    async fn run(
        program: String,
        cmd_map: Arc<Mutex<HashMap<ProcessConfigEnum, Cmd>>>,
        configs: Vec<ProcessConfigEnum>,
        sub_cmd: &'static str,
    ) -> Result<()> {
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
        let cmds = configs
            .into_iter()
            .map(|config| {
                info!("run {sub_cmd} config: {:?}", config);
                let mut cmd = Command::new(program.clone());
                cmd.arg(sub_cmd);
                cmd_config!(cmd, config);
                cmd.spawn()
                    .context(format!("failed to start {sub_cmd} : {:?}", config))
                    .map(|c| (c, config))
            })
            .collect::<Result<Vec<_>, Error>>()?;
        let ready_done_counter = Arc::new(AtomicUsize::new(cmds.len()));

        for (mut c, config) in cmds {
            let program = program.clone();
            let cmd_map = cmd_map.clone();
            let ready_done_counter = ready_done_counter.clone();
            tokio::spawn(async move {
                loop {
                    let start_time = Instant::now();
                    let stdin = c.stdin.take();
                    let stdout = c.stdout.take();
                    let (kill_tx, kill_rx) = oneshot::channel();
                    let (shutdown_tx, shutdown_rx) = oneshot::channel();
                    if let Some(cmd) = cmd_map.lock().await.insert(
                        config.clone(),
                        Cmd {
                            stdin,
                            kill_tx: Some(kill_tx),
                            shutdown_rx: Some(shutdown_rx),
                        },
                    ) {
                        process_shutdown(config.clone(), cmd, send_graceful_shutdown).await;
                    }
                    let reconnect = async {
                        let program = program.clone();
                        let cmds = cmd_map.clone();
                        let config = config.clone();
                        if let Err(e) =
                            Self::sync_run(program, cmds, vec![config.clone()], sub_cmd).await
                        {
                            error!("{sub_cmd} ({config:?}) reconnect sync_run failed: {:?}", e);
                        }
                    };
                    let mut exited = false;
                    tokio::select! {
                        _ = Self::handle_stdout(stdout, config.clone(), sub_cmd,
                                Some(shutdown_tx), Some(reconnect),
                                ready_done_counter.clone(), &mut exited
                            ) => {},
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
                        res = kill_rx => {
                            match res {
                                Ok(_) => {
                                    match c.kill().await {
                                        Ok(_) => info!("{sub_cmd} ({config:?}) killed"),
                                        Err(e) => error!("{sub_cmd} ({config:?}) failed to kill: {:?}", e),
                                    }
                                }
                                Err(_) => {
                                    info!("{sub_cmd} ({config:?}) kill_tx dropped");
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
                    if exited {
                        return;
                    }
                    let mut wait_time = if start_time.elapsed() < Duration::from_secs(60) {
                        warn!("{sub_cmd} ({config:?}) exited too quickly");
                        Duration::from_secs(60)
                    } else {
                        Duration::from_secs(3)
                    };
                    loop {
                        info!("restarting {sub_cmd} ({config:?}) in {wait_time:?}");
                        tokio::time::sleep(wait_time).await;
                        let mut cmd = Command::new(&program);
                        cmd.arg(sub_cmd);
                        cmd_config!(cmd, config);
                        match cmd.spawn() {
                            Ok(child) => {
                                c = child;
                                info!("restarted {sub_cmd} ({config:?})");
                                break;
                            }
                            Err(e) => {
                                wait_time = Duration::from_secs(3 * 60);
                                error!("failed to restart {sub_cmd} ({config:?}): {:?}", e);
                            }
                        }
                    }
                }
            });
        }
        Ok(())
    }

    async fn run_configs(&self, configs: Vec<ProcessConfigEnum>) -> Result<()> {
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
            Self::run(
                self.program.clone(),
                self.cmds.clone(),
                server_config,
                "sub-server",
            )
            .await
            .context("run_server failed")?;
        }

        if !client_config.is_empty() {
            Self::run(
                self.program.clone(),
                self.cmds.clone(),
                client_config,
                "sub-client",
            )
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
            self.run_configs(configs)
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
            tokio::select! {
                _ = tokio::signal::ctrl_c() => {
                    info!("ctrl_c received!");
                }
                res = rx.recv() => {
                    match &res {
                        Some(Signal::Reload) => {
                            info!("reload signal processing");
                            match self.process_signal(send_graceful_shutdown).await {
                                Ok(_) => info!("reload signal processed"),
                                Err(e) => {
                                    error!("reload error: {:?}", e);
                                }
                            }
                        }
                        Some(Signal::Restart) => {
                            info!("restart signal processing");
                            match self.process_signal(send_shutdown).await {
                                Ok(_) => info!("restart signal processed"),
                                Err(e) => {
                                    error!("restart error: {:?}", e);
                                }
                            }
                        }
                        None | Some(Signal::Stop) => {
                            info!("stopping");
                            for (p, cmd) in self.cmds.lock().await.drain() {
                                process_shutdown(p, cmd, send_shutdown).await;
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

async fn process_shutdown<F>(config: ProcessConfigEnum, mut cmd: Cmd, sender: F)
where
    F: for<'a> SendShutdownCallback<'a> + Copy + Send + 'static,
{
    let res = timeout(Duration::from_secs(120), sender(&config, &mut cmd)).await;
    let mut kill = || {
        if let Some(tx) = cmd.kill_tx.take() {
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

async fn spawn_manager() -> Result<Child> {
    let mut args = env::args();
    let mut cmd = process::Command::new(args.next().ok_or(anyhow!("empty args"))?);
    cmd.args(args)
        .envs(env::vars())
        .stdout(Stdio::piped())
        .stdin(Stdio::piped());
    #[cfg(windows)]
    cmd.creation_flags(0x00000200);
    #[cfg(unix)]
    cmd.process_group(0);
    let mut cmd = Command::from(cmd);
    cmd.spawn().context("spawn manager")
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase", tag = "op")]
enum MangerOP {
    Ready(ProcessConfigEnum),
    ReadyDone,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase", tag = "op")]
pub enum OP {
    Ready,
    GracefulShutdown,
    GracefulShutdownDone,
    Shutdown,
    ShutdownDone,
    Reconnect,
}

const MAX_JSON_LENGTH: u32 = 8 * 1024;

async fn read_json<R, T>(reader: &mut R) -> Result<T>
where
    R: AsyncReadExt + Unpin,
    T: de::DeserializeOwned,
{
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
    serde_json::from_reader(Cursor::new(buffer)).context("failed to parse op json")
}

async fn write_json<W, T>(writer: &mut W, op: &T) -> Result<()>
where
    W: AsyncWriteExt + Unpin,
    T: ser::Serialize,
{
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

async fn read_hex_len_json<R, T>(reader: &mut R) -> Result<T>
where
    R: AsyncReadExt + Unpin,
    T: de::DeserializeOwned,
{
    let mut buffer = [0; 8];
    reader
        .read_exact(&mut buffer)
        .await
        .context("failed to receive header")?;
    let length = u32::from_str_radix(String::from_utf8_lossy(&buffer).as_ref(), 16)?;
    if length > MAX_JSON_LENGTH {
        return Err(anyhow!("json too large: {}", length));
    }
    let mut buffer = vec![0; length as usize];
    reader
        .read_exact(&mut buffer)
        .await
        .context("failed to receive json")?;
    serde_json::from_reader(Cursor::new(buffer)).context("failed to parse op json")
}

async fn write_hex_len_json<W, T>(writer: &mut W, op: &T) -> Result<()>
where
    W: AsyncWriteExt + Unpin,
    T: ser::Serialize,
{
    let json = serde_json::to_string(op)?;
    let l = json.len() as u32;
    writer
        .write_all(format!("{:08x}", l).as_ref())
        .await
        .context("write answer len")?;
    writer
        .write_all(json.as_bytes())
        .await
        .context("write answer")?;
    writer.flush().await.context("flush answer")?;
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

#[inline]
async fn send_op(path: &ProcessConfigEnum, c: &mut Cmd, op: OP) -> Result<()> {
    write_json(c.stdin.as_mut().ok_or(anyhow!("no stdin"))?, &op).await?;
    info!("signal({op:?}) {path:?} sent");
    c.shutdown_rx
        .take()
        .ok_or(anyhow!("shutdown_rx has been taken"))?
        .await?;
    info!("signal({op:?}) {path:?} shutdown done recv");
    Ok(())
}

struct Cmd {
    stdin: Option<ChildStdin>,
    kill_tx: Option<Sender<()>>,
    shutdown_rx: Option<Receiver<()>>,
}

const TMP_FOLDER: &str = "gt-runtime";

fn create_signal_files() -> Result<PathBuf> {
    let mut gt = env::temp_dir();
    gt.push(TMP_FOLDER);
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
    gt.push(TMP_FOLDER);
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
