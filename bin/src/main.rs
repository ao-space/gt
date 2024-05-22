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

use std::path::PathBuf;

use clap::Parser;
use clap::Subcommand;
use env_logger::Env;
use log::{error, info};

use gt::*;
use gt::manager::Signal;

use crate::cs::{ClientArgs, ServerArgs};
use crate::manager::ManagerArgs;

#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,

    /// Path to the config file or the directory containing the config files
    #[arg(short, long)]
    config: Option<PathBuf>,
    /// The maximum allowed depth of the subdirectory to be traversed to search config files
    #[arg(long)]
    depth: Option<u8>,
    /// Send signal to the running GT processes
    #[arg(short, long, value_enum)]
    signal: Option<Signal>,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// Run GT Server
    Server(ServerArgs),
    /// Run GT Client
    Client(ClientArgs),

    #[command(hide = true)]
    SubP2P,
    #[command(hide = true)]
    SubServer(ServerArgs),
    #[command(hide = true)]
    SubClient(ClientArgs),
}

fn main() {
    env_logger::Builder::from_env(Env::default().default_filter_or("info")).init();
    let cli = Cli::parse();
    if let Some(signal) = cli.signal {
        if let Err(e) = manager::send_signal(signal) {
            error!("failed to send {signal:?} signal: {:?}", e);
        } else {
            info!("{signal:?} signal sent");
        }
        return;
    }
    let mut manager_args = ManagerArgs {
        config: cli.config,
        depth: cli.depth,
        server_args: None,
        client_args: None,
    };
    if let Some(command) = cli.command {
        match command {
            Commands::Server(args) => {
                manager_args.server_args = Some(args);
            }
            Commands::Client(args) => {
                manager_args.client_args = Some(args);
            }
            Commands::SubP2P => {
                info!("GT SubP2P");
                peer::start_peer_connection();
                info!("GT SubP2P done");
                return;
            }
            Commands::SubServer(args) => {
                info!("GT SubServer");
                cs::run_server(args);
                info!("GT SubServer done");
                return;
            }
            Commands::SubClient(args) => {
                info!("GT SubClient");
                cs::run_client(args);
                info!("GT SubClient done");
                return;
            }
        }
    }

    let m = manager::Manager::new(manager_args);
    info!("GT");
    if let Err(e) = m.run_manager() {
        error!("GT Manager error: {:?}", e);
    }
    info!("GT done");
}
