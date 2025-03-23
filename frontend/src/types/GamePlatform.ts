export interface PlatformFile {
  id: string
  type: string
  url: string
}

export type GamePlatformOS = 'Linux' | 'Windows' | 'macOS'
export type GamePlatformStatus = 'active' | 'maintenance' | 'inactive'

export interface GamePlatformConfig {
  apiKey: string
  apiUrl: string
  callbackUrl: string
}

export interface GamePlatformEnvironment {
  dockerImage: string
  dockerTag: string
  env: Record<string, string>
  volumes: Record<string, string>
}

export interface GamePlatformFeatures {
  gameTypes: string[]
  platforms: string[]
}

export interface GamePlatform {
  id: string
  name: string
  os: GamePlatformOS
  status: GamePlatformStatus
  description: string
  version: string
  image: string
  resources: {
    cpu: number
    memory: number
    storage: number
  }
  config: GamePlatformConfig
  environment: GamePlatformEnvironment
  features: GamePlatformFeatures
  createdAt: string
  updatedAt: string
}

export interface GamePlatformQuery {
  page: number
  pageSize: number
  keyword?: string
  status?: string
}

export interface GamePlatformForm {
  id?: string
  name: string
  type: string
  description: string
  status: 'active' | 'inactive' | 'maintenance'
  version: string
  image: string
  bin: string
  data: string
  files: PlatformFile[]
  features: string[]
  config: Record<string, string>
  game?: {
    exe: string
  }
  installer?: {
    command?: string
    move?: {
      src: string
      dst: string
    }
    chmodx?: string
    extract?: {
      file: string
      dst: string
    }
  }[]
} 