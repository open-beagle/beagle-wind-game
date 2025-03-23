import type { GamePlatform } from "@/types/GamePlatform";

export const mockGamePlatforms: GamePlatform[] = [
  {
    id: "lutris",
    name: "Lutris",
    os: "Linux",
    status: "active",
    description: "支持多种游戏源整合的游戏平台",
    version: "v0.5.18",
    image: "registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc:v1.0.8-desktop",
    resources: {
      cpu: 2,
      memory: 4096,
      storage: 50
    },
    config: {
      apiKey: "",
      apiUrl: "",
      callbackUrl: ""
    },
    environment: {
      dockerImage: "registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc",
      dockerTag: "v1.0.8-desktop",
      env: {
        WINEARCH: "win64",
        WINEPREFIX: "/data/wind"
      },
      volumes: {
        "/data/wind": "/data/wind"
      }
    },
    features: {
      gameTypes: ["pc", "linux"],
      platforms: ["windows", "linux"]
    },
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-03-15T10:30:00Z"
  },
  {
    id: "steam",
    name: "Steam",
    os: "Linux",
    status: "active",
    description: "全球最大的综合性数字游戏发行平台",
    version: "1.0.0.82",
    image: "registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc:v1.0.8-desktop",
    resources: {
      cpu: 2,
      memory: 4096,
      storage: 50
    },
    config: {
      apiKey: "",
      apiUrl: "",
      callbackUrl: ""
    },
    environment: {
      dockerImage: "registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc",
      dockerTag: "v1.0.8-desktop",
      env: {},
      volumes: {
        "/data/wind": "/data/wind"
      }
    },
    features: {
      gameTypes: ["pc"],
      platforms: ["windows", "linux", "macos"]
    },
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-03-15T10:30:00Z"
  },
  {
    id: "switch",
    name: "Nintendo Switch",
    os: "Linux",
    status: "active",
    description: "任天堂 Switch 游戏机模拟器",
    version: "2024.03.04",
    image: "registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc:v1.0.8-desktop",
    resources: {
      cpu: 4,
      memory: 8192,
      storage: 100
    },
    config: {
      apiKey: "",
      apiUrl: "",
      callbackUrl: ""
    },
    environment: {
      dockerImage: "registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc",
      dockerTag: "v1.0.8-desktop",
      env: {},
      volumes: {
        "/data/wind": "/data/wind"
      }
    },
    features: {
      gameTypes: ["console"],
      platforms: ["switch"]
    },
    createdAt: "2024-01-01T00:00:00Z",
    updatedAt: "2024-03-15T10:30:00Z"
  }
];
