<template>
  <el-menu
    :default-active="activeMenu"
    class="main-menu"
    :router="true"
    :collapse="isCollapse"
  >
    <el-menu-item index="/">
      <el-icon><Monitor /></el-icon>
      <template #title>仪表盘</template>
    </el-menu-item>
    
    <el-menu-item index="/nodes">
      <el-icon><Connection /></el-icon>
      <template #title>游戏节点</template>
    </el-menu-item>
    
    <el-menu-item index="/platforms">
      <el-icon><Platform /></el-icon>
      <template #title>游戏平台</template>
    </el-menu-item>
    
    <el-menu-item index="/cards">
      <el-icon><Collection /></el-icon>
      <template #title>游戏卡片</template>
    </el-menu-item>
    
    <el-menu-item index="/instances">
      <el-icon><Box /></el-icon>
      <template #title>游戏实例</template>
    </el-menu-item>
  </el-menu>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import { Monitor, Connection, Platform, Collection, Box } from '@element-plus/icons-vue'

const route = useRoute()
const isCollapse = ref(false)

const activeMenu = computed(() => {
  return route.path
})
</script>

<style scoped>
.main-menu {
  height: 100%;
  border-right: none;
}

.main-menu:not(.el-menu--collapse) {
  width: 200px;
}

.el-menu-item {
  display: flex;
  align-items: center;
}

.el-icon {
  margin-right: 8px;
}
</style> 