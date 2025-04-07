# GameNodeAgent 逻辑重构

## 重构目标

将 `internal/gamenode/gamenode_agent.go` 文件中的硬件信息、系统信息和指标信息的获取逻辑拆分成独立的模块，以提高代码的可维护性和可测试性。

## 重构计划

1. 创建新的模块化文件结构：

   - `internal/sysinfo/hardware.go`：负责采集硬件信息
   - `internal/sysinfo/system.go`：负责采集系统信息
   - `internal/sysinfo/metrics.go`：负责采集监控指标

2. 重构 `GameNodeAgent` 类，移除与系统信息采集相关的直接逻辑，改为调用新模块

3. 调整相关接口和函数签名，确保兼容性

4. 集成第三方库以提高代码质量和功能稳定性

## 第三方库选择

### 硬件信息收集 (hardware.go)

推荐使用以下第三方库：

1. **gopsutil** (https://github.com/shirou/gopsutil)

   - 跨平台系统信息收集库
   - 提供 CPU、内存、磁盘、网络等硬件信息的收集功能
   - API 友好，跨平台支持良好，社区活跃

2. **ghw** (https://github.com/jaypipes/ghw)

   - 专注于硬件信息收集的库
   - 提供详细的硬件规格数据
   - 可作为补充选项

对于 GPU 信息收集（多厂商支持）：

- **NVIDIA GPU**:

  - **nvidia-smi 命令行调用**：通过 exec 包执行 nvidia-smi 命令并解析结果
  - **go-nvml** (https://github.com/NVIDIA/go-nvml)：NVIDIA Management Library 的 Go 封装

- **AMD GPU**:

  - **rocm-smi 命令行调用**：通过 exec 包执行 rocm-smi 命令获取 AMD GPU 信息
  - 考虑使用 **RadeonTop** 的输出进行解析

- **Intel GPU**:

  - **intel_gpu_top 命令行调用**：获取 Intel GPU 使用情况
  - 解析 **/sys/class/drm** 目录下的信息获取 Intel GPU 参数

- **通用方案**:

  - 使用 **lspci** 命令先识别 GPU 厂商和型号
  - 根据识别结果调用对应的专用工具
  - 使用 **hwinfo** 或 **lshw** 获取跨厂商硬件信息

- **预备方案**:

  - 针对不同 GPU 厂商设计专门的检测和数据收集模块
  - 实现厂商无关的接口以统一数据处理流程
  - 支持自动检测可用的 GPU 工具并选择最佳收集方式

### 指标收集 (metrics.go)

可以考虑：

1. 继续使用**gopsutil**作为基础，它也提供了运行时指标收集功能
2. **prometheus/client_golang** (https://github.com/prometheus/client_golang)
   - 与 Prometheus 生态系统兼容
   - 标准化的指标格式
3. **go-metrics** (https://github.com/rcrowley/go-metrics)
   - 轻量级指标收集库
   - 简单易用，低开销

### 系统信息收集 (system.go)

主要使用：

1. **gopsutil**的 OS 模块获取操作系统信息
2. 结合 exec 调用系统命令获取特定信息（如 CUDA 版本、GPU 驱动版本等）

## 重构步骤

### 1. 创建系统信息采集模块

将在 `internal/sysinfo` 目录下创建以下模块：

#### 1.1 hardware.go

提供硬件信息采集功能，包括：

- CPU 信息采集
- 内存信息采集
- GPU 信息采集（支持 NVIDIA、AMD 和 Intel）
- 存储设备信息采集
- 网络设备信息采集

#### 1.2 system.go

提供系统信息采集功能，包括：

- 操作系统信息采集
- 内核版本信息采集
- GPU 驱动信息采集（支持多厂商）
- CUDA/ROCm/oneAPI 版本信息采集
- 容器运行时信息采集

#### 1.3 metrics.go

提供监控指标采集功能，包括：

- CPU 使用率和温度采集
- 内存使用情况采集
- GPU 使用情况和温度采集（支持多厂商）
- 存储设备使用情况采集
- 网络使用情况采集

### 2. 调整 GameNodeAgent 结构

修改 `GameNodeAgent` 类，主要变更包括：

- 移除直接的系统信息采集逻辑
- 添加对新模块的依赖和调用
- 调整相关方法的实现

### 3. 兼容性处理

为确保现有代码的兼容性，将：

- 保持现有的函数签名不变
- 确保返回的数据结构与原来一致
- 适当添加适配层处理格式转换，适配层的代码建议和 agent 放到一起

## 预期成果

1. 代码结构更加清晰，职责划分更加明确
2. 提高代码的可维护性
3. 为未来的功能扩展和性能优化奠定基础
4. 通过使用成熟的第三方库提高代码质量和稳定性
5. 支持多种 GPU 厂商（NVIDIA、AMD、Intel），提高系统兼容性

## 重构进度

- [ ] 创建 `internal/sysinfo/hardware.go`
- [ ] 创建 `internal/sysinfo/system.go`
- [ ] 创建 `internal/sysinfo/metrics.go`
- [ ] 重构 `internal/gamenode/gamenode_agent.go`
- [ ] 添加对第三方库的依赖
