# forza-painter geometrize GPU Version

**Forza is a trademark of Microsoft. This project is unofficial and not affiliated with or endorsed by Microsoft.**

这是一个[forza-painter](https://github.com/forza-painter/forza-painter)的第三方geometrize JSON生成工具，旨在通过优化JSON生成效率来提高单位时间生成的涂装质量

## 本项目和原始项目的差异

- 通过[OpenCL-SDK](https://github.com/KhronosGroup/OpenCL-SDK)/Vulkan调用GPU进行误差计算和图元栅格化，以大幅度提升生成效率
- 仅保留Rotated Ellipse渲染，没有其他几何的渲染（forza-painter用不到）
- 支持PNG格式原图输入alpha通道，保护原有透明像素不会被几何覆盖
- 支持批量候选并行评估，进一步提升生成效率
- 每次添加几何之后不再重算整图的误差，而是只在椭圆包围盒且椭圆内部像素计算选取最小delta
- 整个项目使用Go重构，CGO编译后即可在不同平台直接发布二进制文件，无需runtime

## 编译安装

### 环境需求

```
Go w/ CGO >= v1.24
OpenCL-SDK >= v3.0.19
Vulkan SDK >= 1.4.350.0
```

其中 OpenCL-SDK 仅用于构建时链接，Vulkan SDK 仅用于编译 Vulkan 后端和打包 `vulkan-1.dll`，运行时发布包不需要用户再单独安装 Vulkan SDK。

### 编译Windows版本

克隆本项目，下载[OpenCL-SDK](https://github.com/KhronosGroup/OpenCL-SDK/releases/tag/v2025.07.23)的Windows版本并放到仓库根目录的`OpenCL-SDK`目录下。
同时安装[Vulkan SDK](https://vulkan.lunarg.com/sdk/home)，默认使用`C:\VulkanSDK\1.4.350.0`，也可以通过`VULKAN_SDK`环境变量指定其他版本。

执行 `build.ps1` 后，会自动编译 `.comp` 着色器并生成单文件程序，然后把发布包输出到 `dist`：

- `dist\forza-painter-geometrize-go.exe`
- `dist\vulkan-1.dll`
- `dist\shaders\*.spv`

一次性构建单个二进制并生成发布包：

```PowerShell
powershell -ExecutionPolicy Bypass -File "build.ps1"
```

如果你想指定输出文件名：

```PowerShell
powershell -ExecutionPolicy Bypass -File "build.ps1" -OutputName "forza-painter-geometrize-go.exe"
```

这个二进制同时包含OpenCL和Vulkan后端，运行时通过`-backend opencl`或`-backend vulkan`切换。发布包已经带上 `vulkan-1.dll` 和 SPIR-V 着色器文件，目标机器通常不需要单独安装 Vulkan SDK。

对于Linux/MacOS，请参考`build.ps1`里的`CGO_CFLAGS`和`CGO_LDFLAGS`配置方式自行调整。

## 开始使用

### 命令行参数

```
Usage of forza-painter-geometrize-go.exe:
  -backend string
        GPU backend: opencl (default) or vulkan (default "opencl")
  -output string
        Output path prefix (default: input image path)
  -preview string
        Optional preview PNG output path
  -profile string
        Profile name fragment under ./settings
  -resume string
        Resume from a saved geometry checkpoint JSON
  -seed int
        Optional RNG seed for reproducible output
  -settings string
        Path to settings ini file
```

### 示例

假设你需要导入Forza的图片为C:\work\forza\test.png，期望生成预览图片的路径为C:\work\forza\preview，配置文件在C:\work\forza\settings\c.ini

则最终执行的命令为

```cmd
forza-painter-geometrize-go.exe C:\work\forza\test.png -preview "C:\work\forza\preview" -settings "C:\work\forza\settings\c.ini"
```

执行后即可实时输出生成进度，并且在你指定的saveAt的几何数量时输出JSON到图片所在路径（可通过-output设置输出JSON路径）

生成JSON完成且预览图片效果满意后，请使用你原有的forza-painter分支来将它导入到游戏中

FH4/FH5：https://github.com/forza-painter/forza-painter/

FH6：https://github.com/bvzrays/forza-painter-fh6/

## 性能测试

我使用了我的工作用平板电脑，即使它只有iGPU，速度也远远高于CPU（使用forza-painter-fh6）

同时，在[DavidHuang](https://github.com/hjc4869)帮助我在他的Ryzen AIMAX395设备上测试时，速度得到指数级提升，证明Xe iGPU的性能远未达到程序瓶颈

```
ayylmao.png - c.ini - i5-12500H+iGPU
cpu: ~966ms per ellipse
opencl: ~435ms per ellipse

maozai.jpg - c.ini - i5-12500H+iGPU
cpu: ~11175ms per ellipse
opencl: ~1364ms per ellipse

ayylmao.png - c.ini - aimax395+8060S
cpu: unknown (linux device)
opencl: ~37ms per ellipse

maozai.jpg - c.ini - aimax395+8060S
cpu: unknown (linux device)
opencl: ~340ms per ellipse
```

## 鸣谢

原始项目：https://github.com/forza-painter/forza-painter

geometrize项目提供的生成思路：https://github.com/Tw1ddle/geometrize

@hjc4869帮助我测试STXHalo平台的性能：https://github.com/hjc4869

## 生成式AI披露

在开发过程中我使用了生成式AI辅助安排计划和编写代码，本文档的非中文版本也由生成式AI自动翻译，以下是本项目开发过程中用到的生成式AI信息：

OpenAI GPT-5.5

OpenAI GPT-5.3 Codex

Moonshot K2.6
