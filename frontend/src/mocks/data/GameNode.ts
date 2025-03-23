import type { GameNode } from '@/types/GameNode'

export const mockGameNodes: GameNode[] = [
  {
    id: 'node-1',
    name: '游戏节点 1',
    description: '高性能物理游戏节点，配备 RTX 4090 显卡',
    type: 'physical',
    status: 'online',
    ip: '192.168.1.101',
    port: 8080,
    resources: {
      cpu: 16,
      memory: 65536,
      storage: 2048,
      gpu: {
        model: 'NVIDIA GeForce RTX 4090',
        memory: 24576
      }
    },
    labels: {
      region: 'cn-east-1',
      zone: 'zone-a',
      rack: 'rack-01'
    },
    metrics: {
      cpuUsage: 45.5,
      memoryUsage: 68.2,
      storageUsage: 75.8,
      gpuUsage: 82.3,
      networkUsage: 35.6
    },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-03-15T10:30:00Z'
  },
  {
    id: 'node-2',
    name: '游戏节点 2',
    description: '虚拟化游戏节点，适用于轻量级游戏',
    type: 'virtual',
    status: 'online',
    ip: '192.168.1.102',
    port: 8080,
    resources: {
      cpu: 8,
      memory: 32768,
      storage: 1024
    },
    labels: {
      region: 'cn-east-1',
      zone: 'zone-b',
      rack: 'rack-02'
    },
    metrics: {
      cpuUsage: 32.1,
      memoryUsage: 45.7,
      storageUsage: 62.4,
      networkUsage: 28.9
    },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-03-15T10:30:00Z'
  },
  {
    id: 'node-3',
    name: '游戏节点 3',
    description: '容器化游戏节点，支持快速扩缩容',
    type: 'container',
    status: 'maintenance',
    ip: '192.168.1.103',
    port: 8080,
    resources: {
      cpu: 4,
      memory: 16384,
      storage: 512
    },
    labels: {
      region: 'cn-east-1',
      zone: 'zone-c',
      rack: 'rack-03'
    },
    metrics: {
      cpuUsage: 0,
      memoryUsage: 0,
      storageUsage: 45.2,
      networkUsage: 0
    },
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-03-15T10:30:00Z'
  }
] 