export type GamePlatformOS = 'Linux' | 'Windows' | 'macOS'
export type GamePlatformStatus = 'active' | 'maintenance' | 'inactive'

export interface GamePlatformFile {
  id: string
  type: string
  url: string
}

export interface GamePlatformInstallerCommand {
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
}

export interface GamePlatformConfig {
  wine?: string
  dxvk?: string
  vkd3d?: string
  python?: string
  proton?: string
  'shader-cache'?: string
  'remote-play'?: string
  broadcast?: string
  mode?: string
  resolution?: string
  wifi?: string
  bluetooth?: string
}

export interface GamePlatform {
  id: string
  name: string
  version: string
  os: GamePlatformOS
  status: GamePlatformStatus
  description: string
  image: string
  bin: string
  data: string
  files: GamePlatformFile[]
  features: string[]
  config: GamePlatformConfig
  installer?: GamePlatformInstallerCommand[]
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