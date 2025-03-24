<template>
  <div class="platform-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>平台详情</span>
          <div class="header-actions">
            <el-button v-if="!isEditing" type="primary" @click="handleEdit">编辑</el-button>
            <el-button v-if="isEditing" type="success" @click="handleSave">保存</el-button>
            <el-button v-if="isEditing" type="info" @click="handleCancel">取消</el-button>
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <div v-if="platform" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="平台ID">{{ platform.id }}</el-descriptions-item>
          <el-descriptions-item label="平台名称">
            <template v-if="isEditing">
              <el-input v-model="editForm.name" />
            </template>
            <template v-else>
              {{ platform.name }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="操作系统">
            <template v-if="isEditing">
              <el-select v-model="editForm.os">
                <el-option label="Linux" value="Linux" />
                <el-option label="Windows" value="Windows" />
                <el-option label="macOS" value="macOS" />
              </el-select>
            </template>
            <template v-else>
              <el-tag type="info">{{ platform.os }}</el-tag>
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <template v-if="isEditing">
              <el-select v-model="editForm.status">
                <el-option label="正常" value="active" />
                <el-option label="维护中" value="maintenance" />
                <el-option label="停用" value="inactive" />
              </el-select>
            </template>
            <template v-else>
              <el-tag :type="getStatusType(platform.status)">
                {{ getStatusText(platform.status) }}
              </el-tag>
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="版本">
            <template v-if="isEditing">
              <el-input v-model="editForm.version" />
            </template>
            <template v-else>
              {{ platform.version }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="描述">
            <template v-if="isEditing">
              <el-input v-model="editForm.description" />
            </template>
            <template v-else>
              {{ platform.description }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ platform.createdAt }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{ platform.updatedAt }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">运行环境</div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Docker镜像">
            <template v-if="isEditing">
              <el-input v-model="editForm.image" />
            </template>
            <template v-else>
              {{ platform.image }}
            </template>
          </el-descriptions-item>
        </el-descriptions>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="启动路径">
            <template v-if="isEditing">
              <el-input v-model="editForm.bin" />
            </template>
            <template v-else>
              {{ platform.bin }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="资源文件">
            <template v-if="isEditing">
              <div v-for="(file, index) in editForm.files" :key="file.id" class="file-item">
                <el-input v-model="file.type" placeholder="文件类型" style="width: 150px" />
                <el-input v-model="file.url" placeholder="文件URL" />
                <el-button type="danger" @click="removeFile(index)">删除</el-button>
              </div>
              <el-button type="primary" @click="addFile">添加文件</el-button>
            </template>
            <template v-else>
              <div v-for="file in platform.files" :key="file.id" class="file-item">
                <el-tag>{{ file.type }}</el-tag>
                <span class="file-url">{{ file.url }}</span>
              </div>
            </template>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">平台特性</div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="特性列表">
            <template v-if="isEditing">
              <div v-for="(feature, index) in editForm.features" :key="index" class="feature-item">
                <el-input v-model="editForm.features[index]" />
                <el-button type="danger" @click="removeFeature(index)">删除</el-button>
              </div>
              <el-button type="primary" @click="addFeature">添加特性</el-button>
            </template>
            <template v-else>
              <div v-for="(feature, index) in platform.features" :key="index" class="feature-item">
                {{ feature }}
              </div>
            </template>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">配置信息</div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="Wine版本" v-if="platform.config.wine">
            <template v-if="isEditing">
              <el-input v-model="editForm.config.wine" />
            </template>
            <template v-else>
              {{ platform.config.wine }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="DXVK版本" v-if="platform.config.dxvk">
            <template v-if="isEditing">
              <el-input v-model="editForm.config.dxvk" />
            </template>
            <template v-else>
              {{ platform.config.dxvk }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="VKD3D版本" v-if="platform.config.vkd3d">
            <template v-if="isEditing">
              <el-input v-model="editForm.config.vkd3d" />
            </template>
            <template v-else>
              {{ platform.config.vkd3d }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="Python版本" v-if="platform.config.python">
            <template v-if="isEditing">
              <el-input v-model="editForm.config.python" />
            </template>
            <template v-else>
              {{ platform.config.python }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="Proton版本" v-if="platform.config.proton">
            <template v-if="isEditing">
              <el-input v-model="editForm.config.proton" />
            </template>
            <template v-else>
              {{ platform.config.proton }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="着色器缓存" v-if="platform.config['shader-cache']">
            <template v-if="isEditing">
              <el-select v-model="editForm.config['shader-cache']">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config['shader-cache'] }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="远程游戏" v-if="platform.config['remote-play']">
            <template v-if="isEditing">
              <el-select v-model="editForm.config['remote-play']">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config['remote-play'] }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="广播" v-if="platform.config.broadcast">
            <template v-if="isEditing">
              <el-select v-model="editForm.config.broadcast">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.broadcast }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="运行模式" v-if="platform.config.mode">
            <template v-if="isEditing">
              <el-select v-model="editForm.config.mode">
                <el-option label="主机模式" value="docked" />
                <el-option label="掌机模式" value="handheld" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.mode }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="分辨率" v-if="platform.config.resolution">
            <template v-if="isEditing">
              <el-select v-model="editForm.config.resolution">
                <el-option label="1080p" value="1080p" />
                <el-option label="720p" value="720p" />
                <el-option label="480p" value="480p" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.resolution }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="WiFi" v-if="platform.config.wifi">
            <template v-if="isEditing">
              <el-select v-model="editForm.config.wifi">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.wifi }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="蓝牙" v-if="platform.config.bluetooth">
            <template v-if="isEditing">
              <el-select v-model="editForm.config.bluetooth">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.bluetooth }}
            </template>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">安装配置</div>
        <div v-if="platform.installer">
          <el-descriptions :column="1" border>
            <el-descriptions-item label="安装命令">
              <template v-if="isEditing">
                <div v-for="(command, index) in editForm.installer" :key="index" class="installer-item">
                  <el-select v-model="command.type" placeholder="选择命令类型">
                    <el-option label="命令" value="command" />
                    <el-option label="移动" value="move" />
                    <el-option label="权限" value="chmodx" />
                    <el-option label="解压" value="extract" />
                  </el-select>
                  <template v-if="command.type === 'command'">
                    <el-input v-model="command.command" placeholder="输入命令" />
                  </template>
                  <template v-if="command.type === 'move'">
                    <el-input v-model="command.move.src" placeholder="源文件" />
                    <el-input v-model="command.move.dst" placeholder="目标路径" />
                  </template>
                  <template v-if="command.type === 'chmodx'">
                    <el-input v-model="command.chmodx" placeholder="文件路径" />
                  </template>
                  <template v-if="command.type === 'extract'">
                    <el-input v-model="command.extract.file" placeholder="压缩文件" />
                    <el-input v-model="command.extract.dst" placeholder="解压目标" />
                  </template>
                  <el-button type="danger" @click="removeInstallerCommand(index)">删除</el-button>
                </div>
                <el-button type="primary" @click="addInstallerCommand">添加命令</el-button>
              </template>
              <template v-else>
                <div v-for="(command, index) in platform.installer" :key="index" class="installer-item">
                  <template v-if="command.command">
                    <el-tag type="info">命令</el-tag>
                    <span class="command-text">{{ command.command }}</span>
                  </template>
                  <template v-if="command.move">
                    <el-tag type="success">移动</el-tag>
                    <span class="command-text">从 {{ command.move.src }} 到 {{ command.move.dst }}</span>
                  </template>
                  <template v-if="command.chmodx">
                    <el-tag type="warning">权限</el-tag>
                    <span class="command-text">设置 {{ command.chmodx }} 可执行权限</span>
                  </template>
                  <template v-if="command.extract">
                    <el-tag type="danger">解压</el-tag>
                    <span class="command-text">解压 {{ command.extract.file }} 到 {{ command.extract.dst }}</span>
                  </template>
                </div>
              </template>
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { mockGamePlatforms } from '@/mocks'
import type { GamePlatform, GamePlatformFile, GamePlatformInstallerCommand } from '@/types'

