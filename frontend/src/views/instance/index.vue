<template>
  <div class="instance-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏实例管理</span>
          <el-button type="primary" @click="handleAdd">创建实例</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="instanceList"
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
        <el-table-column
          prop="gameCard.name"
          label="游戏名称"
          min-width="150"
        />
        <el-table-column
          prop="gameCard.platform.name"
          label="平台"
          min-width="120"
        />
        <el-table-column prop="node.name" label="节点" min-width="150" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="config.maxPlayers"
          label="最大玩家数"
          width="120"
        />
        <el-table-column prop="config.port" label="端口" width="100" />
        <el-table-column prop="createdAt" label="创建时间" min-width="180" />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)"
              >编辑</el-button
            >
            <el-button
              link
              type="success"
              @click="handleStart(row)"
              v-if="row.status === 'stopped'"
              >启动</el-button
            >
            <el-button
              link
              type="warning"
              @click="handleStop(row)"
              v-if="row.status === 'running'"
              >停止</el-button
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
.instance-container {
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
import { mockGameInstances } from '@/mocks/data/GameInstance';
import type { GameInstance, GameInstanceQuery } from '@/types/GameInstance';
import { useRouter } from 'vue-router'

const loading = ref(false);
const instanceList = ref<GameInstance[]>([]);
const total = ref(0);
const query = ref<GameInstanceQuery>({
  page: 1,
  pageSize: 10,
});

const router = useRouter()

// 获取实例列表
const getInstanceList = async () => {
  loading.value = true;
  try {
    // 模拟 API 请求延迟
    await new Promise(resolve => setTimeout(resolve, 300));
    
    // 模拟数据处理
    const start = (query.value.page - 1) * query.value.pageSize;
    const end = start + query.value.pageSize;
    instanceList.value = mockGameInstances.slice(start, end);
    total.value = mockGameInstances.length;
  } catch (error) {
    ElMessage.error('加载数据失败');
    console.error('加载数据失败:', error);
  } finally {
    loading.value = false;
  }
};

// 状态类型
const statusTypeMap: Record<string, string> = {
  running: 'success',
  stopped: 'info',
  error: 'danger',
  starting: 'warning',
  stopping: 'warning'
};

const getStatusType = (status: string) => statusTypeMap[status] || 'info';

// 状态文本
const statusTextMap: Record<string, string> = {
  running: '运行中',
  stopped: '已停止',
  error: '错误',
  starting: '启动中',
  stopping: '停止中'
};

const getStatusText = (status: string) => statusTextMap[status] || status;

// 分页处理
const handleSizeChange = (val: number) => {
  query.value.pageSize = val;
  getInstanceList();
};

const handleCurrentChange = (val: number) => {
  query.value.page = val;
  getInstanceList();
};

// 查看详情
const handleViewDetail = (row: GameInstance) => {
  router.push(`/instance/detail/${row.id}`)
}

// 初始化数据
onMounted(() => {
  getInstanceList();
});
</script>
