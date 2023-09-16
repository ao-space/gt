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
export CGO_CXXFLAGS=-I$(shell pwd)/dep/_google-webrtc/src -I$(shell pwd)/dep/_google-webrtc/src/third_party/abseil-cpp -std=c++17 -DWEBRTC_POSIX
export CGO_LDFLAGS=$(shell pwd)/dep/_google-webrtc/src/out/release-$(TARGET)/obj/libwebrtc.a -ldl -pthread
export CGO_ENABLED=1

.PHONY: all build docker_build_linux_arm64 fmt build_client docker_build_linux_arm64_client gofumpt build_server docker_build_linux_arm64_server golangci-lint check_webrtc_dependencies docker_release_linux_amd64 release clean docker_release_linux_amd64_client release_client compile_webrtc docker_release_linux_amd64_server release_server docker_create_image docker_build_linux_amd64 docker_release_linux_arm64 revive docker_build_linux_amd64_client docker_release_linux_arm64_client test docker_build_linux_amd64_server docker_release_linux_arm64_server update_submodule

all: gofumpt golangci-lint test release

prepare: gofumpt golangci-lint test

gofumpt:
	gofumpt --version || go install mvdan.cc/gofumpt@latest
	gofumpt -l -w $(shell find . -name '*.go' | grep -Ev '^\./bufio|^\./client/std|^\./logger/file-rotatelogs|^\./dep')

test: compile_webrtc
	$(eval CGO_CXXFLAGS+=-O0 -g -ggdb)
	go test -race -cover -count 1 ./...

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
build_client: $(SOURCES) Makefile compile_webrtc
	$(eval CGO_CXXFLAGS+=-O0 -g -ggdb)
	$(eval NAME=$(GOOS)-$(GOARCH)-client)
	go build $(DEBUG_OPTIONS) -o build/$(NAME)$(EXE) ./cmd/client
release_client: $(SOURCES) Makefile compile_webrtc
	$(eval CGO_CXXFLAGS+=-O3)
	$(eval NAME=$(GOOS)-$(GOARCH)-client)
	go build $(RELEASE_OPTIONS) -o release/$(NAME)$(EXE) ./cmd/client
build_server: $(SOURCES) Makefile compile_webrtc
	$(eval CGO_CXXFLAGS+=-O0 -g -ggdb)
	$(eval NAME=$(GOOS)-$(GOARCH)-server)
	go build $(DEBUG_OPTIONS) -o build/$(NAME)$(EXE) ./cmd/server
release_server: $(SOURCES) Makefile compile_webrtc
	$(eval CGO_CXXFLAGS+=-O3)
	$(eval NAME=$(GOOS)-$(GOARCH)-server)
	go build $(RELEASE_OPTIONS) -o release/$(NAME)$(EXE) ./cmd/server

clean:
	rm build/* release/*
	rm -r dep/_google-webrtc/src/out/*

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
