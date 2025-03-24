<template>
  <div class="node-detail-container">
    <el-card v-loading="loading" class="detail-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">节点详情</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleEdit">编辑</el-button>
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <div v-if="node" class="detail-content">
        <!-- 基本信息 -->
        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">节点ID</div>
            <div class="item-value">{{ node.id }}</div>
          </div>
          <div class="detail-item">
            <div class="item-label">节点名称</div>
            <div class="item-value">{{ node.name }}</div>
          </div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">型号</div>
            <div class="item-value">{{ node.model || '-' }}</div>
          </div>
          <div class="detail-item">
            <div class="item-label">区域</div>
            <div class="item-value">{{ node.region || '-' }}</div>
          </div>
        </div>

        <!-- 运行状态 -->
        <div class="section-divider">
          <div class="section-title">运行状态</div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">状态</div>
            <div class="item-value">
              <el-tag :type="getStatusType(node.status)" class="status-tag">
                {{ getStatusText(node.status) }}
              </el-tag>
            </div>
          </div>
          <div class="detail-item">
            <div class="item-label">IP地址</div>
            <div class="item-value">
              <span v-if="getNodeIp(node) !== '-'" class="ip-info">
                <el-icon><Location /></el-icon> {{ getNodeIp(node) }}
              </span>
              <span v-else class="no-ip">未设置IP</span>
            </div>
          </div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">最后在线时间</div>
            <div class="item-value">{{ formatTime(node.last_online) }}</div>
          </div>
          <div class="detail-item">
            <div class="item-label">在线状态</div>
            <div class="item-value">
              <el-tag :type="node.online ? 'success' : 'info'" class="status-tag">
                {{ node.online ? '在线' : '离线' }}
              </el-tag>
            </div>
          </div>
        </div>

        <!-- 硬件配置 -->
        <div class="section-divider">
          <div class="section-title">硬件配置</div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">CPU</div>
            <div class="item-value">
              {{ node.hardware?.CPU || formatCpu(node.resources?.cpu) || '-' }}
            </div>
          </div>
          <div class="detail-item">
            <div class="item-label">GPU</div>
            <div class="item-value">
              {{ node.hardware?.GPU || formatGpu(node.resources?.gpu) || '-' }}
            </div>
          </div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">内存</div>
            <div class="item-value">
              {{ node.hardware?.RAM || formatMemory(node.resources?.memory) || '-' }}
            </div>
          </div>
          <div class="detail-item">
            <div class="item-label">磁盘</div>
            <div class="item-value">
              {{ node.hardware?.Storage || formatStorage(node.resources?.storage) || '-' }}
            </div>
          </div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">网络</div>
            <div class="item-value">
              {{ formatNetwork(node) }}
            </div>
          </div>
        </div>

        <!-- 资源使用情况 -->
        <div class="section-divider">
          <div class="section-title">资源使用情况</div>
        </div>

        <div v-if="node?.metrics && node?.resources" class="resource-section">
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">CPU使用率</div>
              <div class="item-value">
                <div class="progress-container">
                  <el-progress
                    :percentage="calculatePercentage(node.metrics?.cpuUsage, node.resources?.cpu)"
                    :color="getResourceColor(node.metrics?.cpuUsage || 0, node.resources?.cpu || 1)"
                    :stroke-width="16"
                    :text-inside="true"
                  >
                    <template #default>
                      <span>{{ node.metrics?.cpuUsage || 0 }}/{{ node.resources?.cpu || 0 }} 核</span>
                    </template>
                  </el-progress>
                </div>
              </div>
            </div>
          </div>
          
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">内存使用率</div>
              <div class="item-value">
                <div class="progress-container">
                  <el-progress
                    :percentage="calculatePercentage(node.metrics?.memoryUsage, node.resources?.memory)"
                    :color="getResourceColor(node.metrics?.memoryUsage || 0, node.resources?.memory || 1)"
                    :stroke-width="16"
                    :text-inside="true"
                  >
                    <template #default>
                      <span>{{ formatMemory(node.metrics?.memoryUsage) }}/{{ formatMemory(node.resources?.memory) }}</span>
                    </template>
                  </el-progress>
                </div>
              </div>
            </div>
          </div>
          
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">存储使用率</div>
              <div class="item-value">
                <div class="progress-container">
                  <el-progress
                    :percentage="calculatePercentage(node.metrics?.storageUsage, node.resources?.storage)"
                    :color="getResourceColor(node.metrics?.storageUsage || 0, node.resources?.storage || 1)"
                    :stroke-width="16"
                    :text-inside="true"
                  >
                    <template #default>
                      <span>{{ formatStorage(node.metrics?.storageUsage) }}/{{ formatStorage(node.resources?.storage) }}</span>
                    </template>
                  </el-progress>
                </div>
              </div>
            </div>
          </div>
          
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">网络使用率</div>
              <div class="item-value">
                <div class="progress-container">
                  <el-progress
                    :percentage="calculatePercentage(node.metrics?.networkUsage, node.resources?.network)"
                    :color="getResourceColor(node.metrics?.networkUsage || 0, node.resources?.network || 1)"
                    :stroke-width="16"
                    :text-inside="true"
                  >
                    <template #default>
                      <span>{{ formatNetwork(node.metrics?.networkUsage) }}/{{ formatNetwork(node.resources?.network) }}</span>
                    </template>
                  </el-progress>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 运行指标 -->
        <div class="section-divider">
          <div class="section-title">运行指标</div>
        </div>

        <div v-if="node?.metrics" class="metrics-section">
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">运行时长</div>
              <div class="item-value">{{ formatUptime(node.metrics?.uptime) }}</div>
            </div>
            <div class="detail-item">
              <div class="item-label">FPS</div>
              <div class="item-value">{{ node.metrics?.fps || 0 }}</div>
            </div>
          </div>

          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">游戏实例数</div>
              <div class="item-value">{{ node.metrics?.instanceCount || 0 }}</div>
            </div>
            <div class="detail-item">
              <div class="item-label">玩家数</div>
              <div class="item-value">{{ node.metrics?.playerCount || 0 }}</div>
            </div>
          </div>
        </div>

        <!-- 标签 -->
        <div class="section-divider">
          <div class="section-title">标签</div>
        </div>

        <div class="detail-row">
          <div class="detail-item full-width">
            <div class="item-label">标签信息</div>
            <div class="item-value">
              <div class="tags-container">
                <template v-if="node.labels && Object.keys(node.labels).length > 0">
                  <el-tag
                    v-for="(value, key) in node.labels"
                    :key="key"
                    class="label-tag"
                    type="info"
                    effect="plain"
                  >
                    {{ key }}: {{ value }}
                  </el-tag>
                </template>
                <span v-else class="no-tags">
                  无标签信息
                </span>
              </div>
            </div>
          </div>
        </div>

        <!-- 时间信息 -->
        <div class="section-divider">
          <div class="section-title">时间信息</div>
        </div>

        <div class="detail-row">
          <div class="detail-item">
            <div class="item-label">创建时间</div>
            <div class="item-value">{{ formatTime(node.created_at) }}</div>
          </div>
          <div class="detail-item">
            <div class="item-label">更新时间</div>
            <div class="item-value">{{ formatTime(node.updated_at) }}</div>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="编辑节点"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入节点名称" />
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择状态">
            <el-option label="在线" value="online" />
            <el-option label="离线" value="offline" />
            <el-option label="维护中" value="maintenance" />
          </el-select>
        </el-form-item>
        <el-form-item label="区域" prop="region">
          <el-input v-model="form.region" placeholder="请输入区域" />
        </el-form-item>
        <el-form-item label="IP地址" prop="network.ip">
          <el-input v-model="form.network.ip" placeholder="请输入IP地址" />
        </el-form-item>
        <el-form-item label="端口" prop="network.port">
          <el-input-number v-model="form.network.port" :min="1" :max="65535" />
        </el-form-item>
        <el-form-item label="协议" prop="network.protocol">
          <el-select v-model="form.network.protocol" placeholder="请选择协议">
            <el-option label="TCP" value="tcp" />
            <el-option label="UDP" value="udp" />
            <el-option label="HTTP" value="http" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { GameNode, GameNodeStatus } from '@/types'
