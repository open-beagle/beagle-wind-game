# 游戏节点硬件信息设计规范

本文档规定了游戏节点硬件信息在系统中的设计格式和规范，
适用于 Agent 采集的硬件信息以及前端展示的硬件详情。
所有相关开发人员应遵循此规范确保系统中硬件信息的一致性和可读性。

最后更新：2025-04-08

# 一、静态数据设计规范-人类可读

此部分规定了硬件信息在前端界面展示时的人类可读格式，主要用于直观展示给用户浏览。

## 1.1 CPU 信息格式规范

### 1.1.1 基本格式

```txt
<插槽序号>,<厂商> <型号> <核心数>core <主频>GHz <功耗>W
```

### 1.1.2 示例

```txt
0,Intel i9-13900K 24核心 3.6GHz 125W
```

### 1.1.3 详细规则

#### 1.1.3.1 插槽序号

- 使用从 0 开始的数字序号，如`0`,`1`,`2`,`3`
- 多 CPU 系统用分号分隔，如`0,Intel...;1,Intel...`

#### 1.1.3.2 厂商

- 使用简洁通用名称，如`Intel`、`AMD`、`ARM`
- 去除`(R)`、`Core(TM)`等商标符号
- 去除重复信息，如`Intel`后不再重复`Intel Core`

#### 1.1.3.3 型号

- 使用简洁型号名称，如`i9-13900K`、`Ryzen 9 7950X`
- 省略冗余信息，如世代与型号重复时（`13th Gen`与`13900`）仅保留型号
- 保留关键后缀（如`K`、`X`、`H`等）

#### 1.1.3.4 核心数

- 格式为`<数字>core`，如`24core`
- 使用物理核心数，非线程数

#### 1.1.3.5 主频

- 格式为`<数字>GHz`，如`3.6GHz`
- 使用基础频率，非加速频率
- 数字保留一位小数

#### 1.1.3.6 功耗

- 格式为`<数字>W`，如`125W`
- 使用 TDP 值（热设计功耗）
- 无需小数

## 1.2 GPU 信息格式规范

### 1.2.1 基本格式

```txt
<插槽序号>,<厂商> <型号> <显存>GB <功耗>W
```

### 1.2.2 示例

```txt
0,NVIDIA RTX 4090 24GB 450W
```

### 1.2.3 详细规则

#### 1.2.3.1 插槽序号

- 使用从 0 开始的数字序号，如`0`,`1`,`2`,`3`
- 多 GPU 系统用分号分隔，如`0,NVIDIA...;1,NVIDIA...`

#### 1.2.3.2 厂商

- 使用简洁通用名称，如`NVIDIA`、`AMD`、`Intel`
- 去除商标符号

#### 1.2.3.3 型号

- 使用简洁型号名称，如`RTX 4090`、`Radeon RX 7900XT`
- 对于笔记本 GPU，可标记为`RTX 4070 Laptop`，但避免过长

#### 1.2.3.4 显存

- 格式为`<数字>GB`，如`24GB`
- 使用整数 GB 值

#### 1.2.3.5 功耗

- 格式为`<数字>W`，如`450W`
- 使用 TDP 值或 GPU 功耗上限
- 无需小数

## 1.3 内存(RAM)信息格式规范

### 1.3.1 基本格式

根据不同环境采用不同格式:

#### 1.3.1.1 物理机多内存条

```txt
<总容量> GB
```

#### 1.3.1.2 虚拟机

```txt
<总容量> GB (Virtual)
```

#### 1.3.1.3 WSL 环境

```txt
<总容量> GB (WSL)
```

### 1.3.2 详细规则

#### 1.3.2.1 物理机内存

- 当有多个不同型号/频率的内存条时，只显示总容量
- 格式为`<数字> GB`，如`32 GB`
- 使用标准容量数值（4、8、16、32、64 等）
- 总是取整到最接近的标准容量
- 无需显示类型和频率信息，避免混淆

#### 1.3.2.2 虚拟机内存

