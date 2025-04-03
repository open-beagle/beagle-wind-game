export interface NodeHardware {
  cpu: string;
  memory: string;
  disk: string;
}

export interface NodeNetwork {
  ip: string;
}

export interface Node {
  id: string;
  name: string;
  status: "online" | "offline" | "maintenance";
  hardware: NodeHardware;
  network: NodeNetwork;
  lastHeartbeat: string;
}

export interface NodeMonitor {
  cpu: {
    usage: number;
    temperature: number;
    cores: number;
  };
  memory: {
    total: number;
    used: number;
    free: number;
  };
  disk: {
    total: number;
    used: number;
    free: number;
  };
  network: {
    upload: number;
    download: number;
  };
}

export type GameNodeType = "physical" | "virtual" | "container";
export type GameNodeStatus = "online" | "offline" | "maintenance" | "ready";

export interface GameNodeSystem {
  os_type?: string;
  os_version?: number;
  gpu_driver?: string;
  cuda_version?: number;
}

export interface GameNodeHardware {
  CPU?: string;
  RAM?: string;
  GPU?: string;
  Storage?: string;
  Network?: string;
}

export interface GameNodeResources {
  cpu?: number;
  memory?: number;
  storage?: number;
  network?: number;
  CPU_Usage?: string;
  RAM_Usage?: string;
  GPU_Usage?: string;
  Storage?: string;
  gpu?: {
    model: string;
    memory: number;
  };
}

export interface GameNodeMetrics {
  cpuUsage?: number;
  memoryUsage?: number;
  storageUsage?: number;
  networkUsage?: number;
  gpuUsage?: number;
  uptime?: number;
  fps?: number;
  instanceCount?: number;
  playerCount?: number;
}

export interface GameNode {
  id: string;
  alias: string;
  model: string;
  type: GameNodeType;
  location?: string;
  labels?: Record<string, string>;
  hardware?: GameNodeHardware;
  system?: GameNodeSystem;
  status?: GameNodeStatus;
  created_at?: string;
  updated_at?: string;
}

// 前端类型: 查询参数类型
export interface GameNodeQuery {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: string;
}

// 前端类型: 表单类型使用 Partial<GameNode>，使所有字段变为可选
export type GameNodeForm = Partial<GameNode>;
