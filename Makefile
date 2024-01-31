linux:
	cd ./libcs && TARGET=aarch64-linux-gnu GOOS=linux GOARCH=arm64 make release_lib
	cargo build --target aarch64-unknown-linux-gnu -r
	cd ./libcs && TARGET=x86_64-linux-gnu GOOS=linux GOARCH=amd64 make release_lib
	cargo build --target x86_64-unknown-linux-gnu -r
	cd ./libcs && TARGET=riscv64-linux-gnu GOOS=linux GOARCH=riscv64 make release_lib
	cargo build --target riscv64gc-unknown-linux-gnu -r

mac:
	cd ./libcs &&  TARGET=aarch64-apple-darwin GOOS=darwin GOARCH=arm64 arch -arch arm64 make release_lib
	cargo build --target aarch64-apple-darwin -r
	cd ./libcs && TARGET=x86_64-apple-darwin GOOS=darwin GOARCH=amd64 arch -arch x86_64 make release_lib
	cargo build --target x86_64-apple-darwin -r
