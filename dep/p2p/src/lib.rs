use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use anyhow::{anyhow, Context, Result};
use env_logger::Env;
use log::*;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tokio::io;
use tokio::io::{stdin, stdout};
use tokio::sync::Mutex;

mod conn;

#[no_mangle]
pub extern "C" fn create_peer_connection() {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();
    start_peer_connection();
}

pub fn start_peer_connection() {
    let rt = tokio::runtime::Builder::new_current_thread()
        .enable_all()
        .build()
        .unwrap();
    rt.block_on(async {
        match process(stdin(), stdout()).await {
            Ok(_) => {
                info!("create_peer_connection done");
            }
            Err(e) => {
                error!("create_peer_connection err: {:?}", e);
            }
        };
    });
    eprintln!("p2p done");
    rt.shutdown_timeout(Duration::from_millis(100));
}

pub async fn process<R, W>(reader: R, writer: W) -> Result<()>
where
    R: io::AsyncReadExt + Unpin + Send + 'static,
    W: io::AsyncWriteExt + Unpin + Send + 'static,
{
    let handler = conn::PeerConnHandler::new(reader, writer).await?;
    handler.handle().await
}

#[derive(Serialize, Deserialize, Debug, Default)]
#[serde(default, rename_all = "camelCase")]
pub struct Config {
    pub stuns: Vec<String>,
    pub http_routes: HashMap<String, String>,
    pub tcp_routes: HashMap<String, String>,
    pub port_min: u16,
    pub port_max: u16,
    pub timeout: u16,
}

#[derive(Serialize, Deserialize, Debug)]
#[serde(rename_all = "camelCase")]
pub enum OP {
    Config(Config),
    OfferSDP(String),
    AnswerSDP(String),
    Candidate(String),
    GetOfferSDP {
        #[serde(rename = "channelName")]
        channel_name: String,
    },
}

pub async fn read_json<R>(reader: Arc<Mutex<R>>) -> Result<String>
where
    R: io::AsyncReadExt + Unpin,
{
    const MAX_JSON_LENGTH: u32 = 8 * 1024;
    let mut buffer = [0; 4];
    let mut reader = reader.lock().await;
    reader
        .read_exact(&mut buffer)
        .await
        .with_context(|| "failed to receive header")?;
    let length = u32::from_be_bytes(buffer);
    if length > MAX_JSON_LENGTH {
        return Err(anyhow!("json too large: {}", length));
    }
    let mut buffer = vec![0; length as usize];
    reader
        .read_exact(&mut buffer)
        .await
        .with_context(|| "failed to receive json")?;
    let result = String::from_utf8(buffer).with_context(|| "not utf8 json")?;
    Ok(result)
}

pub async fn write_json<W>(writer: Arc<Mutex<W>>, json: &str) -> Result<()>
where
    W: io::AsyncWriteExt + Unpin,
{
    let mut writer = writer.lock().await;
    let l = json.len() as u32;
    writer
        .write_all(&l.to_be_bytes())
        .await
        .with_context(|| "write answer len")?;
    writer
        .write_all(json.as_bytes())
        .await
        .with_context(|| "write answer")?;
    writer.flush().await.with_context(|| "flush answer")?;
    Ok(())
}

#[derive(Error, Debug)]
pub enum LibError {
    #[error("no channel in peer connection timeout")]
    NoChannelInPeerConnectionTimeout,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[ignore]
    #[test]
    fn test_op_json() {
        let op = OP::Config(Config {
            stuns: vec!["stun:stun.l.google.com:19302".to_owned()],
            http_routes: HashMap::from([
                ("www".to_owned(), "http://www.baidu.com".to_owned()),
                ("default".to_owned(), "http://www.baidu.com".to_owned()),
            ]),
            ..Default::default()
        });
        println!("{}", serde_json::to_string(&op).unwrap());
        let op = OP::OfferSDP("abc".to_owned());
        println!("{}", serde_json::to_string(&op).unwrap());
        let op = OP::GetOfferSDP {
            channel_name: "abc".to_owned(),
        };
        println!("{}", serde_json::to_string(&op).unwrap());
    }
}
