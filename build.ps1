$WORD_DIR = $PSScriptRoot
$WEBRTC_DIR="$WORD_DIR/libcs/dep/_google-webrtc"
$MSQUIC_DIR="$WORD_DIR/libcs/dep/_msquic"
$WEBRTC_OUT_DIR="$WEBRTC_DIR/src/out/release/obj"
$MSQUIC_OUT_DIR="$MSQUIC_DIR/build/windows/x64_schannel/obj/Release"
$MSVC_BUILD_DIR="$WORD_DIR/libcs/msvc-build"

$env:CC="clang"
$env:CXX="clang++"
$env:CXXFLAGS="-I$WEBRTC_DIR/src -I$WEBRTC_DIR/src/third_party/abseil-cpp -I$MSQUIC_DIR/src/inc -std=c++17 -DWEBRTC_WIN -DQUIC_API_ENABLE_PREVIEW_FEATURES -DNOMINMAX"
$env:CGO_LDFLAGS="-L$MSQUIC_DIR/build/windows/x64_schannel/obj/Release -L$WEBRTC_DIR/src/out/release/obj -lmsquic.lib -lwebrtc.lib"
$env:CARGO_CFG_TARGET_OS="windows"

Write-Host "开始编译webrtc"
Set-Location "$WEBRTC_DIR/src"
$env:DEPOT_TOOLS_WIN_TOOLCHAIN="0"
$env:GYP_GENERATORS="msvs-ninja,ninja"
$env:GYP_MSVS_OVERRIDE_PATH="C:\Program Files\Microsoft Visual Studio\2022\Community"
$env:GYP_MSVS_VERSION="2022"

Set-Location $WORD_DIR
function complie_webrtc{
    Set-Location "$WEBRTC_DIR/src"
    gn gen out/release --args="clang_use_chrome_plugins=false is_clang=true enable_libaom=false is_component_build=false is_debug=false libyuv_disable_jpeg=true libyuv_include_tests=false rtc_build_examples=false rtc_build_tools=false rtc_enable_grpc=false rtc_enable_protobuf=false rtc_include_builtin_audio_codecs=false rtc_include_dav1d_in_internal_decoder_factory=false rtc_include_ilbc=false rtc_include_internal_audio_device=false rtc_include_tests=false rtc_use_h264=false rtc_use_x11=false treat_warnings_as_errors=false use_custom_libcxx=false use_gold=false use_lld=true use_rtti=true use_sysroot=false"
    ninja -C out/release
    Write-Host "webrtc编译完成"
    Set-Location $WORD_DIR
}
if (!(Test-Path -Path "$WEBRTC_OUT_DIR/webrtc.lib")){
    complie_webrtc
}


function complie_msquic{
    Write-Host "msquic开始编译"
    Set-Location $MSQUIC_DIR
    &./scripts/build.ps1 -Config Release -Clean -Static -DisableTest -DisableTools -StaticCRT
    Write-Host "msquic编译完成"
    Set-Location $WORD_DIR
}
if (!(Test-Path -Path "$MSQUIC_OUT_DIR/msquic.lib")){
    complie_msquic
}


function release_gt_dylib{
    Set-Location ./libcs
    Write-Host "开始编译gt server and client"
    go build -tags release -trimpath -ldflags "-s -w"  -buildmode=c-archive -o release/gt.lib ./lib/export

    Write-Host "gt server and client编译完成"
    Set-Location ./msvc-build

    $directory = "$WORD_DIR/libcs/msvc-build/target"
    if (-not (Test-Path -Path $directory -PathType Container)) {
        New-Item -Path $directory -ItemType Directory -Force
        Write-Host "目录已创建：$directory"
    } else {
        Write-Host "目录已存在：$directory"
    }
    Write-Host "开始编译发布gt server and client动态库"
    cl /LD /MT /Fe:$MSVC_BUILD_DIR/gt.dll gt.cpp /link /DEF:gt.def  "../release/gt.lib" "$MSQUIC_OUT_DIR/msquic.lib" "$WEBRTC_OUT_DIR/webrtc.lib" ntdll.lib
    Write-Host "动态库编译完成"
    Set-Location $WORD_DIR
}

release_gt_dylib


function release_gt_exe{
    Write-Host "开始编译gt"
    cargo build --target x86_64-pc-windows-msvc -r
    Write-Host "gt编译完成"
}
release_gt_exe