- 明确标识为虚拟内存，在容量后添加"(Virtual)"标记
- 格式为`<数字> GB (Virtual)`，如`16 GB (Virtual)`
- 便于快速识别这是虚拟分配的内存，而非物理内存

#### 1.3.2.3 WSL 环境内存

- 使用特定标识区分 WSL 环境，在容量后添加"(WSL)"标记
- 格式为`<数字> GB (WSL)`，如`8 GB (WSL)`
- 反映 WSL 环境的特殊性质

### 1.3.3 示例

```txt
# 物理机多内存条
RAM: 64 GB

# 虚拟机
RAM: 16 GB (Virtual)

# WSL环境
RAM: 8 GB (WSL)
```

## 1.4 存储设备信息格式规范

### 1.4.1 基本格式

包含路径信息的格式：

```txt
<挂载路径>,<类型> <型号> <容量>
```

不包含路径信息的格式：

```txt
-,<类型> <型号> <容量>
```

### 1.4.2 示例

包含路径：

```txt
/,SSD Virtual Disk 1TB; /mnt/c,HDD Drive 4TB; /mnt/d,HDD Drive 1TB
```

### 1.4.3 详细规则

#### 1.4.3.1 挂载路径

- 显示磁盘的实际挂载点，如`/`、`/mnt/c`
- 在路径和类型之间使用逗号分隔
- 对于没有明确挂载点的存储设备，可以省略路径信息

#### 1.4.3.2 类型

- 使用标准存储类型，如`SSD`、`HDD`、`NVMe`
- 多设备用分号分隔

#### 1.4.3.3 型号

- 使用简洁型号名称
- 对于 Windows 挂载盘，使用`<盘符> Drive`格式，如`C Drive`
- 虚拟环境可使用`Virtual Disk`

#### 1.4.3.4 容量

- 1TB 以下用`<数字>GB`，如`512GB`
- 1TB 及以上用`<数字>TB`，如`2TB`
- 使用标准容量值，避免显示`0GB`等异常值

## 1.5 系统配置信息规范

### 1.5.1 基本格式（键值对）

```txt
os_distribution: <发行版>
os_version: <版本>
os_architecture: <架构>
kernel_version: <内核版本>
gpu_driver_version: <厂商> Driver <版本号>
gpu_compute_api_version: <版本号>
```

### 1.5.2 示例

```txt
os_distribution: Ubuntu
os_version: 22.04 LTS
os_architecture: x86_64
kernel_version: 5.15.0-58-generic
gpu_driver_version: NVIDIA Driver 535.129.03
gpu_compute_api_version: 12.2
```

## 1.6 规范化示例

### 1.6.1 规范化前

```txt
CPU: CPU-0,13th Gen Intel(R) Core(TM) i9-13900H 2核心 3.0GHz
GPU: 00000000:01:00.0,NVIDIA GeForce RTX 4070 Laptop GPU 8GB 250W
RAM: 15 GB
Storage: HDD Virtual Disk 0GB; HDD Virtual Disk 8GB; HDD Virtual Disk 1024GB
```

### 1.6.2 规范化后

```txt
CPU: 0,Intel i9-13900H 2core 3.0GHz 45W
GPU: 0,NVIDIA RTX 4070 Laptop 8GB 250W
RAM: 16 GB (Virtual)
Storage: /,SSD Virtual Disk 1TB; /mnt/c,HDD Drive 4TB
system:
  os_distribution: Debian
  os_version: 11
  os_architecture: x86_64
  gpu_driver_version: NVIDIA Driver 535.129.03
  gpu_compute_api_version: 12.2
```

## 1.7 实现注意事项

### 1.7.1 标准化处理

- 始终将硬件信息标准化为人类可读形式
- 避免原始数据中的异常值和非标准表示
- 内存和存储容量向上取整到最接近的标准值

### 1.7.2 自动采集限制

- 某些信息可能无法自动准确获取（如 CPU 功耗）
- 对于无法获取的值，可使用合理的估计值或留空
- 确保显示的信息保持一致性和可读性

### 1.7.3 前端显示考虑

- 前端解析时应考虑格式变化的兼容性
- 列表视图应合并显示所有硬件
- 详情视图应按插槽分组展示

