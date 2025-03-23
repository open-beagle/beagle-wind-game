<template>
  <div class="home">
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card" @click="handleCardClick('/nodes')">
          <template #header>
            <div class="card-header">
              <span>游戏节点</span>
              <el-tag type="success">{{ stats.nodes.total }}</el-tag>
            </div>
          </template>
          <div class="stat-content">
            <div class="stat-item">
              <span>在线节点</span>
              <span class="value">{{ stats.nodes.online }}</span>
            </div>
            <div class="stat-item">
              <span>离线节点</span>
              <span class="value">{{ stats.nodes.offline }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card" @click="handleCardClick('/platforms')">
          <template #header>
            <div class="card-header">
              <span>游戏平台</span>
              <el-tag type="primary">{{ stats.platforms.total }}</el-tag>
            </div>
          </template>
          <div class="stat-content">
            <div class="stat-item">
              <span>已配置</span>
              <span class="value">{{ stats.platforms.configured }}</span>
            </div>
            <div class="stat-item">
              <span>未配置</span>
              <span class="value">{{ stats.platforms.unconfigured }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card" @click="handleCardClick('/cards')">
          <template #header>
            <div class="card-header">
              <span>游戏卡片</span>
              <el-tag type="warning">{{ stats.cards.total }}</el-tag>
            </div>
          </template>
          <div class="stat-content">
            <div class="stat-item">
              <span>活跃卡片</span>
              <span class="value">{{ stats.cards.active }}</span>
            </div>
            <div class="stat-item">
              <span>未活跃卡片</span>
              <span class="value">{{ stats.cards.inactive }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card" @click="handleCardClick('/instances')">
          <template #header>
            <div class="card-header">
              <span>游戏实例</span>
              <el-tag type="info">{{ stats.instances.total }}</el-tag>
            </div>
          </template>
          <div class="stat-content">
            <div class="stat-item">
              <span>运行中</span>
              <span class="value">{{ stats.instances.running }}</span>
            </div>
            <div class="stat-item">
              <span>已停止</span>
              <span class="value">{{ stats.instances.stopped }}</span>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    
    <el-row :gutter="20" class="mt-20">
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>资源使用情况</span>
            </div>
          </template>
          <div class="resource-stats">
            <div class="resource-item">
              <span>CPU 使用率</span>
              <el-progress :percentage="stats.resources.cpu" />
            </div>
            <div class="resource-item">
              <span>内存使用率</span>
              <el-progress :percentage="stats.resources.memory" />
            </div>
            <div class="resource-item">
              <span>磁盘使用率</span>
              <el-progress :percentage="stats.resources.disk" />
            </div>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card shadow="hover">
          <template #header>
            <div class="card-header">
              <span>最近活动</span>
            </div>
          </template>
          <el-timeline>
            <el-timeline-item
              v-for="(activity, index) in stats.recentActivities"
              :key="index"
              :timestamp="activity.time"
              :type="activity.type"
            >
              {{ activity.content }}
            </el-timeline-item>
          </el-timeline>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const handleCardClick = (path: string) => {
  router.push(path)
}

const stats = ref({
  nodes: {
    total: 3,
    online: 3,
    offline: 0
  },
  platforms: {
    total: 3,
    configured: 3,
    unconfigured: 0
  },
  cards: {
    total: 3,
    active: 2,
    inactive: 1
  },
  instances: {
    total: 3,
    running: 2,
    stopped: 1
  },
  resources: {
    cpu: 5,
    memory: 12,
    disk: 25
  },
  recentActivities: [
    {
      content: '节点 游戏机-241 资源使用率: CPU 5%, 内存 12%',
      time: '2024-03-21 18:00:00',
      type: 'success'
    },
    {
      content: '配置平台 Lutris v0.5.18 完成',
      time: '2024-03-21 17:30:00',
      type: 'primary'
    },
    {
      content: '节点 游戏机-243 资源使用率: CPU 3%, 内存 10%',
      time: '2024-03-21 17:00:00',
      type: 'success'
    }
  ]
})
</script>

<style scoped>
.home {
  padding: 24px;
  min-height: calc(100vh - 64px);
}

.mt-20 {
  margin-top: 24px;
}

.stat-card {
  height: 180px;
  border-radius: 8px;
  transition: all 0.3s;
  border: none;
  background: #fff;
  cursor: pointer;
}

.stat-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 8px 24px rgba(0, 21, 41, 0.12);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  border-bottom: 1px solid #f0f0f0;
  background: #fff;
  border-radius: 8px 8px 0 0;
}

.card-header span {
  font-size: 16px;
  font-weight: 600;
  color: #262626;
}

.stat-content {
  padding: 24px;
}

.stat-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 20px;
  color: #595959;
  font-size: 14px;
}

.stat-item span:first-child {
  min-width: 80px;
}

.stat-item .value {
  font-weight: 600;
  color: #1890ff;
  font-size: 16px;
  text-align: right;
  min-width: 40px;
}

.resource-stats {
  padding: 24px;
}

.resource-item {
  margin-bottom: 32px;
}

.resource-item:last-child {
  margin-bottom: 0;
}

.resource-item span {
  display: block;
  margin-bottom: 12px;
  color: #595959;
  font-size: 14px;
}

:deep(.el-card) {
  border-radius: 8px;
  border: none;
  box-shadow: 0 2px 12px rgba(0, 21, 41, 0.08);
}

:deep(.el-card__header) {
  padding: 0;
  border-bottom: none;
}

:deep(.el-progress-bar__outer) {
  background-color: #f5f5f5;
  border-radius: 4px;
  height: 8px;
}

:deep(.el-progress-bar__inner) {
  border-radius: 4px;
  background-color: #1890ff;
  transition: width 0.6s ease;
}

:deep(.el-timeline-item__node) {
  background-color: #1890ff;
  width: 10px;
  height: 10px;
}

:deep(.el-timeline-item__tail) {
  border-left-color: #f0f0f0;
  left: 4px;
}

:deep(.el-timeline-item__timestamp) {
  color: #8c8c8c;
  font-size: 13px;
}

:deep(.el-timeline-item__content) {
  color: #262626;
  font-size: 14px;
  line-height: 1.6;
}

:deep(.el-tag) {
  border-radius: 4px;
  padding: 0 12px;
  height: 24px;
  line-height: 24px;
  font-size: 13px;
  font-weight: 500;
}

:deep(.el-tag--success) {
  background-color: #f6ffed;
  border-color: #b7eb8f;
  color: #52c41a;
}

:deep(.el-tag--primary) {
  background-color: #e6f7ff;
  border-color: #91d5ff;
  color: #1890ff;
}

:deep(.el-tag--warning) {
  background-color: #fff7e6;
  border-color: #ffd591;
  color: #fa8c16;
}

:deep(.el-tag--info) {
  background-color: #f5f5f5;
  border-color: #d9d9d9;
  color: #595959;
}
</style> 