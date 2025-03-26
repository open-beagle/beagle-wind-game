import axios from 'axios'
import { useUserStore } from '../stores'

// 使用固定的相对路径作为API基础URL
const API_BASE_URL = '/api/v1'

console.log('[API] 初始化API客户端，基础URL:', API_BASE_URL)

const api = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const userStore = useUserStore()
    if (userStore.token) {
      config.headers.Authorization = `Bearer ${userStore.token}`
    }
    
    // 添加详细的请求日志
    const fullUrl = `${config.baseURL}${config.url}`
    console.log(`[API请求] ${config.method?.toUpperCase()} ${fullUrl}`, {
      params: config.params || {},
      data: config.data || {},
      headers: config.headers
    })
    
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    // 添加详细的响应日志
    console.log(`[API响应] ${response.status} ${response.config.method?.toUpperCase()} ${response.config.url}`, {
      data: response.data,
      headers: response.headers
    })
    return response.data
  },
  (error) => {
    if (error.response) {
      console.error('响应错误:', {
        status: error.response.status,
        statusText: error.response.statusText,
        data: error.response.data,
        url: error.config.url,
        method: error.config.method?.toUpperCase(),
        headers: error.response.headers
      })
    } else if (error.request) {
      console.error('网络错误 (没有收到响应):', {
        request: error.request,
        url: error.config?.url,
        method: error.config?.method?.toUpperCase()
      })
    } else {
      console.error('请求配置错误:', error.message)
    }
    if (error.response?.status === 401) {
      const userStore = useUserStore()
      userStore.logout()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// API接口

// 平台相关API
export const platformApi = {
  // 获取平台列表
  getList: (params?: any) => api.get('/platforms', { params }),
  // 获取平台详情
  getDetail: (id: string) => api.get(`/platforms/${id}`),
  // 创建平台
  create: (data: any) => api.post('/platforms', data),
  // 更新平台
  update: (id: string, data: any) => api.put(`/platforms/${id}`, data),
  // 删除平台
  delete: (id: string) => api.delete(`/platforms/${id}`),
  // 获取平台远程访问链接
  getAccess: (id: string) => api.get(`/platforms/${id}/access`),
  // 刷新平台远程访问链接
  refreshAccess: (id: string) => api.post(`/platforms/${id}/access/refresh`)
}

// 节点相关API
export const nodeApi = {
  // 获取节点列表
  getList: (params?: any) => api.get('/nodes', { params }),
  // 获取节点详情
  getDetail: (id: string) => api.get(`/nodes/${id}`),
  // 创建节点
  create: (data: any) => api.post('/nodes', data),
  // 更新节点
  update: (id: string, data: any) => api.put(`/nodes/${id}`, data),
  // 删除节点
  delete: (id: string) => api.delete(`/nodes/${id}`),
  // 更新节点状态
  UpdateStatusState: (id: string, data: any) => api.put(`/nodes/${id}/status`, data)
}

// 游戏卡片相关API
export const cardApi = {
  // 获取卡片列表
  getList: (params?: any) => api.get('/cards', { params }),
  // 获取卡片详情
  getDetail: (id: string) => api.get(`/cards/${id}`),
  // 创建卡片
  create: (data: any) => api.post('/cards', data),
  // 更新卡片
  update: (id: string, data: any) => api.put(`/cards/${id}`, data),
  // 删除卡片
  delete: (id: string) => api.delete(`/cards/${id}`)
}

// 游戏实例相关API
export const instanceApi = {
  // 获取实例列表
  getList: (params?: any) => api.get('/instances', { params }),
  // 获取实例详情
  getDetail: (id: string) => api.get(`/instances/${id}`),
  // 创建实例
  create: (data: any) => api.post('/instances', data),
  // 更新实例
  update: (id: string, data: any) => api.put(`/instances/${id}`, data),
  // 删除实例
  delete: (id: string) => api.delete(`/instances/${id}`),
  // 启动实例
  start: (id: string) => api.post(`/instances/${id}/start`),
  // 停止实例
  stop: (id: string) => api.post(`/instances/${id}/stop`)
}

// 导出API
export default {
  platform: platformApi,
  node: nodeApi,
  card: cardApi,
  instance: instanceApi
} 