import type { FormInstance, FormRules } from 'element-plus'
import { getNodeDetail, updateNode, deleteNode } from '@/services/nodeService'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const node = ref<GameNode | null>(null)
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()

const form = ref({
  id: '',
  name: '',
  status: 'online' as GameNodeStatus,
  region: '',
  network: {
    ip: '',
    port: 8080,
    protocol: ''
  }
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入节点名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ],
  region: [
    { required: true, message: '请输入区域', trigger: 'blur' }
  ],
  'network.ip': [
    { required: true, message: '请输入IP地址', trigger: 'blur' }
  ],
  'network.port': [
    { required: true, message: '请输入端口', trigger: 'blur' }
  ],
  'network.protocol': [
    { required: true, message: '请选择协议', trigger: 'change' }
  ]
}

// 获取节点详情
const fetchNodeDetail = async () => {
  loading.value = true
  try {
    const id = route.params.id as string
    const result = await getNodeDetail(id)
    
    if (result) {
      node.value = result
    } else {
      ElMessage.error('未找到节点数据')
      node.value = null
    }
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 状态类型
const statusTypeMap: Record<string, string> = {
  online: 'success',
  offline: 'info',
  maintenance: 'warning'
}

const getStatusType = (status: string) => statusTypeMap[status] || 'info'

// 状态文本
const statusTextMap: Record<string, string> = {
  online: '在线',
  offline: '离线',
  maintenance: '维护中'
}

const getStatusText = (status: string) => statusTextMap[status] || status

// 计算百分比，处理undefined值
const calculatePercentage = (used?: number, total?: number): number => {
  if (!used || !total || total === 0) return 0;
  return Math.min(Math.round((used / total) * 100), 100);
};

// 资源使用率颜色
const getResourceColor = (used: number, total: number) => {
  const percentage = (used / total) * 100
  if (percentage >= 90) return '#F56C6C'
  if (percentage >= 70) return '#E6A23C'
  return '#67C23A'
}

// 格式化内存
const formatMemory = (bytes: number | undefined) => {
  if (bytes === undefined) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let value = bytes
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex++
  }
  return `${value.toFixed(2)} ${units[unitIndex]}`
}

// 格式化存储
const formatStorage = (bytes: number | undefined) => {
  return formatMemory(bytes)
}

// 格式化运行时长
const formatUptime = (seconds: number | undefined) => {
  if (seconds === undefined) return '0秒'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const remainingSeconds = seconds % 60

  const parts = []
  if (days > 0) parts.push(`${days}天`)
  if (hours > 0) parts.push(`${hours}小时`)
  if (minutes > 0) parts.push(`${minutes}分钟`)
  if (remainingSeconds > 0) parts.push(`${remainingSeconds}秒`)

  return parts.join(' ') || '0秒'
}

// 格式化时间显示
const formatTime = (timeStr: string | null | undefined) => {
  if (!timeStr) return '-';
  try {
    // 解析日期
    const date = new Date(timeStr);
    if (isNaN(date.getTime())) return timeStr;
    
    // 创建上海时区的日期格式化器
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'Asia/Shanghai'
    }).format(date);
  } catch (e) {
    return '-';
  }
};

