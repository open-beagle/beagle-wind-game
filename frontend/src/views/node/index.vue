<template>
  <div class="node-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏节点管理</span>
          <el-button type="primary" @click="handleAdd">添加节点</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="nodeList"
        style="width: 100%"
        border
        fit
      >
        <el-table-column prop="id" label="ID" width="120">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleViewDetail(row)">
              {{ row.id }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="hardware.cpu" label="CPU" min-width="150" />
        <el-table-column prop="hardware.memory" label="内存" min-width="150" />
        <el-table-column prop="hardware.disk" label="磁盘" min-width="150" />
        <el-table-column prop="network.ip" label="IP地址" min-width="150" />
        <el-table-column
          prop="lastHeartbeat"
          label="最后心跳"
          min-width="180"
        />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)"
              >编辑</el-button
            >
            <el-button link type="primary" @click="handleMonitor(row)"
              >监控</el-button
            >
            <el-button link type="danger" @click="handleDelete(row)"
              >删除</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model:current-page="query.page"
          v-model:page-size="query.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.node-container {
  padding: 20px;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
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

:deep(.el-table) {
  flex: 1;
  height: 100%;
}

:deep(.el-table__body-wrapper) {
  overflow-y: auto;
}
</style>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { mockGameNodes } from '@/mocks/data/GameNode';
import type { GameNode, GameNodeQuery } from '@/types/GameNode';
import { useRouter } from 'vue-router'

const loading = ref(false);
const nodeList = ref<GameNode[]>([]);
const total = ref(0);
const query = ref<GameNodeQuery>({
  page: 1,
  pageSize: 10,
});

const router = useRouter()

// 获取节点列表
const getNodeList = async () => {
  loading.value = true;
  try {
    // 模拟 API 请求延迟
    await new Promise(resolve => setTimeout(resolve, 300));
    
    // 模拟数据处理
    const start = (query.value.page - 1) * query.value.pageSize;
    const end = start + query.value.pageSize;
    nodeList.value = mockGameNodes.slice(start, end);
    total.value = mockGameNodes.length;
  } catch (error) {
    ElMessage.error('加载数据失败');
    console.error('加载数据失败:', error);
  } finally {
    loading.value = false;
  }
};

// 状态类型
const statusTypeMap: Record<string, string> = {
  online: 'success',
  offline: 'danger',
  maintenance: 'warning'
};

const getStatusType = (status: string) => statusTypeMap[status] || 'info';

// 状态文本
const statusTextMap: Record<string, string> = {
  online: '在线',
  offline: '离线',
  maintenance: '维护中'
};

const getStatusText = (status: string) => statusTextMap[status] || status;

// 分页处理
const handleSizeChange = (val: number) => {
  query.value.pageSize = val;
  getNodeList();
};

const handleCurrentChange = (val: number) => {
  query.value.page = val;
  getNodeList();
};

// 查看详情
const handleViewDetail = (row: GameNode) => {
  router.push(`/node/detail/${row.id}`)
}

// 初始化数据
onMounted(() => {
  getNodeList();
});
</script>
