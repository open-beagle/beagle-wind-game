# GameNode 模型重构

## 变更背景

GameNode 是整个系统的核心实体，负责承载游戏实例的运行。随着系统的不断迭代和功能扩展，原有的 GameNode 数据结构已经无法满足日益增长的需求：

1. 硬件信息和监控指标混合在一起，缺乏明确的职责分离
2. 资源采集模型不够清晰，导致数据重复和冗余
3. 静态硬件配置和动态监控指标没有明确界限
4. 缺少对存储设备等关键硬件的详细描述

本次变更旨在重构 GameNode 模型结构，使其能够更好地区分静态硬件配置和动态监控指标，提高系统可维护性和可扩展性。

## 变更内容

### 1. 拆分硬件信息和监控指标

将原有的 ResourceInfo 结构拆分为 HardwareInfo 和 MetricsInfo 两个独立的结构：

- HardwareInfo：专注于描述节点的静态硬件配置，只在节点注册时采集一次
- MetricsInfo：专注于描述节点的动态运行状态，定期采集更新

### 2. 细化硬件配置描述

优化了硬件配置的描述粒度，增加了更多关键参数：

#### CPU 信息

- 型号(model)：CPU 的具体型号
- 物理核心数(cores)：CPU 的物理核心数量
- 线程数(threads)：CPU 支持的并发线程数
- 基准频率(frequency)：CPU 的基准运行频率
- 缓存大小(cache)：CPU 的缓存大小

#### 内存信息

- 总容量(total)：内存总容量
- 类型(type)：内存类型，如 DDR4、DDR5 等
- 频率(frequency)：内存运行频率
- 通道数(channels)：内存通道数

#### GPU 信息

- 型号(model)：GPU 的具体型号
- 显存总量(memory_total)：GPU 的显存总量
- CUDA 核心数(cuda_cores)：GPU 的 CUDA 核心数量

#### 存储信息

将原来的单一存储设备模型改为支持多设备：

- 设备类型(type)：存储设备类型，如 SSD、HDD、NVMe 等
- 设备容量(capacity)：存储设备的总容量

### 3. 细化监控指标描述

优化了监控指标的分类和组织方式：

#### CPU 监控

- 使用率(usage)：CPU 使用率
- 温度(temperature)：CPU 温度

#### 内存监控

- 可用内存(available)：当前可用内存量
- 已用内存(used)：当前已用内存量
- 使用率(usage)：内存使用率

#### GPU 监控

- 使用率(usage)：GPU 使用率
- 显存使用情况(memory_used, memory_free, memory_usage)
- 温度(temperature)：GPU 温度
- 功耗(power)：GPU 功耗

#### 存储监控

- 已用空间(used)：已用存储空间
- 可用空间(free)：可用存储空间
- 使用率(usage)：存储使用率

#### 网络监控

- 带宽使用(bandwidth)：网络带宽使用情况
- 延迟(latency)：网络延迟
- 连接数(connections)：当前网络连接数
- 丢包率(packet_loss)：网络丢包率

### 4. 优化 GameNode 结构

对 GameNode 对象结构进行了优化调整：

```go
// GameNode 游戏节点
type GameNode struct {
    ID        string            // 节点ID
    Alias     string            // 节点别名
    Model     string            // 节点型号
    Type      GameNodeType      // 节点类型
    Location  string            // 节点位置
    Labels    map[string]string // 标签
    Hardware  map[string]string // 硬件配置(简化版)
    System    map[string]string // 系统配置
    Status    GameNodeStatus    // 节点状态信息
    CreatedAt time.Time         // 创建时间
    UpdatedAt time.Time         // 更新时间
}

// GameNodeStatus 节点状态信息
type GameNodeStatus struct {
    State      GameNodeState // 节点状态
    Online     bool          // 是否在线
    LastOnline time.Time     // 最后在线时间
    UpdatedAt  time.Time     // 状态更新时间
    Resource   ResourceInfo  // 资源信息
    Metrics    MetricsReport // 指标报告
}

// ResourceInfo 资源信息
type ResourceInfo struct {
    ID        string       // 节点ID
    Timestamp int64        // 时间戳
    Hardware  HardwareInfo // 硬件信息
    Metrics   MetricsInfo  // 监控指标
}
```

### 5. 数据迁移

为了确保现有数据能够平滑迁移到新的数据模型，编写了数据迁移脚本：

```go
// migrateResourceInfo 迁移资源信息
func migrateResourceInfo(oldResource OldResourceInfo) ResourceInfo {
	// 创建新的资源信息
	newResource := ResourceInfo{
		ID:        oldResource.ID,
		Timestamp: oldResource.Timestamp,
	}

	// 迁移硬件信息
	newResource.Hardware.CPU.Model = oldResource.Hardware.CPU.Model
	newResource.Hardware.CPU.Cores = oldResource.Hardware.CPU.Cores
	newResource.Hardware.CPU.Threads = oldResource.Hardware.CPU.Threads
	// ... 其他硬件信息

	// 迁移监控指标
	newResource.Metrics.CPU.Usage = oldResource.Hardware.CPU.Usage
	newResource.Metrics.CPU.Temperature = oldResource.Hardware.CPU.Temperature
	// ... 其他监控指标

	return newResource
}
```

