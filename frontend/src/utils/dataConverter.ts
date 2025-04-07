import { 
  ResourceInfo, 
  HardwareInfo, 
  MetricsInfo
} from '../types/GameNode';

/**
 * 将旧版ResourceInfo转换为新版ResourceInfo
 * 用于在前端代码中处理可能仍然使用旧格式的API响应
 */
export function convertResourceInfo(oldResource: any): ResourceInfo {
  // 创建基本结构
  const newResource: ResourceInfo = {
    id: oldResource.id || '',
    timestamp: oldResource.timestamp || 0,
    hardware: {
      cpu: {
        model: '',
        cores: 0,
        threads: 0,
        frequency: 0,
        cache: 0
      },
      memory: {
        total: 0,
        type: '',
        frequency: 0,
        channels: 0
      },
      gpu: {
        model: '',
        memoryTotal: 0,
        cudaCores: 0
      },
      storage: {
        devices: []
      }
    },
    metrics: {
      cpu: {
        usage: 0,
        temperature: 0
      },
      memory: {
        available: 0,
        used: 0,
        usage: 0
      },
      gpu: {
        usage: 0,
        memoryUsed: 0,
        memoryFree: 0,
        memoryUsage: 0,
        temperature: 0,
        power: 0
      },
      storage: {
        used: 0,
        free: 0,
        usage: 0
      },
      network: {
        bandwidth: 0,
        latency: 0,
        connections: 0,
        packetLoss: 0
      }
    }
  };

  // 如果输入就是新格式，直接返回
  if (oldResource.metrics) {
    return oldResource as ResourceInfo;
  }

  // 处理旧版数据
  if (oldResource.hardware) {
    const oldHardware = oldResource.hardware;
    
    // 转换CPU信息
    if (oldHardware.cpu) {
      newResource.hardware.cpu.model = oldHardware.cpu.model || '';
      newResource.hardware.cpu.cores = oldHardware.cpu.cores || 0;
      newResource.hardware.cpu.threads = oldHardware.cpu.threads || 0;
      newResource.hardware.cpu.frequency = oldHardware.cpu.frequency || 0;
      newResource.hardware.cpu.cache = oldHardware.cpu.cache || 0;
      
      // 将动态指标转移到metrics中
      newResource.metrics.cpu.usage = oldHardware.cpu.usage || 0;
      newResource.metrics.cpu.temperature = oldHardware.cpu.temperature || 0;
    }
    
    // 转换内存信息
    if (oldHardware.memory) {
      newResource.hardware.memory.total = oldHardware.memory.total || 0;
      newResource.hardware.memory.type = oldHardware.memory.type || '';
      newResource.hardware.memory.frequency = oldHardware.memory.frequency || 0;
      newResource.hardware.memory.channels = oldHardware.memory.channels || 0;
      
      // 将动态指标转移到metrics中
      newResource.metrics.memory.available = oldHardware.memory.available || 0;
      newResource.metrics.memory.used = oldHardware.memory.used || 0;
      newResource.metrics.memory.usage = oldHardware.memory.usage || 0;
    }
    
    // 转换GPU信息
    if (oldHardware.gpu) {
      newResource.hardware.gpu.model = oldHardware.gpu.model || '';
      newResource.hardware.gpu.memoryTotal = oldHardware.gpu.memoryTotal || 0;
      newResource.hardware.gpu.cudaCores = oldHardware.gpu.cudaCores || 0;
      
      // 将动态指标转移到metrics中
      newResource.metrics.gpu.usage = oldHardware.gpu.usage || 0;
      newResource.metrics.gpu.memoryUsed = oldHardware.gpu.memoryUsed || 0;
      newResource.metrics.gpu.memoryFree = oldHardware.gpu.memoryFree || 0;
      newResource.metrics.gpu.memoryUsage = oldHardware.gpu.memoryUsage || 0;
      newResource.metrics.gpu.temperature = oldHardware.gpu.temperature || 0;
      newResource.metrics.gpu.power = oldHardware.gpu.power || 0;
    }
    
    // 转换存储信息
    if (oldHardware.disk) {
      // 添加一个存储设备
      newResource.hardware.storage.devices.push({
        type: oldHardware.disk.type || '',
        capacity: oldHardware.disk.capacity || 0
      });
      
      // 将动态指标转移到metrics中
      newResource.metrics.storage.used = oldHardware.disk.used || 0;
      newResource.metrics.storage.free = oldHardware.disk.free || 0;
      newResource.metrics.storage.usage = oldHardware.disk.usage || 0;
    }
  }
  
  // 处理网络信息
  if (oldResource.network) {
    newResource.metrics.network.bandwidth = oldResource.network.bandwidth || 0;
    newResource.metrics.network.latency = oldResource.network.latency || 0;
    newResource.metrics.network.connections = oldResource.network.connections || 0;
    newResource.metrics.network.packetLoss = oldResource.network.packetLoss || 0;
  }
  
  return newResource;
}

/**
 * 将新版ResourceInfo转换为旧版ResourceInfo
 * 用于在前端代码中支持仍然使用旧格式的组件
 */
export function convertToOldResourceInfo(newResource: ResourceInfo): any {
  const oldResource: any = {
    id: newResource.id,
    timestamp: newResource.timestamp,
    hardware: {
      cpu: {
        model: newResource.hardware.cpu.model,
        cores: newResource.hardware.cpu.cores,
        threads: newResource.hardware.cpu.threads,
        frequency: newResource.hardware.cpu.frequency,
        temperature: newResource.metrics.cpu.temperature,
        usage: newResource.metrics.cpu.usage,
        cache: newResource.hardware.cpu.cache
      },
      memory: {
        total: newResource.hardware.memory.total,
        available: newResource.metrics.memory.available,
        used: newResource.metrics.memory.used,
        usage: newResource.metrics.memory.usage,
        type: newResource.hardware.memory.type,
        frequency: newResource.hardware.memory.frequency,
        channels: newResource.hardware.memory.channels
      },
      gpu: {
        model: newResource.hardware.gpu.model,
        memoryTotal: newResource.hardware.gpu.memoryTotal,
        memoryUsed: newResource.metrics.gpu.memoryUsed,
        memoryFree: newResource.metrics.gpu.memoryFree,
        memoryUsage: newResource.metrics.gpu.memoryUsage,
        usage: newResource.metrics.gpu.usage,
        temperature: newResource.metrics.gpu.temperature,
        power: newResource.metrics.gpu.power,
        cudaCores: newResource.hardware.gpu.cudaCores
      }
    },
    network: {
      bandwidth: newResource.metrics.network.bandwidth,
      latency: newResource.metrics.network.latency,
      connections: newResource.metrics.network.connections,
      packetLoss: newResource.metrics.network.packetLoss
    }
  };
  
  // 处理存储设备
  if (newResource.hardware.storage && newResource.hardware.storage.devices.length > 0) {
    const device = newResource.hardware.storage.devices[0];
    oldResource.hardware.disk = {
      model: device.type, // 使用类型作为模型名称
      capacity: device.capacity,
      used: newResource.metrics.storage.used,
      free: newResource.metrics.storage.free,
      usage: newResource.metrics.storage.usage,
      type: device.type
    };
  } else {
    oldResource.hardware.disk = {
      model: 'Unknown',
      capacity: 0,
      used: newResource.metrics.storage.used,
      free: newResource.metrics.storage.free,
      usage: newResource.metrics.storage.usage,
      type: 'Unknown'
    };
  }
  
  return oldResource;
} 