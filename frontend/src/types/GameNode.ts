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

export enum GameNodeType {
  Physical = 'physical',
  Virtual = 'virtual',
  Container = 'container'
}

export enum GameNodeState {
  Offline = 'offline',
  Online = 'online',
  Maintenance = 'maintenance',
  Ready = 'ready',
  Busy = 'busy',
  Error = 'error'
}

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

export interface CPUMetrics {
  usage: number;
  temperature: number;
}

export interface MemoryMetrics {
  available: number;
  used: number;
  usage: number;
}

export interface GPUMetrics {
  usage: number;
  memoryUsed: number;
  memoryFree: number;
  memoryUsage: number;
  temperature: number;
  power: number;
}

export interface StorageMetrics {
  used: number;
  free: number;
  usage: number;
}

export interface NetworkMetrics {
  bandwidth: number;
  latency: number;
  connections: number;
  packetLoss: number;
}

export interface HardwareInfo {
  cpu: CPUHardware;
  memory: MemoryHardware;
  gpu: GPUHardware;
  storage: StorageHardware;
}

export interface MetricsInfo {
  cpu: CPUMetrics;
  memory: MemoryMetrics;
  gpu: GPUMetrics;
  storage: StorageMetrics;
  network: NetworkMetrics;
}

export interface ResourceInfo {
  id: string;
  timestamp: number;
  hardware: HardwareInfo;
  metrics: MetricsInfo;
}

export interface Metric {
  name: string;
  type: string;
  value: number;
  labels: Record<string, string>;
}

export interface MetricsReport {
  id: string;
  timestamp: number;
  metrics: Metric[];
}

export interface GameNodeStatus {
  state: GameNodeState;
  online: boolean;
  lastOnline: string;
  updatedAt: string;
  resource: ResourceInfo;
  metrics: MetricsReport;
}

export interface GameNode {
  id: string;
  alias: string;
  model: string;
  type: GameNodeType;
  location: string;
  labels: Record<string, string>;
  hardware: Record<string, string>;
  system: Record<string, string>;
  status: GameNodeStatus;
  createdAt: string;
  updatedAt: string;
}

export interface GameNodeQuery {
  page: number;
  pageSize: number;
  keyword?: string;
  status?: string;
}

export type GameNodeForm = Partial<GameNode>;

export interface GameNodeListParams {
  page?: number;
  size?: number;
  keyword?: string;
  status?: GameNodeState;
  type?: GameNodeType;
  sortBy?: 'created_at' | 'updated_at' | 'status';
  sortOrder?: 'asc' | 'desc';
}

export interface GameNodeListResult {
  total: number;
  items: GameNode[];
}
