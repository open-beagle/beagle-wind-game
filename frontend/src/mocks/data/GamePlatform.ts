import type { GamePlatform } from '@/types'

export const mockGamePlatforms: GamePlatform[] = [
  {
    id: 'lutris',
    name: 'Lutris',
    version: 'v0.5.18',
    os: 'Linux',
    status: 'active',
    type: 'gaming',
    image: 'registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc:v1.0.8-desktop',
    bin: '/usr/games/lutris',
    files: [
      {
        id: 'data',
        type: 'system',
        url: 'https://github.com/open-beagle/beagle-wind-game/releases/download/v1.0.0/lutris-data.tar.xz'
      }
    ],
    features: [
      '支持多种游戏源整合',
      'Wine/Proton 支持',
      '自定义运行配置',
      '游戏安装脚本'
    ],
    config: {
      wine: '8.0',
      dxvk: '2.2',
      vkd3d: '2.9',
      python: '3.10'
    },
    created_at: '2024-03-01T00:00:00Z',
    updated_at: '2024-03-01T00:00:00Z'
  },
  {
    id: 'steam',
    name: 'Steam',
    version: '1.0.0.82',
    os: 'Linux',
    status: 'active',
    type: 'gaming',
    image: 'registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc:v1.0.8-desktop',
    bin: '/usr/bin/steam',
    files: [
      {
        id: 'data',
        type: 'system',
        url: 'https://github.com/open-beagle/beagle-wind-game/releases/download/v1.0.0/steam-data.tar.xz'
      }
    ],
    features: [
      'Steam Play 兼容层',
      '云存档同步',
      '好友系统',
      '成就系统',
      '创意工坊'
    ],
    config: {
      proton: '8.0',
      'shader-cache': 'enabled',
      'remote-play': 'enabled',
      broadcast: 'enabled'
    },
    created_at: '2024-03-01T00:00:00Z',
    updated_at: '2024-03-01T00:00:00Z'
  },
  {
    id: 'switch',
    name: 'Nintendo Switch',
    version: '2024.03.04',
    os: 'Linux',
    status: 'active',
    type: 'emulator',
    image: 'registry.cn-qingdao.aliyuncs.com/wod/beagle-wind-vnc:v1.0.8-desktop',
    bin: '$HOME/.local/bin/yuzu',
    files: [
      {
        id: 'data',
        type: 'system',
        url: 'https://github.com/open-beagle/beagle-wind-game/releases/download/v1.0.0/switch-data.tar.xz'
      },
      {
        id: 'appimage',
        type: 'appimage',
        url: 'https://github.com/GamerSwir2/yuzu-linux/releases/download/stable/yuzu-mainline-20240304-537296095.AppImage'
      },
      {
        id: 'keys',
        type: 'keys',
        url: 'https://github.com/GamerSwir2/yuzu-linux/releases/download/stable/prod.keys.zip'
      },
      {
        id: 'firmware',
        type: 'firmware',
        url: 'https://github.com/GamerSwir2/yuzu-linux/releases/download/stable/firmware-16.1.0.zip'
      }
    ],
    features: [
      '掌机/主机双模式',
      'Joy-Con 控制器',
      '本地多人游戏',
      '在线联机服务'
    ],
    config: {
      mode: 'docked',
      resolution: '1080p',
      wifi: 'enabled',
      bluetooth: 'enabled'
    },
    installer: [
      { command: 'mkdir -p $HOME/.local/bin' },
      {
        move: {
          src: 'appimage',
          dst: '$HOME/.local/bin/yuzu/yuzu-mainline-20240304-537296095.AppImage'
        }
      },
      { chmodx: 'yuzu-mainline-20240304-537296095.AppImage' },
      {
        move: {
          src: 'keys',
          dst: '$HOME/.local/share/yuzu/keys'
        }
      },
      {
        extract: {
          file: 'firmware',
          dst: '$HOME/.local/share/yuzu/firmware'
        }
      },
      { command: 'ln -sf $GAMEDIR/yuzu-mainline-20240304-537296095.AppImage $HOME/.local/bin/yuzu' }
    ],
    created_at: '2024-03-01T00:00:00Z',
    updated_at: '2024-03-01T00:00:00Z'
  }
]
