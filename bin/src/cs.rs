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

#![allow(non_upper_case_globals)]
#![allow(non_camel_case_types)]
#![allow(non_snake_case)]
#![allow(unused)]

use clap::Args;
use log::info;
use serde::{Deserialize, Serialize};
use std::{ffi::{c_char, c_void, CString}, fmt::Debug, process::ExitCode};
include!("cs_bindings.rs");

use std::fs;
use std::path::Path;
use std::process;
use std::sync::Arc;
use tokio::sync::Mutex;
use tokio::io::{AsyncReadExt, AsyncWriteExt};
use crate::peer::*;

#[derive(Args, Debug, PartialEq, Eq, Hash, Clone, Serialize, Deserialize)]
pub struct ServerArgs {
    /// Config file path
    #[arg(short, long)]
    pub config: Option<String>,
}

#[derive(Args, Debug, PartialEq, Eq, Hash, Clone, Serialize, Deserialize)]
pub struct ClientArgs {
    /// Config file path
    #[arg(short, long)]
    pub config: Option<String>,
}

#[derive(Args, Debug, PartialEq, Eq, Hash, Clone, Serialize, Deserialize)]
pub struct ConnectArgs {
    /// Config file path
    #[arg(short, long)]
    pub config: Option<String>,
}

fn convert_to_go_slices(vec: &Vec<String>) -> (GoSlice, Vec<GoString>) {
    let mut go_slices: Vec<GoString> = Vec::with_capacity(vec.len());

    for arg in vec {
        let go_string = GoString {
            p: arg.as_ptr() as *const c_char,
            n: arg.as_bytes().len() as isize,
        };
        go_slices.push(go_string);
    }
    (
        GoSlice {
            data: go_slices.as_mut_ptr() as *mut c_void,
            len: go_slices.len() as GoInt,
            cap: go_slices.len() as GoInt,
        },
        go_slices,
    )
}

fn load_config(config_path: &str) -> Result<ConnectConfig, Box<dyn std::error::Error>> {
    // 验证文件是否存在
    if !Path::new(config_path).exists() {
        return Err(format!("Config file '{}' does not exist", config_path).into());
    }

    // 读取文件内容
    let config_content = fs::read_to_string(config_path)
        .map_err(|e| format!("Failed to read config file '{}': {}", config_path, e))?;

    // 验证文件不为空
    if config_content.trim().is_empty() {
        return Err("Config file is empty".into());
    }

    // 解析 YAML
    let config: ConnectConfig = serde_yaml::from_str(&config_content)
        .map_err(|e| format!("Failed to parse YAML config: {}", e))?;

    match serde_yaml::from_str::<ConnectConfig>(&config_content) {
        Ok(config) => println!("解析成功: {:?}", config),
        Err(e) => println!("解析错误: {}", e),
    }
    
    // 验证必要的字段
    validate_config(&config)?;

    Ok(config)
}

fn validate_config(config: &ConnectConfig) -> Result<(), Box<dyn std::error::Error>> {
    // 配置验证
    if config.options.tcp_forward_addr.trim().is_empty() {
        return Err("tcp_forward_addr cannot be empty".into());
    }
    if config.options.tcp_forward_host_prefix.trim().is_empty() {
        return Err("tcp_forward_host_prefix cannot be empty".into());
    }
    Ok(())
}


pub fn run_connect(connect_args: ConnectArgs) {
    let mut args = if let Some(config_path) = &connect_args.config {
        match load_config(config_path) {
            Ok(config) => {
                println!("Successfully loaded config from '{}'", config_path);
                println!("Config details:");
                println!("  TCP Forward Address: {}", config.options.tcp_forward_addr);
                println!("  TCP Forward Host Prefix: {}", config.options.tcp_forward_host_prefix);
                config
            },
            Err(e) => {
                eprintln!("Error loading config: {}", e);
                process::exit(1);
            }
        }
    } else {
        println!("No config file specified, using default configuration");
        ConnectConfig::default()
    };
    info!("Run connect cmd.");
    let rt = tokio::runtime::Runtime::new().unwrap();
    rt.block_on(async move {
        info!("Runtime started.");
        let connect_reader = tokio::io::stdin();
        let connect_writer = tokio::io::stdout();
        // let reader = Arc::new(Mutex::new(connect_reader));
        // let writer = Arc::new(Mutex::new(connect_writer));
        if let Err(e) = process_connect(connect_reader, connect_writer, args).await {
            eprintln!("process p2p connect: {}", e);
            process::exit(1);
        };
    });
    unsafe {}
    // TODO
}

pub fn run_client(client_args: ClientArgs) {
    let mut args = if let Some(config) = client_args.config {
        vec!["client".to_owned(), "-config".to_owned(), config]
    } else {
        vec!["client".to_owned()]
    };
    let (args, go_str) = convert_to_go_slices(&args);
    unsafe {
        #[cfg(target_os = "windows")]
        {
            _rt0_amd64_windows_lib();
        }

        RunClient(args);
    }
}

pub fn run_server(server_args: ServerArgs) {
    let args = if let Some(config) = server_args.config {
        vec!["server".to_owned(), "-config".to_owned(), config]
    } else {
        vec!["server".to_owned()]
    };
    let (args, go_str) = convert_to_go_slices(&args);
    unsafe {
        #[cfg(target_os = "windows")]
        {
            _rt0_amd64_windows_lib();
        }

        RunServer(args);
    }
}


#[cfg(target_os = "windows")]
extern "C" {
    fn _rt0_amd64_windows_lib();
}