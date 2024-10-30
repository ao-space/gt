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

use std::collections::HashMap;
use std::sync::Arc;
use std::time::Duration;

use anyhow::{anyhow, Context, Result};
use log::*;
use serde::{Deserialize, Serialize};
use thiserror::Error;
use tokio::io;
use tokio::io::{stdin, stdout, AsyncBufReadExt, BufReader};
use tokio::sync::Mutex;
use tokio::sync::mpsc;

mod conn;
mod connect;

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
    info!("p2p done");
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

pub async fn process_connect<R, W>(reader: R, writer: W, args: ConnectConfig) -> Result<()>
where
    R: io::AsyncReadExt + Unpin + Send + 'static,
    W: io::AsyncWriteExt + Unpin + Send + 'static,
{
    let handler = connect::ConnectPeerConnHandler::new(reader, writer, args).await?;
    let _ = Arc::clone(&handler).send_offer();
    let (tx, mut rx) = mpsc::channel(8);
    // 在一个独立的任务中读取标准输入
    tokio::spawn(async move {
        let mut stdin = BufReader::new(tokio::io::stdin()).lines();
        while let Some(line) = stdin.next_line().await.unwrap() {
            tx.send(line).await.unwrap();
        }
    });
    // 在主任务中处理读取到的行
    while let Some(line) = rx.recv().await {
        println!("Received line: {}", line);
        // 将读取的行发送给服务端转发
        let handler = Arc::clone(&handler);
        match handler.forward_data_with_server(&line).await {
            Ok(_) => println!("Successfully forwarded: {}", line),
            Err(e) => eprintln!("Error forwarding data: {}", e),
        }
    }
    let _ = Arc::clone(&handler).handle().await;
    Ok(())
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

#[derive(Serialize, Deserialize, Debug, Default)]
#[serde(default)]
pub struct ConnectConfig {
    #[serde(rename = "type")]
    pub typ: String,
    pub options: ConnectOptions,
}

#[derive(Serialize, Deserialize, Debug, Default)]
#[serde(default)]
pub struct ConnectOptions {
    pub remote: String,
    pub stun_addr: String,
    pub tcp_forward_addr: String,
    pub tcp_forward_host_prefix: String,
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
    let result = String::from_utf8(buffer).context("not utf8 json")?;
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
        .context("write answer len")?;
    writer
        .write_all(json.as_bytes())
        .await
        .context("write answer")?;
    writer.flush().await.context("flush answer")?;
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
