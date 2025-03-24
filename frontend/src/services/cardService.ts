import api from '@/api'
import { DataService } from './dataService'
import { mockGameCards } from '@/mocks/data/GameCard'
import type { GameCard } from '@/types/GameCard'

class CardService extends DataService {
  /**
   * 获取卡片列表
   */
  async getList(params?: any): Promise<{
    list: GameCard[]
    total: number
  }> {
    if (this.useMock()) {
      await this.mockDelay(300)
      
      const { page = 1, pageSize = 10 } = params || {}
      const start = (page - 1) * pageSize
      const end = start + pageSize
      
      return {
        list: mockGameCards.slice(start, end),
        total: mockGameCards.length
      }
    } else {
      try {
        const response = await api.card.getList(params)
        return this.safelyExtractListData<GameCard>(response)
      } catch (error) {
        console.error('获取卡片列表失败', error)
        return { list: [], total: 0 }
      }
    }
  }

  /**
   * 获取卡片详情
   */
  async getDetail(id: string): Promise<GameCard | null> {
    if (this.useMock()) {
      await this.mockDelay(300)
      const card = mockGameCards.find(item => item.id === id)
      return card || null
    } else {
      try {
        const response = await api.card.getDetail(id)
        return this.safelyExtractData<GameCard | null>(response, null)
      } catch (error) {
        console.error('获取卡片详情失败', error)
        return null
      }
    }
  }

  /**
   * 创建卡片
   */
  async create(data: any): Promise<string> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return 'mock-card-id-' + Date.now()
    } else {
      try {
        const response = await api.card.create(data)
        const result = this.safelyExtractData(response, { id: '' })
        return result.id || ''
      } catch (error) {
        console.error('创建卡片失败', error)
        return ''
      }
    }
  }

  /**
   * 更新卡片
   */
  async update(id: string, data: any): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.card.update(id, data)
        return true
      } catch (error) {
        console.error('更新卡片失败', error)
        return false
      }
    }
  }

  /**
   * 删除卡片
   */
  async delete(id: string): Promise<boolean> {
    if (this.useMock()) {
      await this.mockDelay(500)
      return true
    } else {
      try {
        await api.card.delete(id)
        return true
      } catch (error) {
        console.error('删除卡片失败', error)
        return false
      }
    }
  }
}

// 导出卡片服务实例
export const cardService = new CardService()

// 导出便捷方法
export const getCardList = (params?: any) => cardService.getList(params)
export const getCardDetail = (id: string) => cardService.getDetail(id)
export const createCard = (data: any) => cardService.create(data)
export const updateCard = (id: string, data: any) => cardService.update(id, data)
export const deleteCard = (id: string) => cardService.delete(id) 