const route = useRoute()
const router = useRouter()
const loading = ref(false)
const platform = ref<GamePlatform | null>(null)
const isEditing = ref(false)
const editForm = ref<GamePlatform | null>(null)

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
  editForm.value = JSON.parse(JSON.stringify(platform.value))
  isEditing.value = true
}

// 取消编辑
const handleCancel = () => {
  isEditing.value = false
  editForm.value = null
}

// 保存编辑
const handleSave = () => {
  if (!editForm.value || !platform.value) return
  
  // 模拟更新
  const index = mockGamePlatforms.findIndex(p => p.id === platform.value?.id)
  if (index > -1) {
    mockGamePlatforms[index] = {
      ...editForm.value,
      updatedAt: new Date().toISOString()
    }
    platform.value = mockGamePlatforms[index]
    ElMessage.success('更新成功')
    isEditing.value = false
    editForm.value = null
  }
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

// 文件操作
const addFile = () => {
  if (!editForm.value) return
  editForm.value.files.push({
    id: `file-${Date.now()}`,
    type: '',
    url: ''
  })
}

const removeFile = (index: number) => {
  if (!editForm.value) return
  editForm.value.files.splice(index, 1)
}

// 特性操作
const addFeature = () => {
  if (!editForm.value) return
  editForm.value.features.push('')
}

const removeFeature = (index: number) => {
  if (!editForm.value) return
  editForm.value.features.splice(index, 1)
}

// 安装命令操作
const addInstallerCommand = () => {
  if (!editForm.value) return
  editForm.value.installer = editForm.value.installer || []
  editForm.value.installer.push({})
}

const removeInstallerCommand = (index: number) => {
  if (!editForm.value?.installer) return
  editForm.value.installer.splice(index, 1)
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

.file-item,
.feature-item,
.installer-item {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
  gap: 8px;
}

.file-url {
  margin-left: 8px;
  color: #666;
  font-size: 14px;
}

.feature-item {
  margin-bottom: 8px;
  color: #666;
}

.command-text {
  margin-left: 8px;
  color: #666;
  font-family: monospace;
}
</style> 