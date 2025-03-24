import { shouldUseMock } from '@/config/mock'

/**
 * 数据服务基类
 * 提供统一的数据访问逻辑和Mock数据支持
 */
export class DataService {
  /**
   * 判断是否使用Mock数据
   * @returns {boolean} 是否使用Mock数据
   */
  protected useMock(): boolean {
    return shouldUseMock()
  }

  /**
   * 模拟API延迟
   * @param {number} ms 延迟毫秒数，默认500ms
   * @returns {Promise<void>}
   */
  protected async mockDelay(ms: number = 500): Promise<void> {
    if (this.useMock()) {
      await new Promise(resolve => setTimeout(resolve, ms))
    }
  }

  // 处理API错误
  protected handleApiError(error: any, errorMessage: string): never {
    console.error(errorMessage, error)
    throw new Error(errorMessage)
  }

  // 安全地从API响应中提取数据
  protected safelyExtractData<T>(response: any, fallback: T): T {
    if (!response) return fallback
    
    // 检查response是否是直接的数据对象
    if (typeof response === 'object') {
      // 如果response有data字段，返回data
      if ('data' in response) {
        return response.data ?? fallback
      }
      // 如果response本身就是数据对象，直接返回
      return response as T
    }
    
    return fallback
  }
  
  /**
   * 安全地从API响应中提取列表数据
   * @param response API响应数据
   * @returns {{ list: T[], total: number }} 标准化的列表数据
   */
  protected safelyExtractListData<T>(response: any): { list: T[], total: number } {
    if (!response) {
      return { list: [], total: 0 }
    }

    // 处理不同的响应数据结构
    if (response.data && Array.isArray(response.data)) {
      return {
        list: response.data,
        total: response.total || response.data.length
      }
    }

    // 处理不同的响应数据结构
    if (response.items && Array.isArray(response.items)) {
      return {
        list: response.items,
        total: response.total || response.items.length
      }
    }

    if (Array.isArray(response)) {
      return {
        list: response,
        total: response.length
      }
    }

    console.warn('无法解析的API响应数据结构:', response)
    return { list: [], total: 0 }
  }
} 