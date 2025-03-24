import api from '@/api'
import { DataService } from './dataService'
import { mockGameInstances } from '@/mocks/data/GameInstance'
import type { GameInstance } from '@/types/GameInstance'

class InstanceService extends DataService {
  /**
   * 获取实例列表
   */
  async getList(params?: any): Promise<{
    list: GameInstance[]
    total: number
  }> {
    if (this.useMock()) {
      await this.mockDelay(300)
      
      const { page = 1, pageSize = 10 } = params || {}
      const start = (page - 1) * pageSize
      const end = start + pageSize
      
      return {
        list: mockGameInstances.slice(start, end),
        total: mockGameInstances.length
      }
    } else {
      try {
        const response = await api.instance.getList(params)
        return this.safelyExtractListData<GameInstance>(response)
      } catch (error) {
        console.error('获取实例列表失败', error)
        return { list: [], total: 0 }
      }
    }
  }

  /**
   * 获取实例详情
   */
  async getDetail(id: string): Promise<GameInstance | null> {
    if (this.useMock()) {
      await this.mockDelay(300)
      const instance = mockGameInstances.find(item => item.id === id)
      return instance || null
    } else {
      try {
        const response = await api.instance.getDetail(id)
        return this.safelyExtractData<GameInstance | null>(response, null)
      } catch (error) {
        console.error('获取实例详情失败', error)
        return null
      }
    }
  }

  /**
   * 创建实例
   */
  async create(data: any): Promise<string> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return 'mock-instance-id-' + Date.now()
    } else {
      try {
        const response = await api.instance.create(data)
        const result = this.safelyExtractData(response, { id: '' })
        return result.id || ''
      } catch (error) {
        console.error('创建实例失败', error)
        return ''
      }
    }
  }

  /**
   * 更新实例
   */
  async update(id: string, data: any): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.instance.update(id, data)
        return true
      } catch (error) {
        console.error('更新实例失败', error)
        return false
      }
    }
  }

  /**
   * 删除实例
   */
  async delete(id: string): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.instance.delete(id)
        return true
      } catch (error) {
        console.error('删除实例失败', error)
        return false
      }
    }
  }
  
  /**
   * 启动实例
   */
  async start(id: string): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.instance.start(id)
        return true
      } catch (error) {
        console.error('启动实例失败', error)
        return false
      }
    }
  }
  
  /**
   * 停止实例
   */
  async stop(id: string): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.instance.stop(id)
        return true
      } catch (error) {
        console.error('停止实例失败', error)
        return false
      }
    }
  }
}

// 导出实例服务实例
export const instanceService = new InstanceService()

// 导出便捷方法
export const getInstanceList = (params?: any) => instanceService.getList(params)
export const getInstanceDetail = (id: string) => instanceService.getDetail(id)
export const createInstance = (data: any) => instanceService.create(data)
export const updateInstance = (id: string, data: any) => instanceService.update(id, data)
export const deleteInstance = (id: string) => instanceService.delete(id)
export const startInstance = (id: string) => instanceService.start(id)
export const stopInstance = (id: string) => instanceService.stop(id) 