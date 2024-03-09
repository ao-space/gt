# 构建指南（Windows）

## 1.环境准备
### webrtc编译环境(webrtc版本差异会导致所依赖的msvc工具差异):
(google搜索Windows编译webrtc能找到详细步骤)</br>
安装visual studio2022，</br>
MSVC版本: v143, </br>
Windows11 SDK(10.0.22621.0) </br>
启用C++ ATL和MFC支持 </br>
勾选【基于Windows的clang(llvm)工具】


### msquic编译环境:
[msquic构建文档](https://github.com/microsoft/msquic/blob/main/docs/BUILD.md) </br>
根据构建文档安装依赖项(.Net Core,Cmake,Perl)


### 本项目编译所需环境:
配置cl(msvc编译器)工具环境，使其能在powershell中调用</br>
配置clang(llvm)工具环境，使其能在powershell中调用</br>
配置INCLUDE和LIB环境，分别包含msvc工具的include/lib目录，和Windows SDK的include/lib目录下的所有子目录
安装powershell 7


## 编译
启动powershell 7，在powershell执行根目录下的build.ps1构建脚本</br>
(如果是首次构建，为了初始化msquic项目，需要在管理员模式执行powershell 7)