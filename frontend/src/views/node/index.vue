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
        <el-table-column prop="id" label="节点信息" min-width="200">
          <template #default="{ row }">
            <div class="node-info">
              <div class="node-id">
                <el-button link type="primary" @click="handleViewDetail(row)">
                  {{ row.id }}
                </el-button>
              </div>
              
              <div class="node-name">{{ row.name }}</div>
              
              <div class="node-model" v-if="row.model">型号: {{ row.model }}</div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="运行状态" width="150">
          <template #default="{ row }">
            <div class="status-info">
              <div class="status-tag">
                <el-tag :type="getStatusType(row.status)">
                  {{ getStatusText(row.status) }}
                </el-tag>
              </div>
              <div class="ip-info">
                <span v-if="getNodeIp(row) !== '-'">
                  <el-icon><Location /></el-icon> IP: {{ getNodeIp(row) }}
                </span>
                <span v-else class="no-ip">未设置IP</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="resources" label="硬件配置" min-width="300">
          <template #default="{ row }">
            <div class="hardware-info">
              <div v-if="row.hardware?.CPU || row.resources?.cpu">
                <el-icon><Cpu /></el-icon> CPU: {{ row.hardware?.CPU || formatCpu(row.resources?.cpu) }}
              </div>
              <div v-if="row.hardware?.GPU || row.resources?.gpu">
                <el-icon><DataAnalysis /></el-icon> GPU: {{ row.hardware?.GPU || formatGpu(row.resources?.gpu) }}
              </div>
              <div v-if="row.hardware?.RAM || row.resources?.memory">
                <el-icon><Monitor /></el-icon> 内存: {{ row.hardware?.RAM || formatMemory(row.resources?.memory) }}
              </div>
              <div v-if="row.hardware?.Storage || row.resources?.storage">
                <el-icon><Document /></el-icon> 磁盘: {{ row.hardware?.Storage || formatStorage(row.resources?.storage) }}
              </div>
              <div v-if="row.hardware?.Network || row.network?.bandwidth || row.network?.speed || row.resources?.network">
                <el-icon><Connection /></el-icon> 网络: {{ formatNetwork(row) }}
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column
          label="时间信息"
          min-width="200"
        >
          <template #default="{ row }">
            <div class="time-info">
              <div class="time-item">
                <span class="time-label">创建:</span>
                <span class="time-value">{{ formatTime(row.created_at) }}</span>
              </div>
              <div class="time-item">
                <span class="time-label">更新:</span>
                <span class="time-value">{{ formatTime(row.updated_at) }}</span>
              </div>
              <div class="time-item">
                <span class="time-label">在线:</span>
                <span class="time-value">{{ formatTime(row.last_online) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleMonitor(row)"
              >监控</el-button
            >
            <el-button link type="primary" @click="handleEdit(row)"
              >编辑</el-button
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

.hardware-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.hardware-info div {
  display: flex;
  align-items: center;
}

.hardware-info i {
  margin-right: 5px;
  font-size: 14px;
}

.node-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.node-id .el-button {
  padding: 0;
  font-size: 13px;
}

.node-name {
  font-size: 14px;
  color: #303133;
}

.node-model {
  color: #909399;
  font-size: 12px;
}

.status-info {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.ip-info {
  font-size: 12px;
  color: #606266;
  display: flex;
  align-items: center;
  gap: 4px;
}

.no-ip {
  color: #909399;
  font-style: italic;
}

.time-info {
  display: flex;
  flex-direction: column;
  gap: 4px;
  font-size: 12px;
}

.time-item {
  display: flex;
  align-items: center;
}

.time-label {
  color: #909399;
  width: 40px;
}

.time-value {
  color: #606266;
}

.time-value:empty::after {
  content: "-";
  color: #909399;
  font-style: italic;
}
</style>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import { Cpu, Monitor, Document, Connection, DataAnalysis, Location } from '@element-plus/icons-vue';
import type { GameNode, NodeQuery } from '@/types/GameNode';
import { getNodeList, deleteNode } from '@/services/nodeService';
import { useRouter } from 'vue-router'

const loading = ref(false);
const nodeList = ref<GameNode[]>([]);
const total = ref(0);
const query = ref<NodeQuery>({
  page: 1,
  pageSize: 10,
});

const router = useRouter()

// 获取节点列表
const fetchNodeList = async () => {
  loading.value = true;
  try {
    const result = await getNodeList(query.value)
    
    // 删除字段映射处理，直接使用后端返回的原始字段
    
    nodeList.value = result.list
    total.value = result.total
  } catch (error) {
    ElMessage.error('加载数据失败');
    console.error('加载数据失败:', error);
  } finally {
    loading.value = false;
  }
};

// 状态类型
const statusTypeMap: Record<string, string> = {
  online: 'info',
  offline: 'danger',
  maintenance: 'warning',
  ready: 'success'
};

const getStatusType = (status: string) => statusTypeMap[status] || 'info';

// 状态文本
const statusTextMap: Record<string, string> = {
  online: '在线',
  offline: '离线',
  maintenance: '维护中',
  ready: '就绪'
};

const getStatusText = (status: string) => statusTextMap[status] || status;

// 格式化内存显示
const formatMemory = (memory?: number) => {
  if (!memory) return '-';
  if (memory >= 1024) {
    return `${(memory / 1024).toFixed(1)} GB`;
  }
  return `${memory} MB`;
};

// 格式化存储显示
const formatStorage = (storage?: number) => {
  if (!storage) return '-';
  if (storage >= 1024) {
    return `${(storage / 1024).toFixed(1)} TB`;
  }
  return `${storage} GB`;
};

// 分页处理
const handleSizeChange = (val: number) => {
  query.value.pageSize = val;
  fetchNodeList();
};

const handleCurrentChange = (val: number) => {
  query.value.page = val;
  fetchNodeList();
};

// 查看详情
const handleViewDetail = (row: GameNode) => {
  router.push(`/nodes/detail/${row.id}`)
}

// 删除节点
const handleDelete = (row: GameNode) => {
  ElMessageBox.confirm(
    `确定要删除节点 "${row.name}" 吗？此操作不可恢复`,
    '删除确认',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning',
    }
  )
    .then(async () => {
      try {
        const success = await deleteNode(row.id)
        if (success) {
          ElMessage.success('删除成功')
          fetchNodeList()
        } else {
          ElMessage.error('删除失败')
        }
      } catch (error) {
        ElMessage.error('删除失败')
        console.error('删除失败:', error)
      }
    })
    .catch(() => {
      // 用户取消删除
    })
}

// 添加节点
const handleAdd = () => {
  router.push('/node/create')
}

// 编辑节点
const handleEdit = (row: GameNode) => {
  router.push(`/node/edit/${row.id}`)
}

// 节点监控
const handleMonitor = (row: GameNode) => {
  router.push(`/node/monitor/${row.id}`)
}

// 获取节点IP地址
const getNodeIp = (node: any): string => {
  // 直接检查顶层ip属性
  if (node.ip) return node.ip;
  
  // 检查network对象中可能的ip字段(大小写不敏感)
  if (node.network) {
    // 检查常见格式
    if (node.network.ip) return node.network.ip;
    if (node.network.IP) return node.network.IP;
    
    // 遍历所有属性
    for (const key in node.network) {
      const lowerKey = key.toLowerCase();
      if (lowerKey === 'ip' || lowerKey.includes('ip')) {
        return node.network[key];
      }
    }
  }
  
  // 检查是否在hardware.Network或resources.ip等嵌套层级
  if (node.hardware?.Network?.includes?.('192.168.')) {
    return node.hardware.Network.split(' ')[0];
  }
  
  // 检查labels
  if (node.labels?.ip) return node.labels.ip;
  
  console.log('未找到IP地址，完整节点数据:', node);
  
  return '-';
};

// 格式化CPU信息
const formatCpu = (cpu?: number): string => {
  if (!cpu) return '-';
  return `${cpu} 核`;
};

// 格式化GPU信息
const formatGpu = (gpu?: any): string => {
  if (!gpu) return '-';
  if (typeof gpu === 'string') return gpu;
  if (gpu.model) {
    if (gpu.memory) {
      return `${gpu.model} (${formatMemory(gpu.memory)})`;
    }
    return gpu.model;
  }
  return '-';
};

// 格式化网络信息
const formatNetwork = (node: any): string => {
  if (node.hardware?.Network) return node.hardware.Network;
  if (node.network?.speed) return node.network.speed;
  if (node.network?.bandwidth) return `${node.network.bandwidth} Mbps`;
  if (node.resources?.network) return `${node.resources.network} Mbps`;
  return '-';
};

// 格式化时间显示
const formatTime = (timeStr: string | null | undefined) => {
  if (!timeStr) return '';
  try {
    // 解析日期
    const date = new Date(timeStr);
    if (isNaN(date.getTime())) return timeStr;
    
    // 创建上海时区的日期格式化器
    return new Intl.DateTimeFormat('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hour12: false,
      timeZone: 'Asia/Shanghai'
    }).format(date);
  } catch (e) {
    return '';
  }
};

// 初始化数据
onMounted(() => {
  fetchNodeList();
});
</script>
