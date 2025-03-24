<template>
  <div class="api-error-state">
    <el-empty :image-size="200">
      <template #description>
        <div class="error-content">
          <h3 class="error-title">{{ title || '数据加载失败' }}</h3>
          <p class="error-message">{{ message || '请稍后重试' }}</p>
          <p class="error-detail" v-if="detail">{{ detail }}</p>
          <div class="error-actions">
            <slot name="actions">
              <el-button type="primary" @click="$emit('retry')" :loading="loading">
                重试
              </el-button>
              <el-button @click="$emit('back')">返回</el-button>
            </slot>
          </div>
        </div>
      </template>
    </el-empty>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title?: string
  message?: string
  detail?: string
  loading?: boolean
}>()

defineEmits<{
  (e: 'retry'): void
  (e: 'back'): void
}>()
</script>

<style scoped>
.api-error-state {
  padding: 40px 0;
  text-align: center;
}

.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.error-title {
  color: #303133;
  font-size: 18px;
  margin: 0;
}

.error-message {
  color: #606266;
  font-size: 14px;
  margin: 0;
}

.error-detail {
  color: #909399;
  font-size: 13px;
  margin: 0;
}

.error-actions {
  margin-top: 16px;
}
</style> 