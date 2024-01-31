use std::env;

fn main() {
    let target = env::var("TARGET").unwrap();
    println!("cargo:rustc-link-search=cs/release/{target}");
    println!("cargo:rustc-link-lib=static=cs");
    println!("cargo:rustc-link-lib=static=webrtc");
    println!("cargo:rustc-link-lib=static=msquic");
    let os = env::var("CARGO_CFG_TARGET_OS").unwrap();
    match os.as_str() {
        "linux" => {
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
