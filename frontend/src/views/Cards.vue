<template>
  <div class="cards">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏卡片管理</span>
          <el-button type="primary" @click="handleAdd">添加卡片</el-button>
        </div>
      </template>
      
      <el-table :data="cards" style="width: 100%" v-loading="loading">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" />
        <el-table-column prop="platform" label="平台" />
        <el-table-column prop="type" label="类型">
          <template #default="{ row }">
            <el-tag :type="getCardTypeTag(row.type)">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态">
          <template #default="{ row }">
            <el-tag :type="row.status === 'active' ? 'success' : 'danger'">
              {{ row.status === 'active' ? '活跃' : '未活跃' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="lastUpdate" label="最后更新" />
        <el-table-column label="操作" width="200">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="primary" link @click="handleDetail(row)">详情</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <div class="pagination">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
    
    <!-- 添加/编辑卡片对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加卡片' : '编辑卡片'"
      width="500px"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="卡片名称" prop="name">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="所属平台" prop="platform">
          <el-select v-model="form.platform" placeholder="请选择平台">
            <el-option label="Steam" value="Steam" />
            <el-option label="Epic Games" value="Epic Games" />
            <el-option label="GOG" value="GOG" />
          </el-select>
        </el-form-item>
        <el-form-item label="游戏类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择游戏类型">
            <el-option label="FPS" value="fps" />
            <el-option label="RPG" value="rpg" />
            <el-option label="Battle Royale" value="battle_royale" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-switch
            v-model="form.status"
            :active-value="'active'"
            :inactive-value="'inactive'"
          />
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
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { FormInstance } from 'element-plus'
import { mockCards } from '../mock/data'

const loading = ref(false)
const currentPage = ref(1)
const pageSize = ref(10)
const total = ref(mockCards.length)
const dialogVisible = ref(false)
const dialogType = ref<'add' | 'edit'>('add')
const formRef = ref<FormInstance>()

const cards = ref(mockCards)

const form = reactive({
  name: '',
  platform: '',
  type: '',
  status: 'active'
})

const rules = {
  name: [{ required: true, message: '请输入卡片名称', trigger: 'blur' }],
  platform: [{ required: true, message: '请选择所属平台', trigger: 'change' }],
  type: [{ required: true, message: '请选择游戏类型', trigger: 'change' }]
}

const getCardTypeTag = (type: string) => {
  const types: Record<string, string> = {
    fps: 'success',
    rpg: 'warning',
    battle_royale: 'info'
  }
  return types[type] || 'info'
}

const handleAdd = () => {
  dialogType.value = 'add'
  dialogVisible.value = true
  form.name = ''
  form.platform = ''
  form.type = ''
  form.status = 'active'
}

const handleEdit = (row: any) => {
  dialogType.value = 'edit'
  dialogVisible.value = true
  Object.assign(form, row)
}

const handleDetail = (row: any) => {
  ElMessage.info('详情功能开发中')
}

const handleDelete = (row: any) => {
  ElMessageBox.confirm('确定要删除该卡片吗？', '提示', {
    type: 'warning'
  }).then(() => {
    const index = cards.value.findIndex(item => item.id === row.id)
    if (index > -1) {
      cards.value.splice(index, 1)
      ElMessage.success('删除成功')
    }
  })
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate((valid) => {
    if (valid) {
      if (dialogType.value === 'add') {
        const newCard = {
          id: String(cards.value.length + 1),
          ...form,
          lastUpdate: new Date().toLocaleString()
        }
        cards.value.push(newCard)
        ElMessage.success('添加成功')
      } else {
        const index = cards.value.findIndex(item => item.id === form.id)
        if (index > -1) {
          cards.value[index] = { ...form }
          ElMessage.success('更新成功')
        }
      }
      dialogVisible.value = false
    }
  })
}

const handleSizeChange = (val: number) => {
  pageSize.value = val
  // 这里应该重新加载数据
}

const handleCurrentChange = (val: number) => {
  currentPage.value = val
  // 这里应该重新加载数据
}
</script>

<style scoped>
.cards {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.pagination {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
</style> 