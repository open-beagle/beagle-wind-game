import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { GameCard, GameQuery, GameForm, GameInstance, GameInstanceForm, GamePlatform } from '../types/game'

export const useGameStore = defineStore('game', () => {
  const gameCards = ref<GameCard[]>([])
  const gameInstances = ref<GameInstance[]>([])
  const total = ref(0)
  const loading = ref(false)

  // 获取游戏卡片列表
  const getGameCards = async (query: GameQuery) => {
    loading.value = true
    try {
      // TODO: 替换为实际的 API 调用
      const mockData = {
        data: [
          {
            id: 'game-1',
            name: '赛博朋克 2077',
            platformId: 'platform-1',
            platform: {
              id: 'platform-1',
              name: 'Steam',
              type: 'steam' as const,
              status: 'active' as const,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString()
            },
            description: '一款开放世界动作冒险游戏',
            status: 'active' as const,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          },
          {
            id: 'game-2',
            name: '艾尔登法环',
            platformId: 'platform-2',
            platform: {
              id: 'platform-2',
              name: 'Epic',
              type: 'epic' as const,
              status: 'active' as const,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString()
            },
            description: '一款动作角色扮演游戏',
            status: 'active' as const,
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          }
        ],
        total: 2
      }
      gameCards.value = mockData.data
      total.value = mockData.total
      return mockData
    } finally {
      loading.value = false
    }
  }

  // 创建游戏卡片
  const createGameCard = async (data: GameForm) => {
    // TODO: 替换为实际的 API 调用
    return Promise.resolve({ success: true })
  }

  // 更新游戏卡片
  const updateGameCard = async (data: GameForm) => {
    // TODO: 替换为实际的 API 调用
    return Promise.resolve({ success: true })
  }

  // 删除游戏卡片
  const deleteGameCard = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    return Promise.resolve({ success: true })
  }

  // 获取游戏实例列表
  const getGameInstances = async (query: GameQuery) => {
    loading.value = true
    try {
      // TODO: 替换为实际的 API 调用
      const mockData = {
        data: [
          {
            id: 'instance-1',
            gameCardId: 'game-1',
            gameCard: {
              id: 'game-1',
              name: '赛博朋克 2077',
              platformId: 'platform-1',
              platform: {
                id: 'platform-1',
                name: 'Steam',
                type: 'steam' as const,
                status: 'active' as const,
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString()
              },
              status: 'active' as const,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString()
            },
            nodeId: 'node-1',
            node: {
              id: 'node-1',
              name: '游戏节点-1'
            },
            status: 'running' as const,
            config: {
              maxPlayers: 100,
              port: 27015,
              settings: {
                difficulty: 'normal',
                maxLevel: 50
              }
            },
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          },
          {
            id: 'instance-2',
            gameCardId: 'game-2',
            gameCard: {
              id: 'game-2',
              name: '艾尔登法环',
              platformId: 'platform-2',
              platform: {
                id: 'platform-2',
                name: 'Epic',
                type: 'epic' as const,
                status: 'active' as const,
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString()
              },
              status: 'active' as const,
              createdAt: new Date().toISOString(),
              updatedAt: new Date().toISOString()
            },
            nodeId: 'node-2',
            node: {
              id: 'node-2',
              name: '游戏节点-2'
            },
            status: 'stopped' as const,
            config: {
              maxPlayers: 50,
              port: 27016,
              settings: {
                difficulty: 'hard',
                maxLevel: 100
              }
            },
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
          }
        ],
        total: 2
      }
      gameInstances.value = mockData.data
      total.value = mockData.total
      return mockData
    } finally {
      loading.value = false
    }
  }

  // 创建游戏实例
  const createGameInstance = async (data: GameInstanceForm) => {
    // TODO: 替换为实际的 API 调用
    const newInstance: GameInstance = {
      id: `instance-${Date.now()}`,
      gameCardId: data.gameCardId,
      gameCard: gameCards.value.find(card => card.id === data.gameCardId) as GameCard,
      nodeId: data.nodeId,
      node: {
        id: data.nodeId,
        name: '游戏节点-' + data.nodeId
      },
      status: 'stopped' as const,
      config: data.config,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    }
    gameInstances.value.push(newInstance)
    total.value++
    return Promise.resolve({ success: true })
  }

  // 更新游戏实例
  const updateGameInstance = async (data: GameInstanceForm) => {
    // TODO: 替换为实际的 API 调用
    const index = gameInstances.value.findIndex(instance => instance.id === data.id)
    if (index !== -1) {
      const instance = gameInstances.value[index]
      instance.gameCardId = data.gameCardId
      instance.gameCard = gameCards.value.find(card => card.id === data.gameCardId) as GameCard
      instance.nodeId = data.nodeId
      instance.node = {
        id: data.nodeId,
        name: '游戏节点-' + data.nodeId
      }
      instance.config = data.config
      instance.updatedAt = new Date().toISOString()
    }
    return Promise.resolve({ success: true })
  }

  // 删除游戏实例
  const deleteGameInstance = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    const index = gameInstances.value.findIndex(instance => instance.id === id)
    if (index !== -1) {
      gameInstances.value.splice(index, 1)
      total.value--
    }
    return Promise.resolve({ success: true })
  }

  // 启动游戏实例
  const startGameInstance = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    const instance = gameInstances.value.find(instance => instance.id === id)
    if (instance) {
      instance.status = 'running' as const
      instance.updatedAt = new Date().toISOString()
    }
    return Promise.resolve({ success: true })
  }

  // 停止游戏实例
  const stopGameInstance = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    const instance = gameInstances.value.find(instance => instance.id === id)
    if (instance) {
      instance.status = 'stopped' as const
      instance.updatedAt = new Date().toISOString()
    }
    return Promise.resolve({ success: true })
  }

  // 重启游戏实例
  const restartGameInstance = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    const instance = gameInstances.value.find(instance => instance.id === id)
    if (instance) {
      instance.status = 'stopped' as const
      instance.updatedAt = new Date().toISOString()
      setTimeout(() => {
        instance.status = 'running' as const
        instance.updatedAt = new Date().toISOString()
      }, 1000)
    }
    return Promise.resolve({ success: true })
  }

  return {
    gameCards,
    gameInstances,
    total,
    loading,
    getGameCards,
    createGameCard,
    updateGameCard,
    deleteGameCard,
    getGameInstances,
    createGameInstance,
    updateGameInstance,
    deleteGameInstance,
    startGameInstance,
    stopGameInstance,
    restartGameInstance
  }
}) 