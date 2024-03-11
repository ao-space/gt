# 构建指南（Windows）

## 1.拉取依赖项代码
### 获取webrtc源代码
[遵循链接内步骤，从官网下载webrtc源代码到.libcs/dep/_google-webrtc (不需要编译)](https://webrtc.googlesource.com/src/+/main/docs/native-code/development/)

### 获取msquic源代码
git submodule update --init libcs/dep/_msquic \
在管理员模式下执行./libcs/dep/_msquic/scripts/prepare-machine.ps1，用于初始化msquic依赖



## 2.环境准备
### webrtc编译环境(webrtc版本差异会导致所依赖的msvc工具差异):
安装visual studio2022及组件\
需要的visual studio组件:
* MSVC: v143
* Windows11 SDK(10.0.22621.0)
* 适用于v143生成工具的C++ MFC
* 适用于v143生成工具的C++ ATL
* 适用于Windows的C++ Clang工具


### msquic编译环境:
[msquic构建文档](https://github.com/microsoft/msquic/blob/main/docs/BUILD.md) \
根据构建文档安装依赖项(.Net Core,Cmake,Perl)


### 本项目编译所需环境(以下示例仅供参考，以实际路径为准):
* 将 Clang 和 MSVC 路径配置到Path环境变量\
MSVC: C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Tools\MSVC\14.39.33519\bin\Hostx64\x64 \
Clang: C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Tools\Llvm\bin \
* 配置INCLUDE和LIB环境，分别包含msvc工具的include/lib目录，和Windows SDK的include/lib目录下的所有子目录\
INCLUDE: \
C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Tools\MSVC\14.39.33519\include \
C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\ucrt \
C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\um \
C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\winrt \
C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\shared \
C:\Program Files (x86)\Windows Kits\10\Include\10.0.22621.0\cppwinrt \
LIB: \
C:\Program Files\Microsoft Visual Studio\2022\Community\VC\Tools\MSVC\14.39.33519\lib\x64 \
C:\Program Files (x86)\Windows Kits\10\Lib\10.0.22621.0\ucrt\x64 \
C:\Program Files (x86)\Windows Kits\10\Lib\10.0.22621.0\um\x64 \
C:\Program Files (x86)\Windows Kits\10\Lib\10.0.22621.0\ucrt_enclave\x64 \
* 安装powershell 7(Windows powershell存在部分命令无法正确执行的问题，需要新版的powershell 7)


## 3.编译
在powershell 7中执行根目录下的build.ps1构建脚本