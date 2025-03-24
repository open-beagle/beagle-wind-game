export * from './GameCard'
export * from './GamePlatform'
export * from './GameNode'
export * from './GameInstance'

export interface GameInstanceForm {
  id?: string;
  name: string;
  gameCardId: string;
  nodeId: string;
  config: {
    maxPlayers: number;
    port: number;
    settings: {
      map: string;
      difficulty: string;
    };
  };
}

// 通用类型
export interface Pagination {
  page: number;
  pageSize: number;
  total: number;
}

export interface ListParams {
  page?: number;
  pageSize?: number;
  keyword?: string;
  [key: string]: any;
}

export interface ListResponse<T> {
  items: T[];
  total: number;
}

// 平台相关类型
export enum PlatformType {
  WINDOWS = 'windows',
  LINUX = 'linux',
  MACOS = 'macos',
  ANDROID = 'android',
  IOS = 'ios',
  WEB = 'web'
}

export interface Platform {
  id: string;
  name: string;
  description: string;
  type: PlatformType;
  logo: string;
  endpoint: string;
  settings: Record<string, any>;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface PlatformAccess {
  access_url: string;
  expires_at: string;
}

// 节点相关类型
export enum NodeType {
  PHYSICAL = 'physical',
  VIRTUAL = 'virtual',
  CONTAINER = 'container'
}

export enum NodeStatus {
  ONLINE = 'online',
  OFFLINE = 'offline',
  MAINTENANCE = 'maintenance'
}

export interface NodeResources {
  cpu: number;
  memory: number;
  storage: number;
  gpu: number;
}

export interface NodeNetwork {
  ip: string;
  port: number;
}

export interface Node {
  id: string;
  name: string;
  description: string;
  type: NodeType;
  status: NodeStatus;
  resources: NodeResources;
  network: NodeNetwork;
  metrics: Record<string, any>;
  labels: Record<string, string>;
  created_at: string;
  updated_at: string;
}

// 游戏卡片相关类型
export enum GameCardType {
  GAME = 'game',
  APP = 'app',
  TOOL = 'tool'
}

export interface GameCardFile {
  name: string;
  path: string;
  size: number;
  hash: string;
}

export interface GameCardUpdate {
  version: string;
  description: string;
  files: GameCardFile[];
  release_date: string;
}

export interface GameCard {
  id: string;
  name: string;
  sort_name: string;
  slug_name: string;
  platform_id: string;
  type: GameCardType;
  description: string;
  cover: string;
  category: string[];
  release_date: string;
  tags: string[];
  files: GameCardFile[];
  updates: GameCardUpdate[];
  patches: any[];
  params: Record<string, any>;
  settings: Record<string, any>;
  permissions: string[];
  created_at: string;
  updated_at: string;
}

// 游戏实例相关类型
export enum GameInstanceStatus {
  CREATED = 'created',
  STARTING = 'starting',
  RUNNING = 'running',
  STOPPING = 'stopping',
  STOPPED = 'stopped',
  ERROR = 'error'
}

export interface GameInstance {
  id: string;
  card_id: string;
  node_id: string;
  user_id: string;
  name: string;
  status: GameInstanceStatus;
  resources: NodeResources;
  settings: Record<string, any>;
  access_info: {
    url: string;
    credentials: Record<string, string>;
  };
  metrics: Record<string, any>;
  created_at: string;
  updated_at: string;
  last_active: string;
} 