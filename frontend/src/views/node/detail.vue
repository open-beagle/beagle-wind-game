<template>
  <div class="node-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>节点详情</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleEdit">编辑</el-button>
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <div v-if="node" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="节点ID">{{ node.id }}</el-descriptions-item>
          <el-descriptions-item label="节点名称">{{ node.name }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(node.status)">
              {{ getStatusText(node.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="区域">{{ node.region }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ node.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ node.updatedAt }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">硬件资源</div>
        <el-descriptions v-if="node?.metrics && node?.resources" :column="2" border>
          <el-descriptions-item label="CPU">
            <el-progress
              :percentage="(node.metrics.cpuUsage / node.resources.cpu) * 100"
              :color="getResourceColor(node.metrics.cpuUsage, node.resources.cpu)"
            >
              <template #default>
                <span>{{ node.metrics.cpuUsage }}/{{ node.resources.cpu }} 核</span>
              </template>
            </el-progress>
          </el-descriptions-item>
          <el-descriptions-item label="内存">
            <el-progress
              :percentage="(node.metrics.memoryUsage / node.resources.memory) * 100"
              :color="getResourceColor(node.metrics.memoryUsage, node.resources.memory)"
            >
              <template #default>
                <span>{{ formatMemory(node.metrics.memoryUsage) }}/{{ formatMemory(node.resources.memory) }}</span>
              </template>
            </el-progress>
          </el-descriptions-item>
          <el-descriptions-item label="存储">
            <el-progress
              :percentage="(node.metrics.storageUsage / node.resources.storage) * 100"
              :color="getResourceColor(node.metrics.storageUsage, node.resources.storage)"
            >
              <template #default>
                <span>{{ formatStorage(node.metrics.storageUsage) }}/{{ formatStorage(node.resources.storage) }}</span>
              </template>
            </el-progress>
          </el-descriptions-item>
          <el-descriptions-item label="网络">
            <el-progress
              :percentage="(node.metrics.networkUsage / node.resources.network) * 100"
              :color="getResourceColor(node.metrics.networkUsage, node.resources.network)"
            >
              <template #default>
                <span>{{ formatNetwork(node.metrics.networkUsage) }}/{{ formatNetwork(node.resources.network) }}</span>
              </template>
            </el-progress>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">运行状态</div>
        <el-descriptions v-if="node?.metrics" :column="2" border>
          <el-descriptions-item label="运行时长">{{ formatUptime(node.metrics.uptime) }}</el-descriptions-item>
          <el-descriptions-item label="FPS">{{ node.metrics.fps }}</el-descriptions-item>
          <el-descriptions-item label="游戏实例数">{{ node.metrics.instanceCount }}</el-descriptions-item>
          <el-descriptions-item label="玩家数">{{ node.metrics.playerCount }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">网络信息</div>
        <el-descriptions v-if="node?.network" :column="2" border>
          <el-descriptions-item label="IP地址">{{ node.network.ip }}</el-descriptions-item>
          <el-descriptions-item label="端口">{{ node.network.port }}</el-descriptions-item>
          <el-descriptions-item label="协议">{{ node.network.protocol }}</el-descriptions-item>
          <el-descriptions-item label="带宽">{{ formatNetwork(node.network.bandwidth) }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">标签</div>
        <div v-if="node?.labels" class="tags-container">
          <el-tag
            v-for="(value, key) in node.labels"
            :key="key"
            class="label-tag"
          >
            {{ key }}: {{ value }}
          </el-tag>
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
import { mockGameNodes } from '@/mocks'
import type { GameNode, GameNodeStatus } from '@/types'
import type { FormInstance, FormRules } from 'element-plus'

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
const getNodeDetail = async () => {
  loading.value = true
  try {
    // 模拟API请求
    await new Promise(resolve => setTimeout(resolve, 300))
    const id = route.params.id as string
    node.value = mockGameNodes.find(n => n.id === id) || null
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

// 格式化网络
const formatNetwork = (bytes: number | undefined) => {
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

// 编辑节点
const handleEdit = () => {
  if (!node.value) return
  form.value = {
    id: node.value.id,
    name: node.value.name,
    status: node.value.status,
    region: node.value.region,
    network: {
      ip: node.value.network.ip,
      port: node.value.network.port,
      protocol: node.value.network.protocol
    }
  }
  dialogVisible.value = true
}

// 删除节点
const handleDelete = () => {
  ElMessageBox.confirm('确定要删除该节点吗？', '提示', {
    type: 'warning'
  }).then(() => {
    // 模拟删除
    const index = mockGameNodes.findIndex(n => n.id === node.value?.id)
    if (index > -1) {
      mockGameNodes.splice(index, 1)
      ElMessage.success('删除成功')
      router.push('/node')
    }
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate((valid) => {
    if (valid) {
      // 模拟更新
      const index = mockGameNodes.findIndex(n => n.id === form.value.id)
      if (index > -1) {
        mockGameNodes[index] = {
          ...mockGameNodes[index],
          name: form.value.name,
          status: form.value.status,
          region: form.value.region,
          network: {
            ...mockGameNodes[index].network,
            ip: form.value.network.ip,
            port: form.value.network.port,
            protocol: form.value.network.protocol
          }
        }
        ElMessage.success('更新成功')
        dialogVisible.value = false
        getNodeDetail()
      }
    }
  })
}

onMounted(() => {
  getNodeDetail()
})
</script>

<style scoped>
.node-detail-container {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-content {
  padding: 20px 0;
}

.section-title {
  font-size: 16px;
  font-weight: bold;
  margin: 20px 0 10px;
  padding-left: 10px;
  border-left: 4px solid #409EFF;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 20px;
}

.label-tag {
  margin-right: 8px;
  margin-bottom: 8px;
}

:deep(.el-descriptions) {
  margin-bottom: 20px;
}

:deep(.el-descriptions__label) {
  width: 120px;
}

:deep(.el-progress) {
  margin-bottom: 8px;
}
</style> 