// 编辑节点
const handleEdit = () => {
  if (!node.value) return
  
  form.value = {
    id: node.value.id,
    name: node.value.name,
    status: node.value.status,
    region: node.value.region || '',
    network: {
      ip: node.value.network?.ip || '',
      port: node.value.network?.port || 8080,
      protocol: node.value.network?.protocol || ''
    }
  }
  dialogVisible.value = true
}

// 删除节点
const handleDelete = () => {
  if (!node.value) return
  
  ElMessageBox.confirm(`确定要删除节点 "${node.value.name}" 吗？此操作不可恢复`, '删除确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    try {
      const success = await deleteNode(node.value?.id || '')
      if (success) {
        ElMessage.success('删除成功')
        router.push('/node')
      } else {
        ElMessage.error('删除失败')
      }
    } catch (error) {
      ElMessage.error('删除失败')
      console.error('删除失败:', error)
    }
  }).catch(() => {
    // 用户取消删除
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        const success = await updateNode(form.value.id, {
          name: form.value.name,
          status: form.value.status,
          region: form.value.region,
          network: {
            ip: form.value.network.ip,
            port: form.value.network.port,
            protocol: form.value.network.protocol
          }
        })
        
        if (success) {
          ElMessage.success('更新成功')
          dialogVisible.value = false
          fetchNodeDetail()
        } else {
          ElMessage.error('更新失败')
        }
      } catch (error) {
        ElMessage.error('更新失败')
        console.error('更新失败:', error)
      }
    }
  })
}

