# 构建指南（Windows）

## 1.拉取子模块代码
git submodule update --init --recursive

## 2.环境准备
### webrtc编译环境(webrtc版本差异会导致所依赖的msvc工具差异):
安装visual studio2022及组件\
需要的visual studio组件:
* MSVC: v143
* Windows11 SDK(10.0.22621.0)
* 适用于v143生成工具的C++ MFC
* 适用于v143生成工具的C++ ATL
* 适用于Windows的C++ Clang工具

通过vs_installer在命令行安装: \
`$ PATH_TO_INSTALLER.EXE
--add Microsoft.VisualStudio.Workload.NativeDesktop
--add Microsoft.VisualStudio.Component.VC.ATLMFC
--add Microsoft.VisualStudio.Component.Windows11SDK.22621
--add Microsoft.VisualStudio.Component.VC.Tools.x86.x64
--add Microsoft.VisualStudio.ComponentGroup.NativeDesktop.Llvm.Clang
--includeRecommended`
### msquic编译环境:
[msquic构建文档](https://github.com/microsoft/msquic/blob/main/docs/BUILD.md) \
根据构建文档安装依赖项(.Net Core,Cmake) \
也可通过winget包管理器安装依赖项: \
winget install Microsoft.DotNet.SDK.8 \
winget install Microsoft.DotNet.DesktopRuntime.8 \
winget install --id=Kitware.CMake  -e


## 3.编译
在powershell 7中执行根目录下的build.ps1构建脚本