# 二、动态数据设计规范-机器可读

此部分规定了硬件状态信息的机器可读格式，主要用于系统内部处理、数据传输和状态监控。

## 2.1 硬件状态属性规范

### 2.1.1 状态数据基本格式

所有硬件状态数据应遵循 GameNodeStatus 结构，符合以下 JSON 格式：

```json
{
  "state": "online|offline|maintenance|ready|busy|error",
  "online": true|false,
  "last_online": "2023-01-01T12:00:00Z",
  "updated_at": "2023-01-01T13:00:00Z",
  "hardware": {
    "cpus": [...],
    "memories": [...],
    "gpus": [...],
    "storages": [...],
    "networks": [...]
  },
  "system": {
    "os_distribution": "...",
    "os_version": "...",
    "os_architecture": "...",
    "kernel_version": "...",
    "gpu_driver_version": "...",
    "gpu_compute_api_version": "...",
    "docker_version": "...",
    "containerd_version": "...",
    "runc_version": "..."
  },
  "metrics": {
    "cpus": [...],
    "memory": {...},
    "gpus": [...],
    "storages": [...],
    "network": {...}
  }
}
```

### 2.1.2 数据获取方式和工具

#### 2.1.2.1 状态基本信息

| 字段        | 获取工具   | 数据表达         | 备注             |
| ----------- | ---------- | ---------------- | ---------------- |
| state       | agent 计算 | 枚举值           | 根据节点状态计算 |
| online      | agent 判断 | 布尔值           | 根据心跳判断     |
| last_online | agent 记录 | ISO8601 时间格式 | 最后一次心跳时间 |
| updated_at  | agent 记录 | ISO8601 时间格式 | 状态更新时间     |

#### 2.1.2.2 硬件配置信息 (Hardware)

##### CPU 设备信息

| 字段         | 获取工具 | 数据表达 | 备注                       |
| ------------ | -------- | -------- | -------------------------- |
| model        | lscpu    | 字符串   | CPU 型号名称               |
| cores        | lscpu    | 整数     | 物理核心数                 |
| threads      | lscpu    | 整数     | 线程数                     |
| frequency    | lscpu    | 浮点数   | 基准频率(GHz)              |
| cache        | lscpu    | 整数     | 缓存大小(KB)               |
| architecture | lscpu    | 字符串   | CPU 架构(x86_64, arm64 等) |

获取命令示例:

```bash
# 获取CPU信息
lscpu
```

##### 内存设备信息

| 字段      | 获取工具                 | 数据表达 | 备注              |
| --------- | ------------------------ | -------- | ----------------- |
| size      | dmidecode, /proc/meminfo | 整数     | 内存条容量(MB)    |
| type      | dmidecode                | 字符串   | 内存类型(DDR4 等) |
| frequency | dmidecode                | 浮点数   | 频率(MHz)         |

获取命令示例:

```bash
# 获取内存设备信息
dmidecode -t memory
cat /proc/meminfo
```

##### GPU 设备信息

| 字段               | 获取工具   | 数据表达 | 备注               |
| ------------------ | ---------- | -------- | ------------------ |
| model              | nvidia-smi | 字符串   | GPU 型号名称       |
| memory_total       | nvidia-smi | 整数     | 显存总量(MB)       |
| architecture       | nvidia-smi | 字符串   | GPU 架构           |
| driver_version     | nvidia-smi | 字符串   | 驱动版本           |
| compute_capability | nvidia-smi | 字符串   | 计算能力(通用属性) |
| tdp                | nvidia-smi | 整数     | 功耗指标(W)        |

获取命令示例:

```bash
# 获取NVIDIA GPU信息
nvidia-smi --query-gpu=name,memory.total,driver_version --format=csv
nvidia-smi -q

# 对于AMD GPU可使用
rocm-smi

# 对于Intel GPU可使用
intel_gpu_top
```

