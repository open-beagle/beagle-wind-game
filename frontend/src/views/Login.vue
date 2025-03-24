<template>
  <div class="login-container">
    <el-card class="login-card">
      <div class="logo">
        <img src="../assets/logo.png" alt="Beagle Wind Game" />
      </div>
      <h2>Beagle Wind Game</h2>
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="0"
        class="login-form"
      >
        <el-form-item prop="username">
          <el-input
            v-model="form.username"
            placeholder="用户名"
            prefix-icon="User"
          />
        </el-form-item>
        <el-form-item prop="password">
          <el-input
            v-model="form.password"
            type="password"
            placeholder="密码"
            prefix-icon="Lock"
            show-password
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" class="login-button" @click="handleLogin">
            登录
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { User, Lock } from '@element-plus/icons-vue'

const router = useRouter()
const formRef = ref<FormInstance>()

const form = reactive({
  username: '',
  password: ''
})

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

const handleLogin = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate((valid) => {
    if (valid) {
      // 模拟登录成功
      localStorage.setItem('isAuthenticated', 'true')
      ElMessage.success('登录成功')
      router.push('/')
    }
  })
}
</script>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  justify-content: center;
  align-items: center;
  background: linear-gradient(135deg, #001529 0%, #003366 100%);
  position: relative;
  overflow: hidden;
}

.login-container::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: url('../assets/pattern.svg') repeat;
  opacity: 0.1;
}

.login-card {
  width: 420px;
  padding: 48px;
  text-align: center;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 21, 41, 0.15);
  position: relative;
  z-index: 1;
  backdrop-filter: blur(10px);
}

.logo {
  margin-bottom: 24px;
  display: flex;
  justify-content: center;
}

.logo img {
  width: 250px;
  height: 250px;
  filter: contrast(1.2) brightness(1.1) drop-shadow(0 8px 16px rgba(0, 21, 41, 0.3));
  object-fit: contain;
  background-color: #ffffff;
  border-radius: 12px;
  padding: 24px;
  border: 2px solid #e6f7ff;
  box-shadow: 0 0 20px rgba(24, 144, 255, 0.25);
}

h2 {
  margin: 0 0 48px;
  color: #262626;
  font-size: 32px;
  font-weight: 600;
  letter-spacing: 0.5px;
}

.login-form {
  margin-top: 32px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 24px;
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  background-color: #1890ff;
  border-color: #1890ff;
  border-radius: 4px;
  transition: all 0.3s;
}

.login-button:hover {
  background-color: #40a9ff;
  border-color: #40a9ff;
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(24, 144, 255, 0.2);
}

:deep(.el-input__wrapper) {
  box-shadow: 0 0 0 1px #d9d9d9 inset;
  transition: all 0.3s;
  padding: 0 16px;
  height: 44px;
}

:deep(.el-input__wrapper:hover) {
  box-shadow: 0 0 0 1px #40a9ff inset;
}

:deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #1890ff inset;
}

:deep(.el-input__inner) {
  height: 44px;
  font-size: 14px;
}

:deep(.el-input__prefix) {
  font-size: 18px;
  color: #8c8c8c;
}

:deep(.el-input__prefix-inner) {
  margin-right: 8px;
}
</style> 