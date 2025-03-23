<template>
  <el-container class="layout-container">
    <el-aside :width="isCollapse ? '64px' : '200px'" class="aside">
      <div class="logo">
        <img src="../assets/logo.svg" alt="Beagle Wind Game" />
        <span v-show="!isCollapse">Beagle Wind Game</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="menu"
        :collapse="isCollapse"
        router
      >
        <el-menu-item index="/">
          <el-icon><Monitor /></el-icon>
          <template #title>首页</template>
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
          <el-icon><Grid /></el-icon>
          <template #title>游戏卡片</template>
        </el-menu-item>
        <el-menu-item index="/instances">
          <el-icon><Box /></el-icon>
          <template #title>游戏实例</template>
        </el-menu-item>
      </el-menu>
    </el-aside>
    
    <el-container>
      <el-header class="header">
        <div class="header-left">
          <el-button
            type="text"
            :icon="isCollapse ? 'Expand' : 'Fold'"
            @click="toggleCollapse"
          />
        </div>
        <div class="header-right">
          <el-dropdown>
            <span class="user-info">
              <el-avatar :size="32" src="https://cube.elemecdn.com/0/88/03b0d39583f48206768a7534e55bcpng.png" />
              <span>管理员</span>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item>个人信息</el-dropdown-item>
                <el-dropdown-item>修改密码</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      
      <el-main class="main">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import {
  Monitor,
  Connection,
  Platform,
  Grid,
  Box,
  Expand,
  Fold
} from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const isCollapse = ref(false)
const activeMenu = computed(() => route.path)

const toggleCollapse = () => {
  isCollapse.value = !isCollapse.value
}

const handleLogout = () => {
  localStorage.removeItem('isAuthenticated')
  router.push('/login')
}
</script>

<style scoped>
.layout-container {
  height: 100vh;
  background-color: #f0f2f5;
}

.aside {
  background-color: #001529;
  transition: all 0.3s;
  overflow: hidden;
  box-shadow: 2px 0 8px rgba(0, 21, 41, 0.15);
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  padding: 0 24px;
  color: #fff;
  font-size: 18px;
  font-weight: 600;
  background-color: #002140;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.logo img {
  width: 36px;
  height: 36px;
  margin-right: 12px;
  filter: drop-shadow(0 2px 4px rgba(0, 21, 41, 0.1));
}

.menu {
  border-right: none;
  background-color: #001529;
  padding: 8px 0;
}

.menu:not(.el-menu--collapse) {
  width: 200px;
}

.header {
  background-color: #fff;
  border-bottom: 1px solid #f0f0f0;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 64px;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
}

.header-left {
  display: flex;
  align-items: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-info {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 0 12px;
  height: 40px;
  border-radius: 4px;
  transition: all 0.3s;
}

.user-info:hover {
  background-color: #f5f5f5;
}

.user-info span {
  margin-left: 8px;
  color: #262626;
  font-size: 14px;
}

.main {
  background-color: #f0f2f5;
  padding: 24px;
  overflow-y: auto;
}

:deep(.el-menu) {
  border-right: none;
}

:deep(.el-menu-item) {
  height: 48px;
  line-height: 48px;
  margin: 4px 0;
  color: rgba(255, 255, 255, 0.65);
  border-radius: 4px;
  margin: 4px 8px;
}

:deep(.el-menu-item:hover) {
  color: #fff;
  background-color: #1890ff;
}

:deep(.el-menu-item.is-active) {
  color: #fff;
  background-color: #1890ff;
}

:deep(.el-menu-item .el-icon) {
  color: rgba(255, 255, 255, 0.65);
  font-size: 18px;
  margin-right: 12px;
}

:deep(.el-menu-item:hover .el-icon),
:deep(.el-menu-item.is-active .el-icon) {
  color: #fff;
}

:deep(.el-menu--collapse .el-menu-item) {
  margin: 4px;
  padding: 0 20px;
}

:deep(.el-menu--collapse .el-menu-item .el-icon) {
  margin-right: 0;
}
</style> 