> `compute_capability` 字段说明:
>
> - 对于 NVIDIA GPU，表示已安装的 CUDA 版本，例如"CUDA 12.8"
> - 对于 AMD GPU，表示已安装的 ROCm 版本，例如"ROCm 5.7"
> - 对于 Intel GPU，表示已安装的 oneAPI 版本，例如"oneAPI 2024.1"
> - 如果未安装对应的计算框架，使用"Unknown"
> - 此字段反映实际可用的 GPU 计算能力，用于判断应用兼容性
> - 与硬件级的架构版本不同，此字段表示软件开发环境版本
> - 如果系统安装了多个计算框架，优先记录主要使用的框架版本

##### 存储设备信息

| 字段     | 获取工具  | 数据表达 | 备注                   |
| -------- | --------- | -------- | ---------------------- |
| type     | lsblk     | 字符串   | 存储类型(SSD/HDD/NVMe) |
| model    | lsblk     | 字符串   | 设备型号               |
| capacity | lsblk, df | 整数     | 总容量(MB)             |
| path     | df, mount | 字符串   | 挂载路径               |

获取命令示例:

```bash
# 获取存储设备信息
lsblk -o NAME,MODEL,SIZE,TYPE,MOUNTPOINT
df -h
```

##### 网络设备信息

| 字段        | 获取工具     | 数据表达 | 备注           |
| ----------- | ------------ | -------- | -------------- |
| name        | ip, ifconfig | 字符串   | 网卡名称       |
| mac_address | ip, ifconfig | 字符串   | MAC 地址       |
| ip_address  | ip, ifconfig | 字符串   | IP 地址        |
| speed       | /sys/class/net | 整数   | 网卡速率(Mbps) |

获取命令示例:

```bash
# 获取网络设备信息 
ip addr                          # 获取名称、MAC地址和IP地址
cat /sys/class/net/eth0/speed    # 获取速率
cat /proc/net/dev                # 列出所有网络接口
```

#### 2.1.2.3 系统信息 (System)

| 字段               | 获取工具                     | 数据表达 | 备注            |
| ------------------ | ---------------------------- | -------- | --------------- |
| os_distribution    | /etc/os-release, lsb_release | 字符串   | 操作系统发行版  |
| os_version         | /etc/os-release, lsb_release | 字符串   | 操作系统版本    |
| os_architecture    | uname -m                     | 字符串   | 操作系统架构    |
| kernel_version     | uname -r                     | 字符串   | 内核版本        |
| gpu_driver_version | nvidia-smi                   | 字符串   | GPU 驱动版本    |
| gpu_compute_api_version | 多种工具                | 字符串   | GPU 计算框架版本 |
| docker_version     | docker version               | 字符串   | Docker 版本     |
| containerd_version | containerd --version         | 字符串   | Containerd 版本 |
| runc_version       | runc --version               | 字符串   | Runc 版本       |

获取命令示例:

```bash
# 获取系统信息
cat /etc/os-release
uname -a
nvidia-smi | grep "Driver Version"
# GPU计算框架版本获取
nvidia-smi -q | grep "CUDA Version"  # NVIDIA GPU
rocminfo | grep "ROCm Version"       # AMD GPU
clinfo | grep "OpenCL"               # 通用
docker version --format '{{.Server.Version}}'
```

> `gpu_compute_api_version` 字段说明:
> - 不同GPU厂商对应的计算框架版本：
>   - NVIDIA GPU: CUDA版本 (例如"CUDA 12.2")
>   - AMD GPU: ROCm版本 (例如"ROCm 5.6")
>   - Intel GPU: oneAPI版本 (例如"oneAPI 2023.2")
>   - 通用GPU支持: OpenCL版本 (例如"OpenCL 3.0")
> - 此字段提供GPU计算能力的软件支持信息，对于判断应用兼容性非常重要
> - 与硬件级的`compute_capability`字段不同，此字段表示软件开发环境版本
> - 如果系统安装了多个计算框架，优先记录主要使用的框架版本

#### 2.1.2.4 监控指标信息 (Metrics)

##### CPU 监控指标

| 字段    | 获取工具      | 数据表达 | 备注           |
| ------- | ------------- | -------- | -------------- |
| model   | /proc/cpuinfo | 字符串   | CPU 型号       |
| cores   | /proc/cpuinfo | 整数     | 物理核心数     |
| threads | /proc/cpuinfo | 整数     | 线程数         |
| usage   | top, /proc/stat | 浮点数 | CPU 使用率(%)  |