### 6. 前端适配

更新前端类型定义 `frontend/src/types/GameNode.ts`，与后端数据模型保持一致：

```typescript
export interface CPUHardware {
  model: string;
  cores: number;
  threads: number;
  frequency: number;
  cache: number;
}

export interface MemoryHardware {
  total: number;
  type: string;
  frequency: number;
  channels: number;
}

export interface GPUHardware {
  model: string;
  memoryTotal: number;
  cudaCores: number;
}

export interface StorageDevice {
  type: string;
  capacity: number;
}

export interface StorageHardware {
  devices: StorageDevice[];
}

// ... 其他接口定义 ...

export interface HardwareInfo {
  cpu: CPUHardware;
  memory: MemoryHardware;
  gpu: GPUHardware;
  storage: StorageHardware;
}

export interface ResourceInfo {
  id: string;
  timestamp: number;
  hardware: HardwareInfo;
  metrics: MetricsInfo;
}
```

为了确保前端代码能够平滑过渡到新的数据模型，创建了数据转换工具类 `frontend/src/utils/dataConverter.ts`：

```typescript
import { ResourceInfo, HardwareInfo, MetricsInfo } from '../types/GameNode';

/**
 * 将旧版ResourceInfo转换为新版ResourceInfo
 * 用于在前端代码中处理可能仍然使用旧格式的API响应
 */
export function convertResourceInfo(oldResource: any): ResourceInfo {
  // 转换逻辑
}

/**
 * 将新版ResourceInfo转换为旧版ResourceInfo
 * 用于在前端代码中支持仍然使用旧格式的组件
 */
export function convertToOldResourceInfo(newResource: ResourceInfo): any {
  // 转换逻辑
}
```

这些工具函数让前端组件能够:
1. 处理从旧版API收到的数据
2. 将新格式的数据转换为旧格式，以供仍使用旧格式的组件使用
3. 在新旧版本之间平滑过渡

## 变更影响

1. **数据结构**：本次变更修改了核心数据结构，影响系统中所有与 GameNode 相关的组件

2. **数据采集**：

   - 硬件信息采集只需在节点注册时进行一次
   - 监控指标采集需要定期更新

3. **API 兼容性**：

   - 内部 API 接口需要适配新的数据结构
   - 对外 API 保持向后兼容

4. **数据存储**：
   - 存储层需要支持新的数据模型
   - 可能需要进行数据迁移

## 实施计划

1. **代码重构**：

   - 修改 models 包中的数据模型定义
   - 更新相关服务和处理器

2. **数据迁移**：

   - 开发迁移工具转换现有数据
   - 制定数据备份和恢复策略

3. **测试验证**：

   - 单元测试验证新模型的正确性
   - 集成测试验证系统兼容性
   - 性能测试评估变更影响

4. **灰度发布**：
   - 在测试环境完成验证
   - 选择非关键节点进行试点
   - 逐步推广到所有节点

## 回滚计划

1. 保留旧版本代码分支
2. 备份部署前的所有数据
3. 准备回滚脚本和数据转换工具
4. 制定详细的回滚操作手册

## 后续工作

完成本次重构后，还需要进行以下工作：

1. **单元测试完善**
   - 为新的数据模型编写单元测试
   - 确保转换函数正确处理各种边缘情况
   - 测试数据迁移脚本对历史数据的处理

2. **性能测试**
   - 验证新的数据结构是否有性能影响
   - 确保在高负载情况下系统仍能正常运行
   - 评估数据库读写性能变化

3. **文档更新**
   - 更新API文档，反映新的数据结构
   - 为开发人员提供迁移指南
   - 记录数据格式变更的原因和影响

4. **监控强化**
   - 添加新的监控指标，关注数据模型变更后的系统行为
   - 设置告警阈值，及时发现潜在问题
   - 跟踪前端组件对新数据格式的适应情况

## 总结

本次 GameNode 模型重构是系统架构优化的重要一步。通过明确区分硬件配置和监控指标，我们不仅使数据结构更加合理，还提高了系统的可维护性和可扩展性。

重构过程涉及以下几个关键步骤：

1. 分析现有数据模型的不足
2. 设计新的数据结构，区分静态和动态信息
3. 修改后端模型定义和Proto文件
4. 调整数据采集和处理逻辑
5. 更新前端类型定义和组件
6. 提供数据转换工具，确保平滑过渡

通过这些变更，我们为未来游戏节点管理系统的发展奠定了更加坚实的基础，使其能够更好地支持复杂的监控和管理需求。
