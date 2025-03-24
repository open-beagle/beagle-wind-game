import type { GameNode } from '@/types/GameNode'
import { mockGameNodes } from '@/mocks/data/GameNode'
import api from '@/api'
import { DataService } from './dataService'

class NodeService extends DataService {
  /**
   * 获取节点列表
   */
  async getList(params?: any): Promise<{
    list: GameNode[]
    total: number
  }> {
    if (this.useMock()) {
      console.log('[节点服务] 使用Mock数据获取节点列表', params)
      // 使用mock数据
      await this.mockDelay(300)
      
      const { page = 1, pageSize = 10 } = params || {}
      const start = (page - 1) * pageSize
      const end = start + pageSize
      
      return {
        list: mockGameNodes.slice(start, end),
        total: mockGameNodes.length
      }
    } else {
      console.log('[节点服务] 使用API获取节点列表', params)
      // 使用真实API
      try {
        const response = await api.node.getList(params)
        return this.safelyExtractListData<GameNode>(response)
      } catch (error) {
        console.error('获取节点列表失败', error)
        return { list: [], total: 0 }
      }
    }
  }

  /**
   * 获取节点详情
   */
  async getDetail(id: string): Promise<GameNode | null> {
    if (this.useMock()) {
      await this.mockDelay(300)
      const node = mockGameNodes.find(item => item.id === id)
      return node || null
    } else {
      try {
        const response = await api.node.getDetail(id)
        return this.safelyExtractData<GameNode | null>(response, null)
      } catch (error) {
        console.error('获取节点详情失败', error)
        return null
      }
    }
  }

  /**
   * 创建节点
   */
  async create(data: any): Promise<string> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return 'mock-node-id-' + Date.now()
    } else {
      try {
        const response = await api.node.create(data)
        const result = this.safelyExtractData(response, { id: '' })
        return result.id || ''
      } catch (error) {
        console.error('创建节点失败', error)
        return ''
      }
    }
  }

  /**
   * 更新节点
   */
  async update(id: string, data: any): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.node.update(id, data)
        return true
      } catch (error) {
        console.error('更新节点失败', error)
        return false
      }
    }
  }

  /**
   * 删除节点
   */
  async delete(id: string): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.node.delete(id)
        return true
      } catch (error) {
        console.error('删除节点失败', error)
        return false
      }
    }
  }
}

// 导出节点服务实例
export const nodeService = new NodeService()

// 导出便捷方法
export const getNodeList = (params?: any) => nodeService.getList(params)
export const getNodeDetail = (id: string) => nodeService.getDetail(id)
export const createNode = (data: any) => nodeService.create(data)
export const updateNode = (id: string, data: any) => nodeService.update(id, data)
export const deleteNode = (id: string) => nodeService.delete(id) 