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
 use std::future::Future;
 use std::pin::Pin;
 use std::sync::atomic::{AtomicUsize, Ordering};
 use std::sync::Arc;
 use std::time::Duration;
 
 use anyhow::{anyhow, bail, Context, Result};
 use log::*;
 use tokio::io::{AsyncReadExt, AsyncWriteExt};
 use tokio::net::TcpStream;
 use tokio::sync::Mutex;
 use tokio::{io, select, time};
 use url::Url;
 use webrtc::api::interceptor_registry::register_default_interceptors;
 use webrtc::api::media_engine::MediaEngine;
 use webrtc::api::setting_engine::SettingEngine;
 use webrtc::api::APIBuilder;
 use webrtc::data::data_channel::PollDataChannel;
 use webrtc::data_channel::RTCDataChannel;
 use webrtc::ice::udp_network;
 use webrtc::ice::udp_network::UDPNetwork;
 use webrtc::ice_transport::ice_candidate::{RTCIceCandidate, RTCIceCandidateInit};
 use webrtc::ice_transport::ice_server::RTCIceServer;
 use webrtc::interceptor::registry::Registry;
 use webrtc::peer_connection::configuration::RTCConfiguration;
 use webrtc::peer_connection::peer_connection_state::RTCPeerConnectionState;
 use webrtc::peer_connection::sdp::session_description::RTCSessionDescription;
 use webrtc::peer_connection::RTCPeerConnection;
 use reqwest::{Client, header};

 use crate::peer::{read_json, write_json, LibError, OP, Config, ConnectConfig};
 
 pub(crate) struct ConnectPeerConnHandler<R, W> {
     http_routes: HashMap<String, String>,
     tcp_routes: HashMap<String, String>,
     reader: Arc<Mutex<R>>,
     writer: Arc<Mutex<W>>,
     channel_count: AtomicUsize,
     no_channel_id: AtomicUsize,
     peer_connection: Arc<RTCPeerConnection>,
     timeout: u16,
 }
 
 impl<R, W> ConnectPeerConnHandler<R, W>
 where
     R: AsyncReadExt + Unpin + Send + 'static,
     W: AsyncWriteExt + Unpin + Send + 'static,
 {
     pub async fn new(reader: R, writer: W) -> Result<Arc<Self>> {
         let reader = Arc::new(Mutex::new(reader));
         let writer = Arc::new(Mutex::new(writer));
         // let json = timeout(Duration::from_secs(5), read_json(Arc::clone(&reader)))
         //     .await
         //     .context("read config json timeout")?
         //     .context("read config json")?;
         // debug!("config json: {}", &json);
         // let op = serde_json::from_str::<OP>(&json)
         //     .with_context(|| format!("deserialize config json failed: {}", json))?;
         let op: OP = OP::Config(Config {
             stuns: vec!["stun:127.0.0.1:3478".to_owned()],
             http_routes: HashMap::from([("@".to_owned(), "http://www.baidu.com".to_owned())]),
             ..Default::default()
         });
         // write json config to stdout
         let output = Arc::new(Mutex::new(tokio::io::stdout()));
         write_json(Arc::clone(&output), &serde_json::to_string(&op).unwrap())
             .await
             .map_err(|e| println!("write json error: {:?}", e))
             .expect("write json");
         // bind origin op to config
         let config = match op {
             OP::Config(config) => config,
             _ => {
                 bail!("invalid config json.");
             }
         };
         // init webrtc configuration
         let rtc_config = RTCConfiguration {
             ice_servers: vec![RTCIceServer {
                 urls: config.stuns,
                 ..Default::default()
             }],
             ..Default::default()
         };
 
         // configure media engine
         let mut m = MediaEngine::default();
         m.register_default_codecs()
             .context("register default codecs")?;
 
         // register default registry
         let mut registry = Registry::new();
 
         registry = register_default_interceptors(registry, &mut m)
             .context("register default interceptors")?;
 
         // set min_port and max_port
         let mut s = SettingEngine::default();
         s.set_udp_network(UDPNetwork::Ephemeral(
             udp_network::EphemeralUDP::new(config.port_min, config.port_max)
                 .context("create udp network")?,
         ));
         // first detach data channel
         s.detach_data_channels();
 
         // build api with configuration
         let api = APIBuilder::new()
             .with_media_engine(m)
             .with_interceptor_registry(registry)
             .with_setting_engine(s)
             .build();
 
         // create pc connection
         let peer_connection = Arc::new(
             api.new_peer_connection(rtc_config)
                 .await
                 .context("new pc")?,
         );
 
         // config max timeout value
         let timeout = config.timeout.max(5);
         // create pc success
         Ok(Arc::new(ConnectPeerConnHandler {
             reader,
             writer,
             peer_connection,
             timeout,
             http_routes: config.http_routes,
             tcp_routes: config.tcp_routes,
             channel_count: Default::default(),
             no_channel_id: Default::default(),
         }))
     }

     pub async fn send_http_request(self: Arc<Self>,
        url: &str,
        method: &str,
        host: Option<&str>,
        headers: Option<Vec<(String, String)>>,
        body: Option<&str>,
    ) -> Result<String> {
        let client = Client::new();
    
        // 创建请求构建器
        let mut request_builder = match method.to_uppercase().as_str() {
            "GET" => client.get(url),
            "POST" => client.post(url),
            "PUT" => client.put(url),
            "DELETE" => client.delete(url),
            _ => todo!(),
        };
    
        // 如果提供了 host，则设置 Host 头
        if let Some(host_value) = host {
            request_builder = request_builder.header(header::HOST, host_value);
        }
    
        // 添加其他自定义头
        if let Some(custom_headers) = headers {
            for (key, value) in custom_headers {
                request_builder = request_builder.header(key, value);
            }
        }
    
        // 如果提供了 body，则添加到请求中
        if let Some(body_content) = body {
            request_builder = request_builder.body(body_content.to_string());
        }
    
        // 发送请求并获取响应
        let response = request_builder.send().await?;
    
        // 检查状态码
        if !response.status().is_success() {
            error!("HTTP error: {}", response.status());
        }
    
        // 获取响应体
        let body = response.text().await?;
    
        Ok(body)
    }
    
    pub async fn forward_data_with_server(self: Arc<Self>, yaml: &str) -> Result<bool> {
        let ya = serde_yaml::from_str::<ConnectConfig>(yaml)?;
        let url = ya.options.tcp_forward_addr;
        let method = "GET";
        let host = Some(ya.options.tcp_forward_host_prefix);
        let headers = Some(vec![
            ("Users-Agent".to_string(), "gt-connect".to_string()),
        ]);
        let body = None;
    
        let resp = self.send_http_request(&url, method, host.as_deref(), headers, body).await?;
        info!("Response from remote: {}", resp);
        Ok(true)
    }
 
     pub async fn send_offer(self: Arc<Self>) -> Result<()> {
         let pc = Arc::clone(&self.peer_connection);
         let offer = pc.create_offer(None).await.context("create offer")?;
         let sdp = serde_json::to_string(&offer).context("serialize answer")?;
         let op = OP::OfferSDP(sdp);
         write_json(
             Arc::clone(&self.writer),
             &serde_json::to_string(&op).context("encode op")?,
         )
         .await
         .context("write answer sdp to stdout")?;
         pc.set_local_description(offer)
             .await
             .context("set local description")?;
         Ok(())
     }
 
     fn setup_data_channel(self: Arc<Self>, d: Arc<RTCDataChannel>) {
         let dc = Arc::clone(&d);
         d.on_open(Box::new(|| {
             self.channel_count.fetch_add(1, Ordering::Relaxed);
             self.new_data_channel_process_handler(dc)
         }));
     }
 
     fn new_data_channel_process_handler(
         self: Arc<Self>,
         d: Arc<RTCDataChannel>,
     ) -> Pin<Box<impl Future<Output = ()> + Sized>> {
         Box::pin(async move {
             let label = d.label();
             info!("data channel '{}'-'{}' open.", label, d.id());
             let target = label.split_once('/').map_or_else(
                 || self.http_routes.get("@"),
                 |(t, _)| {
                     t.get(0..1).map_or_else(
                         || self.http_routes.get("@"),
                         |c| {
                             t.get(1..).map_or_else(
                                 || self.http_routes.get(t),
                                 |r| {
                                     if c == "@" && !r.is_empty() {
                                         self.http_routes.get(r)
                                     } else if c == ":" && !r.is_empty() {
                                         self.tcp_routes.get(r)
                                     } else {
                                         self.http_routes.get(t)
                                     }
                                 },
                             )
                         },
                     )
                 },
             );
             if let Some(target) = target {
                 info!("{} connect to {}", label, target);
                 let dc = Arc::clone(&d);
                 if let Err(err) = self.connect_target(target, dc).await {
                     info!("{} failed to connect to {}: {}", label, target, err);
                 }
             } else {
                 error!("no routes for {}", label);
             }
             info!("data channel '{}'-'{}' done.", label, d.id());
             let _ = self
                 .channel_count
                 .fetch_update(Ordering::Release, Ordering::Relaxed, |v| {
                     if v == 1 {
                         self.no_channel_id.fetch_add(1, Ordering::Relaxed);
                     }
                     Some(v - 1)
                 });
         })
     }
 
     async fn connect_target(&self, target: &str, d: Arc<RTCDataChannel>) -> Result<()> {
         let url = Url::parse(target).context("invalid url")?;
         let addrs = url
             .socket_addrs(|| match url.scheme() {
                 "http" | "ws" | "tcp" => Some(80),
                 "https" | "wss" | "tls" => Some(443),
                 _ => Some(80),
             })
             .context("no address")?;
         let raw = d.detach().await.context("detach data channel")?;
 
         let mut s = TcpStream::connect(&*addrs)
             .await
             .context("connect to service")?;
         let result = io::copy_bidirectional(&mut PollDataChannel::new(raw), &mut s).await;
         match result {
             Ok((a, b)) => {
                 info!("{} copy done: {}, {}", d.label(), a, b);
             }
             Err(err) => {
                 error!("{} copy err: {}", d.label(), err);
                 bail!(err);
             }
         }
         Ok(())
     }
 
     pub async fn handle(self: Arc<Self>) -> Result<()> {
         let writer_on_ice_candidate = Arc::clone(&self.writer);
         self.peer_connection
             // register ice_candidate process function
             .on_ice_candidate(Box::new(move |c: Option<RTCIceCandidate>| {
                 info!("on_ice_candidate {:?}", c);
                 let writer_on_ice_candidate = Arc::clone(&writer_on_ice_candidate);
                 Box::pin(async move {
                     if let Some(c) = c {
                         let json = match c.to_json() {
                             Err(e) => {
                                 error!("failed to serialize ice candidate: {}", e);
                                 return;
                             }
                             Ok(json) => json,
                         };
                         let json = match serde_json::to_string(&json) {
                             Err(e) => {
                                 error!("failed to serialize ice candidate init: {}", e);
                                 return;
                             }
                             Ok(json) => json,
                         };
                         let op = OP::Candidate(json);
                         let json = match serde_json::to_string(&op) {
                             Err(e) => {
                                 error!("failed to serialize op: {}", e);
                                 return;
                             }
                             Ok(json) => json,
                         };
                         if let Err(e) = write_json(writer_on_ice_candidate, &json).await {
                             error!("failed to write ice candidate: {}", e);
                         }
                     } else {
                         // build candidate with default value
                         let op = OP::Candidate("".to_owned());
                         let json = match serde_json::to_string(&op) {
                             Err(e) => {
                                 error!("failed to serialize op: {}", e);
                                 return;
                             }
                             Ok(json) => json,
                         };
                         if let Err(e) = write_json(writer_on_ice_candidate, &json).await {
                             error!("failed to write ice candidate: {}", e);
                         }
                     }
                 })
             }));
 
         let (done_tx, mut done_rx) = tokio::sync::mpsc::channel::<Result<()>>(1);
 
         self.peer_connection
             // register pc state change function
             .on_peer_connection_state_change(Box::new(move |s: RTCPeerConnectionState| {
                 info!("peer Connection State has changed: {s}");
                 match s {
                     RTCPeerConnectionState::Unspecified => {
                         // 未指定状态，通常不需要特殊处理
                         info!("Connection state unspecified");
                     }
                     RTCPeerConnectionState::New => {
                         // 新建连接，记录初始ICE连接状态
                         info!("New peer connection established");
                     }
                     RTCPeerConnectionState::Connecting => {
                         // 正在建立连接
                         info!("Establishing peer connection...");
                     }
                     RTCPeerConnectionState::Connected => {
                         // 连接成功建立
                         info!("Peer connection successfully established");
                     }
                     RTCPeerConnectionState::Disconnected => {
                         // 连接断开，可以尝试重连
                         warn!("Peer connection disconnected, may attempt reconnection");
                         // 可以在这里添加重连逻辑
                     }
                     RTCPeerConnectionState::Failed => {
                         // 连接失败，发送错误信号
                         error!("Peer connection failed");
                         let _ = done_tx.try_send(Err(anyhow!("peer connection state failed")));
                     }
                     RTCPeerConnectionState::Closed => {
                         // 连接已关闭
                         info!("Peer connection closed");
                         // 可以在这里进行清理工作
                         let _ = done_tx.try_send(Ok(()));
                     }
                 }
 
                 Box::pin(async {})
             }));
 
         let handler = Arc::clone(&self);
         self.peer_connection
             // register data channel build function
             .on_data_channel(Box::new(move |d: Arc<RTCDataChannel>| {
                 info!("new dataChannel {} {}", d.label(), d.id());
                 let handler = Arc::clone(&handler);
                 handler.setup_data_channel(d);
                 Box::pin(async {})
             }));
 
         let mut no_channel_id: usize = 0;
         loop {
             let sleep = time::sleep(Duration::from_secs(self.timeout as u64));
             tokio::pin!(sleep);
             let json = select! {
                 result = read_json(Arc::clone(&self.reader)) => {
                     result?
                 },
                 rx = done_rx.recv() => {
                     return match rx {
                         None => {
                             Ok(())
                         }
                         Some(result) => {
                             result
                         }
                     }
                 }
                  _ = &mut sleep => {
                     if self.channel_count.load(Ordering::Acquire) == 0 {
                         let id = self.no_channel_id.load(Ordering::Relaxed);
                         if no_channel_id == id {
                             return Err(LibError::NoChannelInPeerConnectionTimeout.into());
                         } else {
                             no_channel_id = id;
                         }
                     }
                     continue;
                 }
             };
             debug!("op json: {}", &json);
             let op = serde_json::from_str::<OP>(&json)
                 .with_context(|| format!("parse op json: {}", json))?;
 
             let pc = Arc::clone(&self.peer_connection);
             match op {
                 // receive offer from remote, return answer to remote
                 OP::OfferSDP(sdp) => {
                     let sdp = serde_json::from_str::<RTCSessionDescription>(&sdp)
                         .context("offer sdp from op")?;
                     pc.set_remote_description(sdp)
                         .await
                         .context("set remote description")?;
                     let answer = pc.create_answer(None).await.context("create answer")?;
                     let sdp = serde_json::to_string(&answer).context("serialize answer")?;
                     let op = OP::AnswerSDP(sdp);
                     write_json(
                         Arc::clone(&self.writer),
                         &serde_json::to_string(&op).context("encode op")?,
                     )
                     .await
                     .context("write answer sdp to stdout")?;
                     pc.set_local_description(answer)
                         .await
                         .context("set local description")?;
                 }
                 // receive candidate from remote, add candidate
                 OP::Candidate(candidate) => {
                     if candidate.is_empty() {
                         continue;
                     }
                     let candidate = serde_json::from_str::<RTCIceCandidateInit>(&candidate)
                         .context("candidate from op")?;
                     pc.add_ice_candidate(candidate)
                         .await
                         .context("add candidate")?;
                 }
                 // create data channel and local offer
                 OP::GetOfferSDP { channel_name } => {
                     let data_channel = pc
                         .create_data_channel(&channel_name, None)
                         .await
                         .context("create data channel")?;
                     let handler = Arc::clone(&self);
                     handler.setup_data_channel(data_channel);
                     let offer = pc.create_offer(None).await.context("create offer")?;
                     let sdp = serde_json::to_string(&offer).context("serialize answer")?;
                     let op = OP::OfferSDP(sdp);
                     write_json(
                         Arc::clone(&self.writer),
                         &serde_json::to_string(&op).context("encode op")?,
                     )
                     .await
                     .context("write answer sdp to stdout")?;
                     pc.set_local_description(offer)
                         .await
                         .context("set local description")?;
                 }
                 // receive answer from remote, set remote sdp
                 OP::AnswerSDP(sdp) => {
                     let sdp = serde_json::from_str::<RTCSessionDescription>(&sdp)
                         .context("answer sdp from op")?;
                     pc.set_remote_description(sdp)
                         .await
                         .context("set remote description")?;
                 }
                 _ => {
                     bail!("invalid op {:?}", op)
                 }
             };
         }
     }
 }
 