// 游戏节点数据
export const mockNodes = [
  {
    id: '1',
    name: '节点-01',
    status: 'online',
    cpu: '4核',
    memory: '8GB',
    disk: '100GB',
    ip: '192.168.1.101',
    lastHeartbeat: '2024-03-20 18:30:00'
  },
  {
    id: '2',
    name: '节点-02',
    status: 'offline',
    cpu: '8核',
    memory: '16GB',
    disk: '200GB',
    ip: '192.168.1.102',
    lastHeartbeat: '2024-03-20 17:45:00'
  },
  {
    id: '3',
    name: '节点-03',
    status: 'online',
    cpu: '16核',
    memory: '32GB',
    disk: '500GB',
    ip: '192.168.1.103',
    lastHeartbeat: '2024-03-20 18:25:00'
  }
]

// 游戏平台数据
export const mockPlatforms = [
  {
    id: '1',
    name: 'Steam',
    type: 'steam',
    status: 'active',
    config: {
      apiKey: '******',
      region: 'cn'
    },
    lastCheck: '2024-03-20 18:30:00'
  },
  {
    id: '2',
    name: 'Epic Games',
    type: 'epic',
    status: 'active',
    config: {
      clientId: '******',
      clientSecret: '******'
    },
    lastCheck: '2024-03-20 18:25:00'
  },
  {
    id: '3',
    name: 'GOG',
    type: 'gog',
    status: 'inactive',
    config: {
      apiKey: '******'
    },
    lastCheck: '2024-03-20 17:45:00'
  }
]

// 游戏卡片数据
export const mockCards = [
  {
    id: '1',
    name: 'CS2',
    platform: 'Steam',
    type: 'fps',
    status: 'active',
    config: {
      maxPlayers: 64,
      tickRate: 128
    },
    lastUpdate: '2024-03-20 18:30:00'
  },
  {
    id: '2',
    name: 'Fortnite',
    platform: 'Epic Games',
    type: 'battle_royale',
    status: 'active',
    config: {
      maxPlayers: 100,
      region: 'asia'
    },
    lastUpdate: '2024-03-20 18:25:00'
  },
  {
    id: '3',
    name: 'Cyberpunk 2077',
    platform: 'GOG',
    type: 'rpg',
    status: 'inactive',
    config: {
      maxPlayers: 1,
      dlc: true
    },
    lastUpdate: '2024-03-20 17:45:00'
  }
]

// 游戏实例数据
export const mockInstances = [
  {
    id: '1',
    name: 'CS2-服务器-01',
    card: 'CS2',
    node: '节点-01',
    status: 'running',
    config: {
      port: 27015,
      map: 'de_dust2',
      maxPlayers: 32
    },
    createdAt: '2024-03-20 10:00:00'
  },
  {
    id: '2',
    name: 'Fortnite-服务器-01',
    card: 'Fortnite',
    node: '节点-02',
    status: 'stopped',
    config: {
      port: 7777,
      region: 'asia',
      maxPlayers: 100
    },
    createdAt: '2024-03-20 11:00:00'
  },
  {
    id: '3',
    name: 'Cyberpunk-服务器-01',
    card: 'Cyberpunk 2077',
    node: '节点-03',
    status: 'error',
    config: {
      port: 8080,
      dlc: true
    },
    createdAt: '2024-03-20 12:00:00'
  }
]

// 仪表盘数据
export const mockDashboard = {
  nodes: {
    total: 3,
    online: 2,
    offline: 1
  },
  platforms: {
    total: 3,
    active: 2,
    inactive: 1
  },
  cards: {
    total: 3,
    active: 2,
    inactive: 1
  },
  instances: {
    total: 3,
    running: 1,
    stopped: 1,
    error: 1
  },
  resources: {
    cpu: {
      total: 28,
      used: 15,
      free: 13
    },
    memory: {
      total: 56,
      used: 25,
      free: 31
    },
    disk: {
      total: 800,
      used: 300,
      free: 500
    }
  }
} 