获取命令示例:

```bash
# 获取CPU监控指标
mpstat -P ALL 1 1
cat /proc/stat
top -bn1 | grep "Cpu(s)"
```

##### 内存监控指标

| 字段      | 获取工具            | 数据表达 | 备注         |
| --------- | ------------------- | -------- | ------------ |
| total     | free, /proc/meminfo | 整数     | 总容量(MB)   |
| available | free, /proc/meminfo | 整数     | 可用内存(MB) |
| used      | free, /proc/meminfo | 整数     | 已用内存(MB) |
| usage     | 计算                | 浮点数   | 使用率(%)    |

获取命令示例:

```bash
# 获取内存监控指标
free -m
cat /proc/meminfo
```

##### GPU 监控指标

| 字段            | 获取工具   | 数据表达 | 备注             |
| --------------- | ---------- | -------- | ---------------- |
| model           | nvidia-smi | 字符串   | GPU 型号         |
| index           | nvidia-smi | 整数     | GPU 索引号       |
| memory_total    | nvidia-smi | 整数     | 显存总量(MB)     |
| memory_used     | nvidia-smi | 整数     | 已用显存(MB)     |
| memory_free     | nvidia-smi | 整数     | 可用显存(MB)     |
| memory_usage    | nvidia-smi | 浮点数   | 显存使用率(%)    |
| utilization_gpu | nvidia-smi | 浮点数   | GPU 使用率(%)    |

获取命令示例:

```bash
# 获取NVIDIA GPU监控指标
nvidia-smi --query-gpu=index,name,memory.total,memory.used,memory.free,utilization.gpu --format=csv,noheader
nvidia-smi dmon -s u

# 对于AMD GPU可使用
rocm-smi --showuse

# 对于Intel GPU可使用
intel_gpu_top
```

> 注意：不同厂商 GPU 的监控指标可能有所不同，采取"能获取什么就收集什么"的原则，保证核心性能指标被监控到。

##### 存储监控指标

| 字段     | 获取工具 | 数据表达 | 备注         |
| -------- | -------- | -------- | ------------ |
| path     | df       | 字符串   | 挂载路径     |
| type     | lsblk    | 字符串   | 存储类型     |
| model    | lsblk    | 字符串   | 设备型号     |
| capacity | df       | 整数     | 总容量(MB)   |
| used     | df       | 整数     | 已用空间(MB) |
| free     | df       | 整数     | 可用空间(MB) |
| usage    | df       | 浮点数   | 使用率(%)    |

获取命令示例:

```bash
# 获取存储监控指标
df -m
lsblk -o NAME,MODEL,SIZE,TYPE,MOUNTPOINT
```

##### 网络监控指标

| 字段             | 获取工具             | 数据表达 | 备注           |
| ---------------- | -------------------- | -------- | -------------- |
| inbound_traffic  | /proc/net/dev, iftop | 浮点数   | 流入流量(Mbps) |
| outbound_traffic | /proc/net/dev, iftop | 浮点数   | 流出流量(Mbps) |
| connections      | netstat, ss          | 整数     | 连接数         |

获取命令示例:

```bash
# 获取网络监控指标
cat /proc/net/dev
ss -s
netstat -an | wc -l
```

## 2.2 数据收集周期

| 数据类型     | 收集频率    | 存储周期 | 备注               |
| ------------ | ----------- | -------- | ------------------ |
| 硬件配置信息 | 启动时+每日 | 永久     | 硬件信息变动不频繁 |
| 系统信息     | 启动时+每周 | 永久     | 系统信息变动不频繁 |
| CPU 指标     | 5-30 秒     | 7-30 天  | 根据需求调整精度   |
| 内存指标     | 5-30 秒     | 7-30 天  | 根据需求调整精度   |
| GPU 指标     | 5-30 秒     | 7-30 天  | 游戏服务器核心指标 |
| 存储指标     | 30-60 秒    | 7-30 天  | 较 CPU/GPU 次要    |
| 网络指标     | 10-30 秒    | 7-30 天  | 游戏服务器核心指标 |

