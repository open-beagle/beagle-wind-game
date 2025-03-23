export interface NodeHardware {
  cpu: string
  memory: string
  disk: string
}

export interface NodeNetwork {
  ip: string
}

export interface Node {
  id: string
  name: string
  status: 'online' | 'offline' | 'maintenance'
  hardware: NodeHardware
  network: NodeNetwork
  lastHeartbeat: string
}

export interface NodeQuery {
  page: number
  pageSize: number
  keyword?: string
  status?: string
}

export interface NodeForm {
  id?: string
  name: string
  hardware: NodeHardware
  network: NodeNetwork
}

export interface NodeMonitor {
  cpu: {
    usage: number
    temperature: number
    cores: number
  }
  memory: {
    total: number
    used: number
    free: number
  }
  disk: {
    total: number
    used: number
    free: number
  }
  network: {
    upload: number
    download: number
  }
}

export type GameNodeType = 'physical' | 'virtual' | 'container'
export type GameNodeStatus = 'online' | 'offline' | 'maintenance'

export interface GameNodeNetwork {
  ip: string
  port: number
  protocol: string
  bandwidth: number
}

export interface GameNodeResources {
  cpu: number
  memory: number
  storage: number
  network: number
  gpu?: {
    model: string
    memory: number
  }
}

export interface GameNodeMetrics {
  cpuUsage: number
  memoryUsage: number
  storageUsage: number
  networkUsage: number
  gpuUsage?: number
  uptime: number
  fps: number
  instanceCount: number
  playerCount: number
}

export interface GameNode {
  id: string
  name: string
  description: string
  type: GameNodeType
  status: GameNodeStatus
  region: string
  network: GameNodeNetwork
  resources: GameNodeResources
  metrics: GameNodeMetrics
  labels: Record<string, string>
  createdAt: string
  updatedAt: string
} 