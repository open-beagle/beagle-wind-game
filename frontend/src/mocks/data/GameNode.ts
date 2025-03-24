import type { GameNode } from '@/types/GameNode'

export const mockGameNodes: GameNode[] = [
  {
    id: 'node-1',
    name: '游戏节点 1',
    type: 'physical',
    status: 'online',
    ip: '192.168.1.101',
    port: 8080,
    network: {
      ip: '192.168.1.101', 
      port: 8080,
      protocol: 'tcp',
      bandwidth: 1000
    },
    resources: {
      cpu: 16,
      memory: 65536,
      storage: 2048,
      network: 1000,
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
      networkUsage: 35.6,
      uptime: 86400 * 3 + 3600 * 5,
      fps: 60,
      instanceCount: 8,
      playerCount: 32
    },
    region: 'cn-east-1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-03-15T10:30:00Z',
    last_online: '2024-03-20T08:45:00Z'
  },
  {
    id: 'node-2',
    name: '游戏节点 2',
    type: 'virtual',
    status: 'online',
    ip: '192.168.1.102',
    port: 8080,
    network: {
      ip: '192.168.1.102',
      port: 8080,
      protocol: 'tcp',
      bandwidth: 500
    },
    resources: {
      cpu: 8,
      memory: 32768,
      storage: 1024,
      network: 500
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
      networkUsage: 28.9,
      uptime: 86400 * 1 + 3600 * 12,
      fps: 45,
      instanceCount: 4,
      playerCount: 16
    },
    region: 'cn-east-1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-03-15T10:30:00Z',
    last_online: '2024-03-20T09:15:00Z'
  },
  {
    id: 'node-3',
    name: '游戏节点 3',
    type: 'container',
    status: 'maintenance',
    ip: '192.168.1.103',
    port: 8080,
    network: {
      ip: '192.168.1.103',
      port: 8080,
      protocol: 'tcp',
      bandwidth: 250
    },
    resources: {
      cpu: 4,
      memory: 16384,
      storage: 512,
      network: 250
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
      networkUsage: 0,
      uptime: 3600 * 2,
      fps: 0,
      instanceCount: 0,
      playerCount: 0
    },
    region: 'cn-east-1',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-03-15T10:30:00Z',
    last_online: '2024-03-19T22:30:00Z'
  }
] 