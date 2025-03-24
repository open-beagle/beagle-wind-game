// 基础类型定义
export type GamePlatformStatus = "active" | "maintenance" | "inactive";
export type GamePlatformOS = "Linux" | "Windows" | "macOS";

export interface GamePlatformFile {
  id: string;
  type: string;
  url: string;
}

export interface GamePlatformInstallerCommand {
  command?: string;
  move?: {
    src: string;
    dst: string;
  };
  chmodx?: string;
  extract?: {
    file: string;
    dst: string;
  };
}

// 查询参数类型
export interface GamePlatformQuery {
  page: number;
  pageSize: number;
  keyword?: string;
  type?: string;
}

// 主类型定义
export interface GamePlatform {
  id: string;
  name: string;
  version: string;
  type: string;
  image: string;
  bin: string;
  os: GamePlatformOS;
  files: GamePlatformFile[];
  features: string[];
  config: Record<string, string>;
  installer?: GamePlatformInstallerCommand[];
  status: GamePlatformStatus;
  created_at: string;
  updated_at: string;
}

// 表单类型使用 Partial<GamePlatform>，使所有字段变为可选
export type GamePlatformForm = Partial<GamePlatform>;
