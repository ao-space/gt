# Copyright (c) 2022 Institute of Software, Chinese Academy of Sciences (ISCAS)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DATE=$(shell date '+%F %T')
BRANCH=$(shell git branch --show-current)
COMMIT=$(shell git rev-parse HEAD | cut -c1-7)
NAME=gt
EXE=$(shell go env GOEXE)
VERSION=$(NAME) - $(DATE) - $(BRANCH) $(COMMIT)
ifdef STATIC_LINK
	export GO_STATIC_LINK_FLAG=-extldflags=-static
endif
ifdef RACE_CHECK
	export GO_RACE=-race
endif
UPDATE_SUBMODULE_COMMAND=git submodule update --init --recursive
ifdef WITH_OFFICIAL_WEBRTC
	UPDATE_SUBMODULE_COMMAND=echo 'skiped update_submodule'
endif
RELEASE_OPTIONS=$(GO_RACE) -tags release -trimpath -ldflags "$(GO_STATIC_LINK_FLAG) -s -w -X 'github.com/isrc-cas/gt/predef.Version=$(VERSION)'"
DEBUG_OPTIONS=$(GO_RACE) -trimpath -ldflags "$(GO_STATIC_LINK_FLAG) -X 'github.com/isrc-cas/gt/predef.Version=$(VERSION)'"
SOURCES=$(shell ls -1 **/*.go)
FRONTEND_DIR=web/front
SOURCES_FRONT = $(shell find $(FRONTEND_DIR) -type d \( -name 'node_modules' -o -name 'dist' \) -prune -o -type f \( -name '*.ts' -o -name '*.vue' -o -name '*.scss' -o -name '*.json' -o -name '*.cjs' -o -name '*.config.ts' -o -name '*.html' \) -print)
SERVER_FRONT_DIR=server/web
CLIENT_FRONT_DIR=client/web
TARGET?=$(shell gcc -dumpmachine)
TARGET_OS=$(shell echo $(TARGET) | awk -F '-' '{print $$2}')
ifeq ($(TARGET_OS), native)
	TARGET_OS=
endif
TARGET_CPU=$(shell echo $(TARGET) | awk -F '-' '{print $$1}')
ifeq ($(TARGET_CPU), native)
	TARGET_CPU=
endif
ifeq ($(TARGET_CPU), aarch64)
    TARGET_CPU=arm64
endif
ifeq ($(TARGET_CPU), x86_64)
    TARGET_CPU=x64
endif
ifeq ($(TARGET_CPU), i386)
    TARGET_CPU=x86
endif
export GOOS?=$(shell go env GOOS)
export GOARCH?=$(shell go env GOARCH)
export CC=$(TARGET)-gcc -w
export CXX=$(TARGET)-g++ -w
export CGO_CXXFLAGS=-I$(shell pwd)/dep/_google-webrtc/src \
	-I$(shell pwd)/dep/_google-webrtc/src/third_party/abseil-cpp \
	-I$(shell pwd)/dep/msquic/src/inc \
	-std=c++17 -DWEBRTC_POSIX -DQUIC_API_ENABLE_PREVIEW_FEATURES
export CGO_LDFLAGS= $(shell pwd)/dep/_google-webrtc/src/out/release-$(TARGET)/obj/libwebrtc.a \
 	$(shell pwd)/dep/msquic/$(TARGET)/bin/Release/libmsquic.a \
	-ldl -pthread
export CGO_ENABLED=1

.PHONY: all build docker_build_linux_arm64 fmt build_client docker_build_linux_arm64_client gofumpt build_server docker_build_linux_arm64_server golangci-lint check_webrtc_dependencies docker_release_linux_amd64 release clean docker_release_linux_amd64_client release_client compile_webrtc docker_release_linux_amd64_server release_server docker_create_image docker_build_linux_amd64 docker_release_linux_arm64 revive docker_build_linux_amd64_client docker_release_linux_arm64_client test docker_build_linux_amd64_server docker_release_linux_arm64_server update_submodule  build_web_server build_web_client release_web_server release_web_client check_npm front_release duplicate_dist_server clean_duplication_client clean_web clean_dist clean_duplication clean_duplication_server clean_duplication_client  check_msquic_dependencies compile_msquic

all: gofumpt golangci-lint test release

prepare: gofumpt golangci-lint test

gofumpt:
	gofumpt --version || go install mvdan.cc/gofumpt@latest
	gofumpt -l -w $(shell find . -name '*.go' | grep -Ev '^\./bufio|^\./client/std|^\./logger/file-rotatelogs|^\./dep')

test: compile_webrtc compile_msquic
	$(eval CGO_CXXFLAGS+=-O0 -g -ggdb)
	go test -race -cover -count 1 ./bufio ./client ./config ./server ./test ./util

golangci-lint:
	golangci-lint --version || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	golangci-lint run \
		--skip-dirs client/std \
		--skip-dirs dep \
		--skip-dirs bufio \
		--skip-dirs logger/file-rotatelogs \
		--skip-dirs build \
		--skip-dirs release \
		--exclude 'SA6002: argument should be pointer-like to avoid allocations' \
		--exclude 'S1000: should use a simple channel send/receive instead of `select` with a single case'

update_submodule:
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/_google-webrtc
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/clog
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/googletest
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl/boringssl
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/gost-engine
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl/krb5
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/gost-engine/libprov
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl/pyca-cryptography
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/krb5
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl/wycheproof
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/oqs-provider
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/pyca-cryptography
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/python-ecdsa
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/tlsfuzzer
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/tlslite-ng
	git config --global --add safe.directory /go/src/github.com/isrc-cas/gt/dep/msquic/submodules/openssl3/wycheproof
	$(UPDATE_SUBMODULE_COMMAND)

docker_create_image: update_submodule
	docker images | grep -cim1 -E "^gtbuild\s+?v1" || docker build -t gtbuild:v1 .

docker_build_linux_amd64: docker_build_linux_amd64_server docker_build_linux_amd64_client
docker_release_linux_amd64: docker_release_linux_amd64_server docker_release_linux_amd64_client
docker_build_linux_amd64_client: docker_create_image
	$(eval MAKE_ENV=TARGET=x86_64-linux-gnu GOOS=linux GOARCH=amd64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make build_client'
docker_release_linux_amd64_client: docker_create_image
	$(eval MAKE_ENV=TARGET=x86_64-linux-gnu GOOS=linux GOARCH=amd64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make release_client'
docker_build_linux_amd64_server: docker_create_image
	$(eval MAKE_ENV=TARGET=x86_64-linux-gnu GOOS=linux GOARCH=amd64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make build_server'
docker_release_linux_amd64_server: docker_create_image
	$(eval MAKE_ENV=TARGET=x86_64-linux-gnu GOOS=linux GOARCH=amd64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make release_server'

docker_build_linux_arm64: docker_build_linux_arm64_server docker_build_linux_arm64_client
docker_release_linux_arm64: docker_release_linux_arm64_server docker_release_linux_arm64_client
docker_build_linux_arm64_client: docker_create_image
	$(eval MAKE_ENV=TARGET=aarch64-linux-gnu GOOS=linux GOARCH=arm64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make build_client'
docker_release_linux_arm64_client: docker_create_image
	$(eval MAKE_ENV=TARGET=aarch64-linux-gnu GOOS=linux GOARCH=arm64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make release_client'
docker_build_linux_arm64_server: docker_create_image
	$(eval MAKE_ENV=TARGET=aarch64-linux-gnu GOOS=linux GOARCH=arm64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make build_server'
docker_release_linux_arm64_server: docker_create_image
	$(eval MAKE_ENV=TARGET=aarch64-linux-gnu GOOS=linux GOARCH=arm64 STATIC_LINK=$(STATIC_LINK) RACE_CHECK=$(RACE_CHECK) WITH_OFFICIAL_WEBRTC=$(WITH_OFFICIAL_WEBRTC))
	docker run --rm -v $(shell pwd):/go/src/github.com/isrc-cas/gt -w /go/src/github.com/isrc-cas/gt gtbuild:v1 sh -c '$(MAKE_ENV) make release_server'

build: build_server build_client
release: release_server release_client
build_client: $(SOURCES) Makefile compile_msquic compile_webrtc build_web_client
	$(eval CGO_CXXFLAGS+=-O0 -g -ggdb)
	$(eval NAME=$(GOOS)-$(GOARCH)-client)
	go build $(DEBUG_OPTIONS) -o build/$(NAME)$(EXE) ./cmd/client
release_client: $(SOURCES) Makefile compile_msquic compile_webrtc release_web_client
	$(eval CGO_CXXFLAGS+=-O3)
	$(eval NAME=$(GOOS)-$(GOARCH)-client)
	go build $(RELEASE_OPTIONS) -o release/$(NAME)$(EXE) ./cmd/client
build_server: $(SOURCES) Makefile compile_msquic compile_webrtc build_web_server
	$(eval CGO_CXXFLAGS+=-O0 -g -ggdb)
	$(eval NAME=$(GOOS)-$(GOARCH)-server)
	go build $(DEBUG_OPTIONS) -o build/$(NAME)$(EXE) ./cmd/server
release_server: $(SOURCES) Makefile compile_msquic compile_webrtc release_web_server
	$(eval CGO_CXXFLAGS+=-O3)
	$(eval NAME=$(GOOS)-$(GOARCH)-server)
	go build $(RELEASE_OPTIONS) -o release/$(NAME)$(EXE) ./cmd/server

build_web_server: $(SOURCES_FRONT) Makefile check_npm front_build duplicate_dist_server
build_web_client: $(SOURCES_FRONT) Makefile check_npm front_build duplicate_dist_client

release_web_server: $(SOURCES_FRONT) Makefile check_npm front_release duplicate_dist_server
release_web_client: $(SOURCES_FRONT) Makefile check_npm front_release duplicate_dist_client

check_npm:
	npm --version || curl -qL https://www.npmjs.com/install.sh | sh

front_build: $(SOURCES_FRONT)
	(cd $(FRONTEND_DIR) && npm install && npm run "build:test")

front_release: $(SOURCES_FRONT)
	(cd $(FRONTEND_DIR) && npm install && npm run "build:pro")

duplicate_dist_server:
	cp -r $(FRONTEND_DIR)/dist $(SERVER_FRONT_DIR)/dist

duplicate_dist_client:
	cp -r $(FRONTEND_DIR)/dist $(CLIENT_FRONT_DIR)/dist

clean: clean_web
	rm -rf build/* release/*
	rm -rf dep/_google-webrtc/src/out/*

clean_web: clean_dist
	rm -rf $(FRONTEND_DIR)/node_modules
	rm -f $(FRONTEND_DIR)/package-lock.json

clean_dist: clean_duplication
	rm -rf $(FRONTEND_DIR)/dist

clean_duplication: clean_duplication_server clean_duplication_client

clean_duplication_server:
	rm -rf $(SERVER_FRONT_DIR)/dist
clean_duplication_client:
	rm -rf $(CLIENT_FRONT_DIR)/dist


check_webrtc_dependencies:
	sh -c "command -v gn"
	sh -c "command -v ninja"
	sh -c "command -v $(CC)"
	sh -c "command -v $(CXX)"

compile_webrtc: check_webrtc_dependencies update_submodule
	cd ./dep/_google-webrtc/src && gn gen out/release-$(TARGET) --args=" \
        clang_use_chrome_plugins=false \
        enable_google_benchmarks=false \
        enable_libaom=false \
        is_clang=false \
        is_component_build=false \
        is_debug=false \
        libyuv_disable_jpeg=true \
        libyuv_include_tests=false \
        rtc_build_examples=false \
        rtc_build_tools=false \
        rtc_enable_grpc=false \
        rtc_enable_protobuf=false \
        rtc_include_builtin_audio_codecs=false \
        rtc_include_builtin_video_codecs=false \
        rtc_include_dav1d_in_internal_decoder_factory=false \
        rtc_include_ilbc=false \
        rtc_include_internal_audio_device=false \
        rtc_include_tests=false \
        rtc_use_h264=false \
        rtc_use_x11=false \
        target_cpu=\"$(TARGET_CPU)\" \
        target_os=\"$(TARGET_OS)\" \
        treat_warnings_as_errors=false \
        use_custom_libcxx=false \
        use_gold=false \
        use_lld=false \
        use_rtti=true \
        use_sysroot=false"
	sed -i 's| [^ ]*gcc | $(CC) |g' ./dep/_google-webrtc/src/out/release-$(TARGET)/toolchain.ninja
	sed -i 's| [^ ]*g++ | $(CXX) |g' ./dep/_google-webrtc/src/out/release-$(TARGET)/toolchain.ninja
	sed -i 's|"ar"|$(TARGET)-ar|g' ./dep/_google-webrtc/src/out/release-$(TARGET)/toolchain.ninja
	ninja -C ./dep/_google-webrtc/src/out/release-$(TARGET)

check_msquic_dependencies:
	sh -c "command -v cmake"

compile_msquic: check_msquic_dependencies update_submodule
	mkdir -p ./dep/msquic/$(TARGET)
	sed -i 's|\(^ *msquic_lib\)$$|\1 ALL|g' ./dep/msquic/src/bin/CMakeLists.txt
	cmake -B./dep/msquic/$(TARGET) -S./dep/msquic -DQUIC_BUILD_SHARED=OFF -DCMAKE_TARGET_ARCHITECTURE=$(TARGET_CPU)
	make -C./dep/msquic/$(TARGET) -j$(shell nproc)