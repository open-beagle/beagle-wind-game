<template>
  <div class="nodes-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏节点管理</span>
          <el-button type="primary" @click="handleAdd">
            <el-icon><Plus /></el-icon>添加节点
          </el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="nodes"
        style="width: 100%"
        border
      >
        <el-table-column prop="id" label="节点ID" width="120" />
        <el-table-column prop="name" label="节点名称" width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="hardware.cpu" label="CPU" width="120">
          <template #default="{ row }">
            {{ row.hardware?.cpu || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="hardware.memory" label="内存" width="120">
          <template #default="{ row }">
            {{ row.hardware?.memory || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="hardware.disk" label="磁盘" width="120">
          <template #default="{ row }">
            {{ row.hardware?.disk || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="network.ip" label="IP地址" width="150">
          <template #default="{ row }">
            {{ row.network?.ip || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="lastHeartbeat" label="最后心跳" width="180">
          <template #default="{ row }">
            {{ formatTime(row.lastHeartbeat) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" fixed="right" width="200">
          <template #default="{ row }">
            <el-button-group>
              <el-button type="primary" size="small" @click="handleEdit(row)">
                编辑
              </el-button>
              <el-button type="success" size="small" @click="handleMonitor(row)">
                监控
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

    <!-- 添加/编辑节点对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加节点' : '编辑节点'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="节点名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="CPU" prop="hardware.cpu">
          <el-input v-model="form.hardware.cpu" />
        </el-form-item>
        <el-form-item label="内存" prop="hardware.memory">
          <el-input v-model="form.hardware.memory" />
        </el-form-item>
        <el-form-item label="磁盘" prop="hardware.disk">
          <el-input v-model="form.hardware.disk" />
        </el-form-item>
        <el-form-item label="IP地址" prop="network.ip">
          <el-input v-model="form.network.ip" />
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { useNodeStore } from '../stores'

const nodeStore = useNodeStore()
const loading = ref(false)
const dialogVisible = ref(false)
const dialogType = ref<'add' | 'edit'>('add')
const formRef = ref<FormInstance>()
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(0)

const form = reactive({
  name: '',
  hardware: {
    cpu: '',
    memory: '',
    disk: ''
  },
  network: {
    ip: ''
  }
})

const rules = {
  name: [{ required: true, message: '请输入节点名称', trigger: 'blur' }],
  'hardware.cpu': [{ required: true, message: '请输入CPU信息', trigger: 'blur' }],
  'hardware.memory': [{ required: true, message: '请输入内存信息', trigger: 'blur' }],
  'hardware.disk': [{ required: true, message: '请输入磁盘信息', trigger: 'blur' }],
  'network.ip': [{ required: true, message: '请输入IP地址', trigger: 'blur' }]
}

const nodes = ref([])

const getStatusType = (status: string) => {
  const statusMap: Record<string, string> = {
    online: 'success',
    offline: 'danger',
    maintenance: 'warning'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status: string) => {
  const statusMap: Record<string, string> = {
    online: '在线',
    offline: '离线',
    maintenance: '维护中'
  }
  return statusMap[status] || '未知'
}

const formatTime = (time: string) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

const handleAdd = () => {
  dialogType.value = 'add'
  dialogVisible.value = true
  Object.assign(form, {
    name: '',
    hardware: {
      cpu: '',
      memory: '',
      disk: ''
    },
    network: {
      ip: ''
    }
  })
}

const handleEdit = (row: any) => {
  dialogType.value = 'edit'
  dialogVisible.value = true
  Object.assign(form, row)
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm(
    '确定要删除该节点吗？',
    '警告',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(async () => {
    try {
      await nodeStore.deleteNode(row.id)
      ElMessage.success('删除成功')
      fetchNodes()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

const handleMonitor = (row: any) => {
  // TODO: 实现节点监控功能
  ElMessage.info('监控功能开发中')
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      try {
        if (dialogType.value === 'add') {
          await nodeStore.createNode(form)
          ElMessage.success('添加成功')
        } else {
          await nodeStore.updateNode(form)
          ElMessage.success('更新成功')
        }
        dialogVisible.value = false
        fetchNodes()
      } catch (error) {
        ElMessage.error(dialogType.value === 'add' ? '添加失败' : '更新失败')
      }
    }
  })
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  fetchNodes()
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  fetchNodes()
}

const fetchNodes = async () => {
  loading.value = true
  try {
    const response = await nodeStore.getNodes({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    nodes.value = response.data
    total.value = response.total
  } catch (error) {
    ElMessage.error('获取节点列表失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchNodes()
})
</script>

<style scoped>
.nodes-container {
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
</style> 