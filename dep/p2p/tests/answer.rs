use std::collections::HashMap;
use std::process;
use std::sync::Arc;

use anyhow::anyhow;
use bytes::Bytes;
use http_body_util::Empty;
use hyper::{Request, StatusCode};
use log::error;
use tokio::io;
use tokio::sync::Mutex;
use webrtc::api::interceptor_registry::register_default_interceptors;
use webrtc::api::media_engine::MediaEngine;
use webrtc::api::setting_engine::SettingEngine;
use webrtc::api::APIBuilder;
use webrtc::data::data_channel::PollDataChannel;
use webrtc::data_channel::RTCDataChannel;
use webrtc::ice_transport::ice_candidate::{RTCIceCandidate, RTCIceCandidateInit};
use webrtc::ice_transport::ice_server::RTCIceServer;
use webrtc::interceptor::registry::Registry;
use webrtc::peer_connection::configuration::RTCConfiguration;
use webrtc::peer_connection::peer_connection_state::RTCPeerConnectionState;
use webrtc::peer_connection::sdp::session_description::RTCSessionDescription;

use common::compat::*;
use common::log::*;
use p2p::*;

mod common;

#[test]
fn answer_works() {
    init();
    let (server_writer, client_reader) = io::duplex(8 * 1024);
    let (server_reader, client_writer) = io::duplex(8 * 1024);

    let reader = Arc::new(Mutex::new(client_reader));
    let writer = Arc::new(Mutex::new(client_writer));
    let rt = tokio::runtime::Runtime::new().unwrap();
    rt.spawn(async move {
        if let Err(e) = process(server_reader, server_writer).await {
            eprintln!("process: {}", e);
            process::exit(1);
        };
    });
    rt.block_on(async move {
        let op = OP::Config(Config {
            stuns: vec!["stun:stun.l.google.com:19302".to_owned()],
            http_routes: HashMap::from([("www".to_owned(), "http://www.baidu.com".to_owned())]),
            ..Default::default()
        });
        write_json(Arc::clone(&writer), &serde_json::to_string(&op).unwrap())
            .await
            .map_err(|e| println!("write json error: {:?}", e))
            .expect("write json");

        let op = OP::GetOfferSDP {
            channel_name: "@www/uuid".to_owned(),
        };
        write_json(
            Arc::clone(&writer),
            &serde_json::to_string(&op).expect("encode op"),
        )
        .await
        .expect("write get offer to stdout");

        // begin
        let config = RTCConfiguration {
            ice_servers: vec![RTCIceServer {
                urls: vec!["stun:stun.l.google.com:19302".to_owned()],
                ..Default::default()
            }],
            ..Default::default()
        };

        let mut m = MediaEngine::default();
        m.register_default_codecs()
            .expect("register default codecs");

        let mut registry = Registry::new();

        registry =
            register_default_interceptors(registry, &mut m).expect("register default interceptors");

        let mut s = SettingEngine::default();
        s.detach_data_channels();

        let api = APIBuilder::new()
            .with_media_engine(m)
            .with_interceptor_registry(registry)
            .with_setting_engine(s)
            .build();

        let peer_connection = Arc::new(api.new_peer_connection(config).await.expect("new pc"));

        let writer_on_ice_candidate = Arc::clone(&writer);
        peer_connection.on_ice_candidate(Box::new(move |c: Option<RTCIceCandidate>| {
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

        let (done_tx, mut done_rx) = tokio::sync::mpsc::channel::<anyhow::Result<()>>(1);
        let done_tx_failed = done_tx.clone();
        peer_connection.on_peer_connection_state_change(Box::new(
            move |s: RTCPeerConnectionState| {
                println!("Peer Connection State has changed: {s}");
                match s {
                    RTCPeerConnectionState::Unspecified => {}
                    RTCPeerConnectionState::New => {}
                    RTCPeerConnectionState::Connecting => {}
                    RTCPeerConnectionState::Connected => {}
                    RTCPeerConnectionState::Disconnected => {}
                    RTCPeerConnectionState::Failed => {
                        let _ = done_tx_failed.try_send(Err(anyhow!("connection state failed")));
                    }
                    RTCPeerConnectionState::Closed => {}
                }

                Box::pin(async {})
            },
        ));

        peer_connection.on_data_channel(Box::new(move |dc: Arc<RTCDataChannel>| {
            println!("New DataChannel {} {}", dc.label(), dc.id());
            let done_tx_on_open = done_tx.clone();
            let data_channel = Arc::clone(&dc);
            data_channel.on_open(Box::new(|| {
                Box::pin(async move {
                    println!(
                        "Data channel '{}'-'{}' open. Request is being sent",
                        dc.label(),
                        dc.id()
                    );
                    let raw = dc.detach().await.expect("detach data channel");
                    let stream = Compat::new(PollDataChannel::new(raw));
                    let (mut sender, conn) = hyper::client::conn::http1::handshake(stream)
                        .await
                        .expect("handshake");
                    tokio::task::spawn(async move {
                        if let Err(err) = conn.await {
                            println!("Connection failed: {:?}", err);
                        }
                    });

                    let req = Request::builder()
                        .header(hyper::header::HOST, "www.baidu.com")
                        .method("GET")
                        .body(Empty::<Bytes>::new())
                        .expect("build request");

                    let res = sender.send_request(req).await.expect("send request");
                    // let mut res = sender.send_request(req).await.expect("send request");
                    // println!("Response: {}", res.status());
                    // while let Some(next) = res.frame().await {
                    //     let frame = next.expect("read frame");
                    //     if let Some(chunk) = frame.data_ref() {
                    //         io::stderr().write_all(chunk).await.expect("write chunk");
                    //     }
                    // }
                    // println!("\n");
                    if res.status() == StatusCode::OK {
                        let _ = done_tx_on_open.try_send(Ok(()));
                    }
                })
            }));
            Box::pin(async {})
        }));

        loop {
            let json = tokio::select! {
                Ok(json) = read_json(Arc::clone(&reader)) => json.clone(),
                result = done_rx.recv() => {
                    match result {
                        Some(r) => {
                            if let Err(err) = r {
                                println!("received pc failed signal: {}", err);
                            } else {
                                // received 200 http response
                                println!("received 200 success signal");
                                return;
                            }
                        }
                        None => {
                            println!("received pc failed signal!");
                        }
                    }
                    process::exit(1);
                },
                else => {
                    process::exit(2);
                }
            };
            let op = serde_json::from_str::<OP>(&json).expect("parse op json");

            let pc = Arc::clone(&peer_connection);
            match op {
                OP::OfferSDP(sdp) => {
                    let sdp = serde_json::from_str::<RTCSessionDescription>(&sdp)
                        .expect("offer sdp from op");
                    pc.set_remote_description(sdp)
                        .await
                        .expect("set remote description");
                    let answer = pc.create_answer(None).await.expect("create answer");
                    let sdp = serde_json::to_string(&answer).expect("serialize answer");
                    let op = OP::AnswerSDP(sdp);
                    write_json(
                        Arc::clone(&writer),
                        &serde_json::to_string(&op).expect("encode op"),
                    )
                    .await
                    .expect("write answer sdp to stdout");
                    pc.set_local_description(answer)
                        .await
                        .expect("set local description");
                }
                OP::Candidate(candidate) => {
                    if candidate.is_empty() {
                        continue;
                    }
                    let candidate = serde_json::from_str::<RTCIceCandidateInit>(&candidate)
                        .expect("candidate from op");
                    pc.add_ice_candidate(candidate)
                        .await
                        .expect("add candidate")
                }
                _ => {
                    panic!("invalid op {:?}", op)
                }
            };
        }
    });
}
