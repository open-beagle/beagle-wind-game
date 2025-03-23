<template>
  <div class="instance-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>实例详情</span>
          <div class="header-actions">
            <el-button
              v-if="instance?.status === 'stopped'"
              type="success"
              @click="handleStart"
            >
              启动
            </el-button>
            <el-button
              v-if="instance?.status === 'running'"
              type="warning"
              @click="handleStop"
            >
              停止
            </el-button>
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <div v-if="instance" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="实例ID">{{ instance.id }}</el-descriptions-item>
          <el-descriptions-item label="实例名称">{{ instance.name }}</el-descriptions-item>
          <el-descriptions-item label="游戏">
            <el-tag v-if="instance.gameCard" :type="getPlatformType(instance.gameCard.platform.type)">
              {{ instance.gameCard.name }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="节点">
            <el-tag v-if="instance.node" :type="getNodeStatusType(instance.node.status)">
              {{ instance.node.name }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(instance.status)">
              {{ getStatusText(instance.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ instance.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ instance.updatedAt }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">配置信息</div>
        <el-descriptions v-if="instance?.config" :column="2" border>
          <el-descriptions-item label="最大玩家数">{{ instance.config.maxPlayers }}</el-descriptions-item>
          <el-descriptions-item label="端口">{{ instance.config.port }}</el-descriptions-item>
          <el-descriptions-item label="地图">{{ instance.config.settings.map }}</el-descriptions-item>
          <el-descriptions-item label="难度">{{ instance.config.settings.difficulty }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">运行状态</div>
        <el-descriptions v-if="instance?.metrics" :column="2" border>
          <el-descriptions-item label="当前玩家数">{{ instance.metrics.playerCount }}</el-descriptions-item>
          <el-descriptions-item label="FPS">{{ instance.metrics.fps }}</el-descriptions-item>
          <el-descriptions-item label="CPU使用率">
            <el-progress
              :percentage="instance.metrics.cpuUsage"
              :color="getResourceColor(instance.metrics.cpuUsage)"
            />
          </el-descriptions-item>
          <el-descriptions-item label="内存使用率">
            <el-progress
              :percentage="instance.metrics.memoryUsage"
              :color="getResourceColor(instance.metrics.memoryUsage)"
            />
          </el-descriptions-item>
          <el-descriptions-item label="运行时长">{{ formatUptime(instance.metrics.uptime) }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">日志</div>
        <div v-if="instance?.logs" class="log-container">
          <el-scrollbar height="200px">
            <pre class="log-content">{{ instance.logs.join('\n') }}</pre>
          </el-scrollbar>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mockGameInstances } from '@/mocks'
import type { GameInstance } from '@/types'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const instance = ref<GameInstance | null>(null)

// 获取实例详情
const getInstanceDetail = async () => {
  loading.value = true
  try {
    // 模拟API请求
    await new Promise(resolve => setTimeout(resolve, 300))
    const id = route.params.id as string
    instance.value = mockGameInstances.find(i => i.id === id) || null
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 状态类型
const statusTypeMap: Record<string, string> = {
  running: 'success',
  stopped: 'info',
  error: 'danger',
  starting: 'warning',
  stopping: 'warning'
}

const getStatusType = (status: string) => statusTypeMap[status] || 'info'

// 状态文本
const statusTextMap: Record<string, string> = {
  running: '运行中',
  stopped: '已停止',
  error: '错误',
  starting: '启动中',
  stopping: '停止中'
}

const getStatusText = (status: string) => statusTextMap[status] || status

// 节点状态类型
const nodeStatusTypeMap: Record<string, string> = {
  online: 'success',
  offline: 'info',
  maintenance: 'warning'
}

const getNodeStatusType = (status: string) => nodeStatusTypeMap[status] || 'info'

// 平台类型
const platformTypeMap: Record<string, string> = {
  steam: 'primary',
  epic: 'success',
  gog: 'warning',
  xbox: 'info',
  psn: 'danger',
  nintendo: 'success'
}

const getPlatformType = (type: string) => platformTypeMap[type] || 'info'

// 资源使用率颜色
const getResourceColor = (percentage: number) => {
  if (percentage >= 90) return '#F56C6C'
  if (percentage >= 70) return '#E6A23C'
  return '#67C23A'
}

// 格式化运行时长
const formatUptime = (seconds: number) => {
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

// 启动实例
const handleStart = () => {
  if (!instance.value) return
  ElMessageBox.confirm('确定要启动该实例吗？', '提示', {
    type: 'warning'
  }).then(() => {
    // 模拟启动
    const index = mockGameInstances.findIndex(i => i.id === instance.value?.id)
    if (index > -1) {
      mockGameInstances[index] = {
        ...mockGameInstances[index],
        status: 'running',
        updatedAt: new Date().toISOString()
      }
      ElMessage.success('启动成功')
      getInstanceDetail()
    }
  })
}

// 停止实例
const handleStop = () => {
  if (!instance.value) return
  ElMessageBox.confirm('确定要停止该实例吗？', '提示', {
    type: 'warning'
  }).then(() => {
    // 模拟停止
    const index = mockGameInstances.findIndex(i => i.id === instance.value?.id)
    if (index > -1) {
      mockGameInstances[index] = {
        ...mockGameInstances[index],
        status: 'stopped',
        updatedAt: new Date().toISOString()
      }
      ElMessage.success('停止成功')
      getInstanceDetail()
    }
  })
}

// 删除实例
const handleDelete = () => {
  if (!instance.value) return
  ElMessageBox.confirm('确定要删除该实例吗？', '提示', {
    type: 'warning'
  }).then(() => {
    // 模拟删除
    const index = mockGameInstances.findIndex(i => i.id === instance.value?.id)
    if (index > -1) {
      mockGameInstances.splice(index, 1)
      ElMessage.success('删除成功')
      router.push('/instance')
    }
  })
}

onMounted(() => {
  getInstanceDetail()
})
</script>

<style scoped>
.instance-detail-container {
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

.log-container {
  background-color: #1e1e1e;
  border-radius: 4px;
  padding: 10px;
  margin-bottom: 20px;
}

.log-content {
  margin: 0;
  color: #fff;
  font-family: monospace;
  white-space: pre-wrap;
  word-wrap: break-word;
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