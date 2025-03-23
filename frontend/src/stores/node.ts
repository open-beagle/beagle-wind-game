import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { Node, NodeQuery, NodeForm } from '../types/node'

export const useNodeStore = defineStore('node', () => {
  const nodes = ref<Node[]>([])
  const total = ref(0)
  const loading = ref(false)

  // 获取节点列表
  const getNodes = async (query: NodeQuery) => {
    loading.value = true
    try {
      // TODO: 替换为实际的 API 调用
      const mockData = {
        data: [
          {
            id: 'node-1',
            name: '游戏节点-1',
            status: 'online' as const,
            hardware: {
              cpu: 'Intel i9-13900K',
              memory: '128GB DDR5',
              disk: '2TB NVMe'
            },
            network: {
              ip: '192.168.1.100'
            },
            lastHeartbeat: new Date().toISOString()
          },
          {
            id: 'node-2',
            name: '游戏节点-2',
            status: 'offline' as const,
            hardware: {
              cpu: 'Intel i9-13900K',
              memory: '128GB DDR5',
              disk: '2TB NVMe'
            },
            network: {
              ip: '192.168.1.101'
            },
            lastHeartbeat: new Date(Date.now() - 3600000).toISOString()
          }
        ],
        total: 2
      }
      nodes.value = mockData.data
      total.value = mockData.total
      return mockData
    } finally {
      loading.value = false
    }
  }

  // 创建节点
  const createNode = async (data: NodeForm) => {
    // TODO: 替换为实际的 API 调用
    const newNode: Node = {
      id: `node-${Date.now()}`,
      name: data.name,
      status: 'offline' as const,
      hardware: data.hardware,
      network: data.network,
      lastHeartbeat: new Date().toISOString()
    }
    nodes.value.push(newNode)
    total.value++
    return Promise.resolve({ success: true })
  }

  // 更新节点
  const updateNode = async (data: NodeForm) => {
    // TODO: 替换为实际的 API 调用
    const index = nodes.value.findIndex(node => node.id === data.id)
    if (index !== -1) {
      const node = nodes.value[index]
      node.name = data.name
      node.hardware = data.hardware
      node.network = data.network
    }
    return Promise.resolve({ success: true })
  }

  // 删除节点
  const deleteNode = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    const index = nodes.value.findIndex(node => node.id === id)
    if (index !== -1) {
      nodes.value.splice(index, 1)
      total.value--
    }
    return Promise.resolve({ success: true })
  }

  // 获取节点监控数据
  const getNodeMonitor = async (id: string) => {
    // TODO: 替换为实际的 API 调用
    return Promise.resolve({
      cpu: {
        usage: 45,
        temperature: 65,
        cores: 24
      },
      memory: {
        total: 128,
        used: 64,
        free: 64
      },
      disk: {
        total: 2048,
        used: 512,
        free: 1536
      },
      network: {
        upload: 1024,
        download: 2048
      }
    })
  }

  return {
    nodes,
    total,
    loading,
    getNodes,
    createNode,
    updateNode,
    deleteNode,
    getNodeMonitor
  }
}) 