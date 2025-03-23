import { GameCard } from './GameCard'
import { GameNode } from './GameNode'

export type GameInstanceStatus = 'running' | 'stopped' | 'error'

export interface GameInstanceConfig {
  maxPlayers: number
  port: number
  settings: {
    map: string
    difficulty: string
  }
}

export interface GameInstanceMetrics {
  cpuUsage: number
  memoryUsage: number
  storageUsage: number
  gpuUsage?: number
  networkUsage: number
  uptime: number
  fps: number
  playerCount: number
}

export interface GameInstance {
  id: string
  gameId: string
  gameCardId: string
  gameCard: {
    id: string
    name: string
    platform: {
      id: string
      name: string
      type: string
      status: string
      description: string
      version: string
      image: string
      config: {
        apiKey: string
        apiUrl: string
        callbackUrl: string
      }
      environment: {
        dockerImage: string
        envVars: Record<string, string>
      }
      features: {
        autoScale: boolean
        backup: boolean
        monitoring: boolean
      }
      createdAt: string
      updatedAt: string
    }
    description: string
    coverImage: string
    status: string
    createdAt: string
  }
  nodeId: string
  node: {
    id: string
    name: string
    description: string
    type: string
    status: string
    region: string
    network: {
      ip: string
      port: number
      protocol: string
      bandwidth: number
    }
    resources: {
      cpu: number
      memory: number
      storage: number
      network: number
      gpu?: {
        model: string
        memory: number
      }
    }
    metrics: {
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
    labels: Record<string, string>
    createdAt: string
    updatedAt: string
  }
  name: string
  status: GameInstanceStatus
  config: GameInstanceConfig
  metrics: GameInstanceMetrics
  logs: string[]
  createdAt: string
  updatedAt: string
}

export interface GameInstanceQuery {
  page: number
  pageSize: number
  keyword?: string
  gameCardId?: string
  nodeId?: string
  status?: string
}

export interface GameInstanceForm {
  id?: string
  gameCardId: string
  nodeId: string
  config: {
    maxPlayers: number
    port: number
    settings: Record<string, any>
  }
} 