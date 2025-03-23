<template>
  <div class="platform-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>平台详情</span>
          <div class="header-actions">
            <el-button type="primary" @click="handleEdit">编辑</el-button>
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <div v-if="platform" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="平台ID">{{ platform.id }}</el-descriptions-item>
          <el-descriptions-item label="平台名称">{{ platform.name }}</el-descriptions-item>
          <el-descriptions-item label="操作系统">
            <el-tag type="info">{{ platform.os }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(platform.status)">
              {{ getStatusText(platform.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="版本">{{ platform.version }}</el-descriptions-item>
          <el-descriptions-item label="描述">{{ platform.description }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ platform.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ platform.updatedAt }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">运行环境</div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Docker镜像">{{ platform.image }}</el-descriptions-item>
          <el-descriptions-item label="启动路径">{{ platform.bin }}</el-descriptions-item>
          <el-descriptions-item label="数据目录">{{ platform.data }}</el-descriptions-item>
          <el-descriptions-item label="资源文件">
            <el-tag v-for="file in platform.files" :key="file" class="file-tag">
              {{ file }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">平台特性</div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="支持的游戏类型">
            <el-tag v-for="type in platform.features.gameTypes" :key="type" class="feature-tag">
              {{ type }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="支持的平台">
            <el-tag v-for="platform in platform.features.platforms" :key="platform" class="feature-tag">
              {{ platform }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">配置信息</div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="API密钥">
            <el-input v-model="platform.config.apiKey" readonly>
              <template #append>
                <el-button @click="handleCopyApiKey">复制</el-button>
              </template>
            </el-input>
          </el-descriptions-item>
          <el-descriptions-item label="API地址">{{ platform.config.apiUrl }}</el-descriptions-item>
          <el-descriptions-item label="回调地址">{{ platform.config.callbackUrl }}</el-descriptions-item>
          <el-descriptions-item label="环境变量">
            <el-tag v-for="(value, key) in platform.config.env" :key="key" class="env-tag">
              {{ key }}: {{ value }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="挂载目录">
            <el-tag v-for="(value, key) in platform.config.volumes" :key="key" class="env-tag">
              {{ key }}: {{ value }}
            </el-tag>
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-card>

    <!-- 编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      title="编辑平台"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="平台名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入平台名称" />
        </el-form-item>
        <el-form-item label="操作系统" prop="os">
          <el-select v-model="form.os" placeholder="请选择操作系统">
            <el-option label="Linux" value="Linux" />
            <el-option label="Windows" value="Windows" />
            <el-option label="macOS" value="macOS" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择状态">
            <el-option label="正常" value="active" />
            <el-option label="维护中" value="maintenance" />
            <el-option label="停用" value="inactive" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="form.version" placeholder="请输入版本号" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="Docker镜像" prop="image">
          <el-input v-model="form.image" placeholder="请输入Docker镜像" />
        </el-form-item>
        <el-form-item label="启动路径" prop="bin">
          <el-input v-model="form.bin" placeholder="请输入启动路径" />
        </el-form-item>
        <el-form-item label="数据目录" prop="data">
          <el-input v-model="form.data" placeholder="请输入数据目录" />
        </el-form-item>
        <el-form-item label="API密钥" prop="config.apiKey">
          <el-input v-model="form.config.apiKey" placeholder="请输入API密钥" />
        </el-form-item>
        <el-form-item label="API地址" prop="config.apiUrl">
          <el-input v-model="form.config.apiUrl" placeholder="请输入API地址" />
        </el-form-item>
        <el-form-item label="回调地址" prop="config.callbackUrl">
          <el-input v-model="form.config.callbackUrl" placeholder="请输入回调地址" />
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
import { mockGamePlatforms } from '@/mocks'
import type { GamePlatform } from '@/types'
import type { FormInstance, FormRules } from 'element-plus'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const platform = ref<GamePlatform | null>(null)
const dialogVisible = ref(false)
const formRef = ref<FormInstance>()

const form = ref({
  id: '',
  name: '',
  os: 'Linux',
  status: '',
  version: '',
  description: '',
  image: '',
  bin: '',
  data: '',
  files: [],
  features: {
    gameTypes: [],
    platforms: []
  },
  config: {
    apiKey: '',
    apiUrl: '',
    callbackUrl: '',
    env: {},
    volumes: {}
  }
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入平台名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  os: [
    { required: true, message: '请选择操作系统', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ],
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入描述', trigger: 'blur' }
  ],
  image: [
    { required: true, message: '请输入Docker镜像', trigger: 'blur' }
  ],
  bin: [
    { required: true, message: '请输入启动路径', trigger: 'blur' }
  ],
  data: [
    { required: true, message: '请输入数据目录', trigger: 'blur' }
  ],
  'config.apiKey': [
    { required: true, message: '请输入API密钥', trigger: 'blur' }
  ],
  'config.apiUrl': [
    { required: true, message: '请输入API地址', trigger: 'blur' }
  ],
  'config.callbackUrl': [
    { required: true, message: '请输入回调地址', trigger: 'blur' }
  ]
}

// 获取平台详情
const getPlatformDetail = async () => {
  loading.value = true
  try {
    // 模拟API请求
    await new Promise(resolve => setTimeout(resolve, 300))
    const id = route.params.id as string
    platform.value = mockGamePlatforms.find(p => p.id === id) || null
  } catch (error) {
    ElMessage.error('加载数据失败')
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 状态类型
const statusTypeMap: Record<string, string> = {
  active: 'success',
  maintenance: 'warning',
  inactive: 'info'
}

const getStatusType = (status: string) => statusTypeMap[status] || 'info'

// 状态文本
const statusTextMap: Record<string, string> = {
  active: '正常',
  maintenance: '维护中',
  inactive: '停用'
}

const getStatusText = (status: string) => statusTextMap[status] || status

// 编辑平台
const handleEdit = () => {
  if (!platform.value) return
  form.value = {
    id: platform.value.id,
    name: platform.value.name,
    os: platform.value.os,
    status: platform.value.status,
    version: platform.value.version,
    description: platform.value.description,
    image: platform.value.image,
    bin: platform.value.bin,
    data: platform.value.data,
    files: [...platform.value.files],
    features: {
      gameTypes: [...platform.value.features.gameTypes],
      platforms: [...platform.value.features.platforms]
    },
    config: {
      apiKey: platform.value.config.apiKey,
      apiUrl: platform.value.config.apiUrl,
      callbackUrl: platform.value.config.callbackUrl,
      env: { ...platform.value.config.env },
      volumes: { ...platform.value.config.volumes }
    }
  }
  dialogVisible.value = true
}

// 删除平台
const handleDelete = () => {
  ElMessageBox.confirm('确定要删除该平台吗？', '提示', {
    type: 'warning'
  }).then(() => {
    // 模拟删除
    const index = mockGamePlatforms.findIndex(p => p.id === platform.value?.id)
    if (index > -1) {
      mockGamePlatforms.splice(index, 1)
      ElMessage.success('删除成功')
      router.push('/platform')
    }
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate((valid) => {
    if (valid) {
      // 模拟更新
      const index = mockGamePlatforms.findIndex(p => p.id === form.value.id)
      if (index > -1) {
        mockGamePlatforms[index] = {
          ...mockGamePlatforms[index],
          name: form.value.name,
          os: form.value.os,
          status: form.value.status,
          version: form.value.version,
          description: form.value.description,
          image: form.value.image,
          bin: form.value.bin,
          data: form.value.data,
          files: [...form.value.files],
          features: {
            gameTypes: [...form.value.features.gameTypes],
            platforms: [...form.value.features.platforms]
          },
          config: {
            apiKey: form.value.config.apiKey,
            apiUrl: form.value.config.apiUrl,
            callbackUrl: form.value.config.callbackUrl,
            env: { ...form.value.config.env },
            volumes: { ...form.value.config.volumes }
          }
        }
        ElMessage.success('更新成功')
        dialogVisible.value = false
        getPlatformDetail()
      }
    }
  })
}

// 复制API密钥
const handleCopyApiKey = () => {
  if (!platform.value) return
  navigator.clipboard.writeText(platform.value.config.apiKey)
    .then(() => {
      ElMessage.success('复制成功')
    })
    .catch(() => {
      ElMessage.error('复制失败')
    })
}

onMounted(() => {
  getPlatformDetail()
})
</script>

<style scoped>
.platform-detail-container {
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

.env-tag,
.feature-tag,
.file-tag {
  margin-right: 8px;
  margin-bottom: 8px;
}

:deep(.el-descriptions) {
  margin-bottom: 20px;
}

:deep(.el-descriptions__label) {
  width: 120px;
}
</style> 