## 2.3 状态数据示例

```json
{
  "state": "online",
  "online": true,
  "last_online": "2023-04-02T08:11:02Z",
  "updated_at": "2023-04-02T08:11:02Z",
  "hardware": {
    "cpus": [
      {
        "model": "Intel i9-13900K",
        "cores": 24,
        "threads": 32,
        "frequency": 3.6,
        "cache": 36864,
        "architecture": "x86_64"
      }
    ],
    "memories": [
      {
        "size": 32768,
        "type": "DDR5",
        "frequency": 5600,
        "serial": "12345678",
        "slot": "DIMM1",
        "part_number": "KHX5600C40/32G",
        "form_factor": "DIMM"
      },
      {
        "size": 32768,
        "type": "DDR5",
        "frequency": 5600,
        "serial": "87654321",
        "slot": "DIMM2",
        "part_number": "KHX5600C40/32G",
        "form_factor": "DIMM"
      }
    ],
    "gpus": [
      {
        "model": "NVIDIA RTX 4090",
        "memory_total": 24576,
        "architecture": "Ada Lovelace",
        "driver_version": "535.129.03",
        "compute_capability": "8.9",
        "tdp": 450
      }
    ],
    "storages": [
      {
        "type": "NVMe",
        "model": "Samsung SSD 980 PRO",
        "capacity": 2048000,
        "path": "/"
      }
    ],
    "networks": [
      {
        "name": "eth0",
        "mac_address": "00:1A:C2:7B:00:47",
        "ip_address": "10.0.1.23",
        "speed": 2500
      }
    ]
  },
  "system": {
    "os_distribution": "Debian",
    "os_version": "11",
    "os_architecture": "x86_64",
    "kernel_version": "5.10.0-20-amd64",
    "gpu_driver_version": "535.129.03",
    "gpu_compute_api_version": "12.2",
    "docker_version": "24.0.5",
    "containerd_version": "1.6.20",
    "runc_version": "1.1.9"
  },
  "metrics": {
    "cpus": [
      {
        "model": "Intel i9-13900K",
        "cores": 24,
        "threads": 32,
        "usage": 45.3
      }
    ],
    "memory": {
      "total": 65536,
      "available": 47104,
      "used": 18432,
      "usage": 28.1
    },
    "gpus": [
      {
        "model": "NVIDIA RTX 4090",
        "memory_total": 24576,
        "usage": 78.5,
        "memory_used": 6452,
        "memory_free": 18124,
        "memory_usage": 26.3
      }
    ],
    "storages": [
      {
        "path": "/",
        "type": "NVMe",
        "model": "Samsung SSD 980 PRO",
        "capacity": 2048000,
        "used": 245760,
        "free": 1802240,
        "usage": 12.0
      }
    ],
    "network": {
      "inbound_traffic": 125.6,
      "outbound_traffic": 78.2,
      "connections": 128
    }
  }
}
```

## 2.4 兼容性考虑

1. **数据版本兼容**：

   - 在 API 响应中添加版本字段`api_version`
   - 对于不同版本的结构，提供转换函数
   - 保持向后兼容支持旧版客户端

2. **缺失数据处理**：

   - 对于无法获取的硬件信息，字段使用默认值或 null
   - CPU/GPU 信息为空数组而非 null
   - 监控指标如无法获取应设置为-1 而非 0

3. **异构环境适配**：

   - 支持不同硬件配置的服务器
   - 支持无 GPU、多 GPU 环境
   - 支持虚拟机和物理机环境

## 2.5 安全性考虑

1. **敏感信息处理**：

   - 序列号、MAC 地址等敏感信息可选择性脱敏
   - 在非管理员接口中移除敏感字段

2. **权限控制**：

   - 完整硬件信息仅管理员可见
   - 普通用户只能看到摘要信息

3. **数据传输安全**：

   - 使用 TLS/HTTPS 传输
   - 添加签名或验证机制确保数据真实性
