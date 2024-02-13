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

use std::env;
use std::path::PathBuf;
use std::process::Command;

fn main() {
    let target = env::var("TARGET").unwrap();
    println!("cargo:rerun-if-changed=libcs/release/{target}");
    println!("cargo:rustc-link-search=libcs/release/{target}");
    println!("cargo:rustc-link-lib=static=cs");
    println!("cargo:rustc-link-lib=static=webrtc");
    println!("cargo:rustc-link-lib=static=msquic");
    let os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    match os.as_str() {
        "linux" => {
            let output = Command::new(format!(
                "{}-linux-gnu-gcc",
                env::var("CARGO_CFG_TARGET_ARCH").unwrap()
            ))
            .arg("--print-file-name")
            .arg("libstdc++.a")
            .output()
            .unwrap();
            let mut path = PathBuf::from(String::from_utf8_lossy(&output.stdout).into_owned());
            path.pop();
            println!("cargo:rustc-link-search=native={}", path.to_str().unwrap());
            println!("cargo:rustc-link-lib=static=stdc++");
        }
        "macos" => {
            println!("cargo:rustc-link-lib=dylib=resolv");
            println!("cargo:rustc-link-lib=dylib=c++");
            println!("cargo:rustc-link-lib=dylib=c++abi");
            println!("cargo:rustc-link-lib=framework=Security");
            println!("cargo:rustc-link-lib=framework=Cocoa");
            println!("cargo:rustc-link-lib=framework=IOKit");
            println!("cargo:rustc-link-lib=framework=CoreMedia");
            println!("cargo:rustc-link-lib=framework=AVFoundation");
        }
        os => {
            panic!("Unsupported OS: {}", os)
        }
    }
}