// 获取节点IP地址
const getNodeIp = (node: any): string => {
  // 直接检查顶层ip属性
  if (node.ip) return node.ip;
  
  // 检查network对象中可能的ip字段(大小写不敏感)
  if (node.network) {
    // 检查常见格式
    if (node.network.ip) return node.network.ip;
    if (node.network.IP) return node.network.IP;
    
    // 遍历所有属性
    for (const key in node.network) {
      const lowerKey = key.toLowerCase();
      if (lowerKey === 'ip' || lowerKey.includes('ip')) {
        return node.network[key];
      }
    }
  }
  
  // 检查是否在hardware.Network或resources.ip等嵌套层级
  if (node.hardware?.Network?.includes?.('192.168.')) {
    return node.hardware.Network.split(' ')[0];
  }
  
  // 检查labels
  if (node.labels?.ip) return node.labels.ip;
  
  return '-';
};

// 格式化CPU信息
const formatCpu = (cpu?: number): string => {
  if (!cpu) return '-';
  return `${cpu} 核`;
};

// 格式化GPU信息
const formatGpu = (gpu?: any): string => {
  if (!gpu) return '-';
  if (typeof gpu === 'string') return gpu;
  if (gpu.model) {
    if (gpu.memory) {
      return `${gpu.model} (${formatMemory(gpu.memory)})`;
    }
    return gpu.model;
  }
  return '-';
};

// 格式化网络信息
const formatNetwork = (node: any): string => {
  if (node.hardware?.Network) return node.hardware.Network;
  if (node.network?.speed) return node.network.speed;
  if (node.network?.bandwidth) return `${node.network.bandwidth} Mbps`;
  if (node.resources?.network) return `${node.resources.network} Mbps`;
  return '-';
};

onMounted(() => {
  fetchNodeDetail()
})
</script>

<style scoped>
.node-detail-container {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
}

.detail-card {
  height: 100%;
  overflow: auto;
  background-color: #fff;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  border-bottom: 1px solid #EBEEF5;
}

.card-title {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
}

.header-actions {
  display: flex;
  gap: 10px;
}

.detail-content {
  padding: 0;
}

.section-divider {
  margin-top: 20px;
  margin-bottom: 10px;
}

.section-title {
  position: relative;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  border-left: 4px solid #409EFF;
  padding-left: 10px;
  line-height: 20px;
  margin: 0;
}

.detail-row {
  display: flex;
  border-bottom: 1px solid #EBEEF5;
  background-color: #fff;
}

.detail-row:nth-child(even) {
  background-color: #fff;
}

.detail-item {
  display: flex;
  width: 50%;
  min-height: 40px;
}

.detail-item.full-width {
  width: 100%;
}

.item-label {
  width: 120px;
  min-width: 120px;
  padding: 12px 15px;
  background-color: #f5f7fa;
  border-right: 1px solid #EBEEF5;
  color: #606266;
  text-align: left;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.item-value {
  flex: 1;
  padding: 12px 15px;
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #303133;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.label-tag {
  margin: 2px;
}

.no-tags {
  color: #909399;
  font-style: italic;
}

.progress-container {
  width: 100%;
}

/* 进度条样式 */
:deep(.el-progress) {
  margin-bottom: 0;
  width: 100%;
}

:deep(.el-progress-bar__outer) {
  height: 16px !important;
  border-radius: 4px;
}

:deep(.el-progress-bar__inner) {
  border-radius: 4px;
}

:deep(.el-progress-bar__innerText) {
  font-size: 12px;
  line-height: 16px;
  color: #fff;
}

/* 标签样式 */
:deep(.el-tag) {
  display: inline-flex;
  align-items: center;
  height: 24px;
  font-size: 12px;
  padding: 0 8px;
}

.status-tag {
  min-width: 60px;
  text-align: center;
  justify-content: center;
}

/* IP地址样式 */
.ip-info {
  display: flex;
  align-items: center;
  gap: 5px;
}

.no-ip {
  color: #909399;
  font-style: italic;
}

/* 资源部分和指标部分样式 */
.resource-section,
.metrics-section {
  width: 100%;
}

.resource-section .detail-row,
.metrics-section .detail-row {
  width: 100%;
}

.resource-section .detail-item,
.metrics-section .detail-item {
  width: 100%;
}

.resource-section .item-value,
.metrics-section .item-value {
  flex: 1;
}
</style> 