$WORD_DIR = $PSScriptRoot
$WEBRTC_DIR="$WORD_DIR/libcs/dep/_google-webrtc"
$MSQUIC_DIR="$WORD_DIR/libcs/dep/_msquic"
$WEBRTC_OUT_DIR="$WEBRTC_DIR/src/out/release/obj"
$MSQUIC_OUT_DIR="$MSQUIC_DIR/build/windows/x64_schannel/obj/Release"
$MSVC_BUILD_DIR="$WORD_DIR/libcs/msvc-build"
$RUST_TARGET_DIR="$WORD_DIR/target/x86_64-pc-windows-msvc/release"

$env:CC="clang"
$env:CXX="clang++"
$env:CXXFLAGS="-I$WEBRTC_DIR/src -I$WEBRTC_DIR/src/third_party/abseil-cpp -I$MSQUIC_DIR/src/inc -std=c++17 -DWEBRTC_WIN -DQUIC_API_ENABLE_PREVIEW_FEATURES -DNOMINMAX"
$env:CGO_LDFLAGS="-L$MSQUIC_DIR/build/windows/x64_schannel/obj/Release -L$WEBRTC_DIR/src/out/release/obj -lmsquic.lib -lwebrtc.lib"
$env:CARGO_CFG_TARGET_OS="windows"

Set-Location $WORD_DIR
function complie_webrtc{
    Set-Location "$WEBRTC_DIR/src"
    gn gen out/release --args="clang_use_chrome_plugins=false is_clang=true enable_libaom=false is_component_build=false is_debug=false libyuv_disable_jpeg=true libyuv_include_tests=false rtc_build_examples=false rtc_build_tools=false rtc_enable_grpc=false rtc_enable_protobuf=false rtc_include_builtin_audio_codecs=false rtc_include_dav1d_in_internal_decoder_factory=false rtc_include_ilbc=false rtc_include_internal_audio_device=false rtc_include_tests=false rtc_use_h264=false rtc_use_x11=false treat_warnings_as_errors=false use_custom_libcxx=false use_gold=false use_lld=true use_rtti=true use_sysroot=false"
    ninja -C out/release
    Set-Location $WORD_DIR
}
if (!(Test-Path -Path "$WEBRTC_OUT_DIR/webrtc.lib")){
    complie_webrtc
}


function complie_msquic{
    Set-Location $MSQUIC_DIR
    &./scripts/prepare-machine.ps1
    &./scripts/build.ps1 -Config Release -Clean -Static -DisableTest -DisableTools -StaticCRT
    Set-Location $WORD_DIR
}
if (!(Test-Path -Path "$MSQUIC_OUT_DIR/msquic.lib")){
    if (!([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator")) {
        Write-Host "请以管理员权限运行此脚本"
        exit
    }
    complie_msquic
}


function release_gt_dylib{
    Set-Location ./libcs
    go build -tags release -trimpath -ldflags "-s -w"  -buildmode=c-archive -o release/gt.lib ./lib/export
    Set-Location ./msvc-build

    # 检查target目录是否存在
    $directory = "$WORD_DIR/libcs/msvc-build/target"
    if (-not (Test-Path -Path $directory -PathType Container)) {
        New-Item -Path $directory -ItemType Directory -Force
        Write-Host "目录已创建：$directory"
    } else {
        Write-Host "目录已存在：$directory"
    }

    cl /LD /MT /Fe:./target/gt.dll gt.cpp /link /DEF:gt.def  "../release/gt.lib" "$MSQUIC_OUT_DIR\msquic.lib" "$WEBRTC_OUT_DIR\webrtc.lib" ntdll.lib
    Set-Location $WORD_DIR
}
function release_gt_lib{
    Set-Location ./libcs
    go build -tags release -trimpath -ldflags "-s -w"  -buildmode=c-archive -o release/gt.lib ./lib/export
    Set-Location ./msvc-build

    # 检查target目录是否存在
    $directory = "$WORD_DIR/libcs/msvc-build/target"
    if (-not (Test-Path -Path $directory -PathType Container)) {
        New-Item -Path $directory -ItemType Directory -Force
        Write-Host "目录已创建：$directory"
    } else {
        Write-Host "目录已存在：$directory"
    }
    Set-Location $WORD_DIR
}
release_gt_dylib


function release_gt_exe{
    cargo build --target x86_64-pc-windows-msvc -r
}
release_gt_exe


function release_gt_with_dll{
    # 设置要打包的文件和文件夹路径
    $filesToCompress = @("$RUST_TARGET_DIR/gt.exe", "$MSVC_BUILD_DIR/target/gt.dll")

    # 设置自解压文件的输出路径和名称
    $outputFile = "$RUST_TARGET_DIR/gt-manager.exe"

    # 使用 7-Zip 创建自解压文件
    & 7z a -sfx"D:\Tools\7-Zip\7z.sfx" $outputFile $filesToCompress
}