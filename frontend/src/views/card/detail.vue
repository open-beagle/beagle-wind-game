<template>
  <div class="game-card-detail">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-button link @click="router.back()">
              <el-icon><ArrowLeft /></el-icon>
              返回
            </el-button>
            <span class="title">游戏详情</span>
          </div>
          <div class="header-right">
            <el-button type="primary" @click="handleEdit">编辑</el-button>
            <el-button type="success" @click="handleCreateInstance">创建实例</el-button>
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <div class="detail-content">
        <div class="basic-info">
          <div class="cover-image">
            <el-image
              :src="gameCard?.coverImage"
              fit="cover"
              :preview-src-list="[gameCard?.coverImage]"
            >
              <template #error>
                <div class="image-placeholder">
                  <el-icon><Picture /></el-icon>
                </div>
              </template>
            </el-image>
          </div>
          <div class="info-content">
            <h2>{{ gameCard?.name }}</h2>
            <div class="info-item">
              <span class="label">游戏ID：</span>
              <span class="value">{{ gameCard?.id }}</span>
            </div>
            <div class="info-item">
              <span class="label">游戏平台：</span>
              <el-tag :type="getPlatformType(gameCard?.platform?.type)">
                {{ gameCard?.platform?.name }}
              </el-tag>
            </div>
            <div class="info-item">
              <span class="label">游戏状态：</span>
              <el-tag :type="getStatusType(gameCard?.status)">
                {{ getStatusText(gameCard?.status) }}
              </el-tag>
            </div>
            <div class="info-item">
              <span class="label">创建时间：</span>
              <span class="value">{{ gameCard?.createdAt }}</span>
            </div>
          </div>
        </div>

        <div class="description-section">
          <h3>游戏描述</h3>
          <p>{{ gameCard?.description }}</p>
        </div>

        <div class="instances-section">
          <div class="section-header">
            <h3>游戏实例</h3>
            <el-button type="primary" @click="handleCreateInstance">创建实例</el-button>
          </div>
          <el-table :data="instances" style="width: 100%">
            <el-table-column prop="id" label="实例ID" width="120" />
            <el-table-column prop="node.name" label="运行节点" width="150" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getInstanceStatusType(row.status)">
                  {{ getInstanceStatusText(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="createdAt" label="创建时间" width="180" />
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button link type="primary" @click="handleViewInstance(row)">
                  查看
                </el-button>
                <el-button link type="success" @click="handleStartInstance(row)">
                  启动
                </el-button>
                <el-button link type="warning" @click="handleStopInstance(row)">
                  停止
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ArrowLeft, Picture } from '@element-plus/icons-vue'
import { mockGameCards, mockGameInstances } from '@/mocks'
import type { GameCard, GameInstance } from '@/types'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const gameCard = ref<GameCard>()
const instances = ref<GameInstance[]>([])

// 获取游戏卡片详情
const getGameCardDetail = async () => {
  loading.value = true
  try {
    // 模拟 API 请求延迟
    await new Promise(resolve => setTimeout(resolve, 300))
    
    const id = route.params.id as string
    const card = mockGameCards.find(c => c.id === id)
    if (!card) {
      ElMessage.error('游戏不存在')
      router.push('/card')
      return
    }
    
    gameCard.value = card
    // 获取该游戏的实例列表
    instances.value = mockGameInstances.filter(i => i.gameId === id)
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 平台类型
const platformTypeMap: Record<string, string> = {
  steam: "primary",
  epic: "success",
  gog: "warning",
  xbox: "info",
  psn: "danger",
  nintendo: "success",
}

const getPlatformType = (type?: string) => platformTypeMap[type || ''] || "info"

// 状态类型
const statusTypeMap: Record<string, string> = {
  active: "success",
  inactive: "info",
  maintenance: "warning",
}

const getStatusType = (status?: string) => statusTypeMap[status || ''] || "info"

// 状态文本
const statusTextMap: Record<string, string> = {
  active: "正常",
  inactive: "停用",
  maintenance: "维护中",
}

const getStatusText = (status?: string) => statusTextMap[status || ''] || status

// 实例状态类型
const instanceStatusTypeMap: Record<string, string> = {
  running: "success",
  stopped: "info",
  error: "danger",
  starting: "warning",
  stopping: "warning"
}

const getInstanceStatusType = (status: string) => instanceStatusTypeMap[status] || "info"

// 实例状态文本
const instanceStatusTextMap: Record<string, string> = {
  running: "运行中",
  stopped: "已停止",
  error: "错误",
  starting: "启动中",
  stopping: "停止中"
}

const getInstanceStatusText = (status: string) => instanceStatusTextMap[status] || status

// 编辑游戏
const handleEdit = () => {
  router.push(`/card/edit/${gameCard.value?.id}`)
}

// 创建实例
const handleCreateInstance = () => {
  ElMessage.info('创建实例功能开发中')
}

// 删除游戏
const handleDelete = () => {
  ElMessageBox.confirm('确定要删除该游戏吗？', '提示', {
    type: 'warning'
  }).then(() => {
    ElMessage.success('删除成功')
    router.push('/card')
  })
}

// 查看实例
const handleViewInstance = (instance: GameInstance) => {
  router.push(`/instance/detail/${instance.id}`)
}

// 启动实例
const handleStartInstance = (instance: GameInstance) => {
  ElMessage.info('启动实例功能开发中')
}

// 停止实例
const handleStopInstance = (instance: GameInstance) => {
  ElMessage.info('停止实例功能开发中')
}

onMounted(() => {
  getGameCardDetail()
})
</script>

<style scoped>
.game-card-detail {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 20px;
}

.title {
  font-size: 18px;
  font-weight: bold;
}

.header-right {
  display: flex;
  gap: 10px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.basic-info {
  display: flex;
  gap: 30px;
}

.cover-image {
  width: 200px;
  height: 200px;
  border-radius: 8px;
  overflow: hidden;
}

.image-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  background-color: #f5f7fa;
  color: #909399;
  font-size: 30px;
}

.info-content {
  flex: 1;
}

.info-content h2 {
  margin: 0 0 20px 0;
  font-size: 24px;
}

.info-item {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.label {
  color: #606266;
  width: 100px;
}

.description-section {
  background-color: #f5f7fa;
  padding: 20px;
  border-radius: 8px;
}

.description-section h3 {
  margin: 0 0 15px 0;
  font-size: 18px;
}

.description-section p {
  margin: 0;
  line-height: 1.6;
  color: #606266;
}

.instances-section {
  background-color: #fff;
  border-radius: 8px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.section-header h3 {
  margin: 0;
  font-size: 18px;
}
</style> 