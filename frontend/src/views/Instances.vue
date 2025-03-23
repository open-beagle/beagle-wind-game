<template>
  <div class="instances-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏实例管理</span>
          <el-button type="primary" @click="handleAdd">创建实例</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="instances"
        style="width: 100%"
        border
      >
        <el-table-column prop="id" label="实例ID" width="120" />
        <el-table-column prop="name" label="实例名称" width="150" />
        <el-table-column prop="card" label="游戏卡片" width="150">
          <template #default="{ row }">
            <el-tag>{{ row.card }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="node" label="运行节点" width="150">
          <template #default="{ row }">
            <el-tag type="info">{{ row.node }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" fixed="right" width="250">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleControl(row)">
                控制
              </el-button>
              <el-button type="success" size="small" @click="handleMonitor(row)">
                监控
              </el-button>
              <el-button type="warning" size="small" @click="handleConfig(row)">
                配置
              </el-button>
              <el-button type="danger" size="small" @click="handleDelete(row)">
                删除
              </el-button>
            </el-button-group>
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 实例创建对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '创建实例' : '编辑实例'"
      width="600px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="实例名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="游戏卡片" prop="card">
          <el-select v-model="form.card" placeholder="请选择游戏卡片">
            <el-option
              v-for="card in cards"
              :key="card.id"
              :label="card.name"
              :value="card.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="运行节点" prop="node">
          <el-select v-model="form.node" placeholder="请选择运行节点">
            <el-option
              v-for="node in nodes"
              :key="node.id"
              :label="node.name"
              :value="node.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="配置" prop="config">
          <el-input
            v-model="form.config"
            type="textarea"
            :rows="5"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">
            确定
          </el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 实例控制对话框 -->
    <el-dialog
      v-model="controlDialogVisible"
      title="实例控制"
      width="400px"
    >
      <div class="control-buttons">
        <el-button-group>
          <el-button type="success" @click="handleStart">启动</el-button>
          <el-button type="warning" @click="handleStop">停止</el-button>
          <el-button type="danger" @click="handleRestart">重启</el-button>
        </el-button-group>
      </div>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { useInstanceStore } from '../stores'
import { useCardStore } from '../stores'
import { useNodeStore } from '../stores'

const instanceStore = useInstanceStore()
const cardStore = useCardStore()
const nodeStore = useNodeStore()

const loading = ref(false)
const dialogVisible = ref(false)
const controlDialogVisible = ref(false)
const dialogType = ref<'add' | 'edit'>('add')
const formRef = ref<FormInstance>()

const instances = ref([])
const cards = ref([])
const nodes = ref([])
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const form = reactive({
  id: '',
  name: '',
  card: '',
  node: '',
  config: ''
})

const rules = {
  name: [{ required: true, message: '请输入实例名称', trigger: 'blur' }],
  card: [{ required: true, message: '请选择游戏卡片', trigger: 'change' }],
  node: [{ required: true, message: '请选择运行节点', trigger: 'change' }]
}

const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    running: 'success',
    stopped: 'info',
    error: 'danger',
    starting: 'warning',
    stopping: 'warning'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    running: '运行中',
    stopped: '已停止',
    error: '错误',
    starting: '启动中',
    stopping: '停止中'
  }
  return statusMap[status] || '未知'
}

const handleAdd = () => {
  dialogType.value = 'add'
  dialogVisible.value = true
  Object.assign(form, {
    id: '',
    name: '',
    card: '',
    node: '',
    config: ''
  })
}

const handleEdit = (row: any) => {
  dialogType.value = 'edit'
  dialogVisible.value = true
  Object.assign(form, row)
}

const handleControl = (row: any) => {
  controlDialogVisible.value = true
  Object.assign(form, row)
}

const handleMonitor = (row: any) => {
  // TODO: 实现实例监控功能
  ElMessage.info('监控功能开发中')
}

const handleConfig = (row: any) => {
  // TODO: 实现实例配置功能
  ElMessage.info('配置功能开发中')
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(
    '确定要删除该实例吗？',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await instanceStore.deleteInstance(row.id)
      ElMessage.success('删除成功')
      fetchInstances()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

const handleStart = async () => {
  try {
    await instanceStore.startInstance(form.id)
    ElMessage.success('启动成功')
    controlDialogVisible.value = false
    fetchInstances()
  } catch (error) {
    ElMessage.error('启动失败')
  }
}

const handleStop = async () => {
  try {
    await instanceStore.stopInstance(form.id)
    ElMessage.success('停止成功')
    controlDialogVisible.value = false
    fetchInstances()
  } catch (error) {
    ElMessage.error('停止失败')
  }
}

const handleRestart = async () => {
  try {
    await instanceStore.restartInstance(form.id)
    ElMessage.success('重启成功')
    controlDialogVisible.value = false
    fetchInstances()
  } catch (error) {
    ElMessage.error('重启失败')
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (dialogType.value === 'add') {
          await instanceStore.createInstance(form)
          ElMessage.success('创建成功')
        } else {
          await instanceStore.updateInstance(form)
          ElMessage.success('更新成功')
        }
        dialogVisible.value = false
        fetchInstances()
      } catch (error) {
        ElMessage.error(dialogType.value === 'add' ? '创建失败' : '更新失败')
      }
    }
  })
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  fetchInstances()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  fetchInstances()
}

const fetchInstances = async () => {
  loading.value = true
  try {
    const response = await instanceStore.getInstances({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    instances.value = response.items
    total.value = response.total
  } catch (error) {
    ElMessage.error('获取实例列表失败')
  } finally {
    loading.value = false
  }
}

const fetchCards = async () => {
  try {
    const response = await cardStore.getCards({ page: 1, pageSize: 100 })
    cards.value = response.items
  } catch (error) {
    ElMessage.error('获取游戏卡片列表失败')
  }
}

const fetchNodes = async () => {
  try {
    const response = await nodeStore.getNodes({ page: 1, pageSize: 100 })
    nodes.value = response.items
  } catch (error) {
    ElMessage.error('获取节点列表失败')
  }
}

onMounted(() => {
  fetchInstances()
  fetchCards()
  fetchNodes()
})
</script>

<style scoped>
.instances-container {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination-container {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}

.control-buttons {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}
</style> 