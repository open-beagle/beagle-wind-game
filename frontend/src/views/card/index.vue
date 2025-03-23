<template>
  <div class="game-cards-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏卡片管理</span>
          <el-button type="primary" @click="handleAdd">添加游戏</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="gameCards"
        style="width: 100%"
        border
        fit
      >
        <el-table-column prop="id" label="游戏ID" width="120">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleViewDetail(row)">
              {{ row.id }}
            </el-button>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="游戏名称" width="150" />
        <el-table-column prop="platform.name" label="平台" width="120">
          <template #default="{ row }">
            <el-tag :type="getPlatformType(row.platform.type)">
              {{ row.platform.name }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column
          prop="description"
          label="描述"
          min-width="200"
          show-overflow-tooltip
        />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createdAt" label="创建时间" width="180" />
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)"
              >编辑</el-button
            >
            <el-button link type="success" @click="handleCreateInstance(row)"
              >创建实例</el-button
            >
            <el-button link type="danger" @click="handleDelete(row)"
              >删除</el-button
            >
          </template>
        </el-table-column>
      </el-table>

      <div class="pagination">
        <el-pagination
          v-model="query.page"
          :page-size="query.pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </el-card>

    <!-- 创建/编辑游戏对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'create' ? '添加游戏' : '编辑游戏'"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
      >
        <el-form-item label="游戏名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入游戏名称" />
        </el-form-item>
        <el-form-item label="游戏平台" prop="platformId">
          <el-select v-model="form.platformId" placeholder="请选择游戏平台">
            <el-option
              v-for="platform in platforms"
              :key="platform.id"
              :label="platform.name"
              :value="platform.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="游戏描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入游戏描述"
          />
        </el-form-item>
        <el-form-item label="封面图片" prop="coverImage">
          <el-upload
            class="cover-uploader"
            action="/api/upload"
            :show-file-list="false"
            :on-success="handleCoverSuccess"
            :before-upload="beforeCoverUpload"
          >
            <img v-if="form.coverImage" :src="form.coverImage" class="cover" />
            <el-icon v-else class="cover-uploader-icon"><Plus /></el-icon>
          </el-upload>
        </el-form-item>
        <el-form-item label="游戏状态" prop="status">
          <el-select v-model="form.status" placeholder="请选择游戏状态">
            <el-option label="草稿" value="draft" />
            <el-option label="已发布" value="published" />
            <el-option label="已归档" value="archived" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 创建实例对话框 -->
    <el-dialog
      v-model="instanceDialogVisible"
      title="创建游戏实例"
      width="500px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="instanceFormRef"
        :model="instanceForm"
        :rules="instanceRules"
        label-width="100px"
      >
        <el-form-item label="实例名称" prop="name">
          <el-input v-model="instanceForm.name" placeholder="请输入实例名称" />
        </el-form-item>
        <el-form-item label="游戏节点" prop="nodeId">
          <el-select v-model="instanceForm.nodeId" placeholder="请选择游戏节点">
            <el-option
              v-for="node in availableNodes"
              :key="node.id"
              :label="node.name"
              :value="node.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="最大玩家数" prop="config.maxPlayers">
          <el-input-number v-model="instanceForm.config.maxPlayers" :min="1" :max="100" />
        </el-form-item>
        <el-form-item label="端口" prop="config.port">
          <el-input-number v-model="instanceForm.config.port" :min="1024" :max="65535" />
        </el-form-item>
        <el-form-item label="地图" prop="config.settings.map">
          <el-select v-model="instanceForm.config.settings.map" placeholder="请选择地图">
            <el-option label="默认地图" value="default" />
            <el-option label="竞技场" value="arena" />
            <el-option label="生存模式" value="survival" />
          </el-select>
        </el-form-item>
        <el-form-item label="难度" prop="config.settings.difficulty">
          <el-select v-model="instanceForm.config.settings.difficulty" placeholder="请选择难度">
            <el-option label="简单" value="easy" />
            <el-option label="普通" value="normal" />
            <el-option label="困难" value="hard" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="instanceDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleInstanceSubmit">确定</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import { Plus } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'
import { mockGameCards, mockGamePlatforms, mockGameNodes, mockGameInstances } from "@/mocks";
import type { GameCard, GameCardQuery, GamePlatform, GameCardStatus, GameInstance, GameInstanceForm } from "@/types";
import type { FormInstance, FormRules } from 'element-plus'

const router = useRouter()
const loading = ref(false);
const gameCards = ref<GameCard[]>([]);
const total = ref(0);
const query = ref<GameCardQuery>({
  page: 1,
  pageSize: 10,
});

// 对话框相关
const dialogVisible = ref(false)
const dialogType = ref<'create' | 'edit'>('create')
const formRef = ref<FormInstance>()
const platforms = ref(mockGamePlatforms)

const form = ref({
  id: '',
  name: '',
  platformId: '',
  description: '',
  coverImage: '',
  status: 'draft' as GameCardStatus
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入游戏名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  platformId: [
    { required: true, message: '请选择游戏平台', trigger: 'change' }
  ],
  description: [
    { required: true, message: '请输入游戏描述', trigger: 'blur' },
    { max: 500, message: '长度不能超过 500 个字符', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择游戏状态', trigger: 'change' }
  ]
}

// 实例创建相关
const instanceDialogVisible = ref(false)
const instanceFormRef = ref<FormInstance>()
const currentGame = ref<GameCard | null>(null)
const availableNodes = ref(mockGameNodes)

const instanceForm = ref<GameInstanceForm>({
  id: '',
  name: '',
  gameCardId: '',
  nodeId: '',
  config: {
    maxPlayers: 10,
    port: 8080,
    settings: {
      map: 'default',
      difficulty: 'normal'
    }
  }
})

const instanceRules: FormRules = {
  name: [
    { required: true, message: '请输入实例名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  nodeId: [
    { required: true, message: '请选择游戏节点', trigger: 'change' }
  ],
  'config.maxPlayers': [
    { required: true, message: '请输入最大玩家数', trigger: 'blur' }
  ],
  'config.port': [
    { required: true, message: '请输入端口号', trigger: 'blur' }
  ],
  'config.settings.map': [
    { required: true, message: '请选择地图', trigger: 'change' }
  ],
  'config.settings.difficulty': [
    { required: true, message: '请选择难度', trigger: 'change' }
  ]
}

// 获取游戏卡片列表
const getGameCardList = async () => {
  loading.value = true;
  try {
    // 模拟 API 请求延迟
    await new Promise((resolve) => setTimeout(resolve, 300));

    // 模拟数据处理
    const start = (query.value.page - 1) * query.value.pageSize;
    const end = start + query.value.pageSize;
    gameCards.value = mockGameCards.slice(start, end);
    total.value = mockGameCards.length;
  } catch (error) {
    ElMessage.error("加载数据失败");
    console.error("加载数据失败:", error);
  } finally {
    loading.value = false;
  }
};

// 平台类型
const platformTypeMap: Record<string, string> = {
  steam: "primary",
  epic: "success",
  gog: "warning",
  xbox: "info",
  psn: "danger",
  nintendo: "success",
};

const getPlatformType = (type: string) => platformTypeMap[type] || "info";

// 状态类型
const statusTypeMap: Record<string, string> = {
  draft: "info",
  published: "success",
  archived: "warning",
};

const getStatusType = (status: string) => statusTypeMap[status] || "info";

// 状态文本
const statusTextMap: Record<string, string> = {
  draft: "草稿",
  published: "已发布",
  archived: "已归档",
};

const getStatusText = (status: string) => statusTextMap[status] || status;

// 分页处理
const handleSizeChange = (val: number) => {
  query.value.pageSize = val;
  getGameCardList();
};

const handleCurrentChange = (val: number) => {
  query.value.page = val;
  getGameCardList();
};

// 添加游戏
const handleAdd = () => {
  dialogType.value = 'create'
  form.value = {
    id: '',
    name: '',
    platformId: '',
    description: '',
    coverImage: '',
    status: 'draft' as GameCardStatus
  }
  dialogVisible.value = true
}

// 编辑游戏
const handleEdit = (row: GameCard) => {
  dialogType.value = 'edit'
  form.value = {
    id: row.id,
    name: row.name,
    platformId: row.platform.id,
    description: row.description,
    coverImage: row.coverImage,
    status: row.status
  }
  dialogVisible.value = true
}

// 创建实例
const handleCreateInstance = (game: GameCard) => {
  currentGame.value = game
  instanceForm.value = {
    id: '',
    name: `${game.name} 实例`,
    gameCardId: game.id,
    nodeId: '',
    config: {
      maxPlayers: 10,
      port: 8080,
      settings: {
        map: 'default',
        difficulty: 'normal'
      }
    }
  }
  instanceDialogVisible.value = true
}

// 提交实例创建
const handleInstanceSubmit = async () => {
  if (!instanceFormRef.value || !currentGame.value) return
  
  const game = currentGame.value
  await instanceFormRef.value.validate((valid) => {
    if (valid) {
      try {
        // 模拟创建实例
        const newInstance: GameInstance = {
          id: `instance_${Date.now()}`,
          gameId: game.id,
          gameCardId: game.id,
          gameCard: game,
          nodeId: instanceForm.value.nodeId,
          node: availableNodes.value.find(n => n.id === instanceForm.value.nodeId)!,
          userId: 'user_001', // 模拟用户ID
          name: instanceForm.value.name,
          description: `${game.name} 的实例`,
          status: 'stopped',
          resources: {
            cpu: 2,
            memory: 4096,
            storage: 50
          },
          metrics: {
            cpuUsage: 0,
            memoryUsage: 0,
            storageUsage: 0,
            networkUsage: 0,
            uptime: 0,
            fps: 0
          },
          network: {
            ip: '127.0.0.1',
            port: instanceForm.value.config.port
          },
          labels: {},
          config: instanceForm.value.config,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString()
        }
        
        mockGameInstances.push(newInstance)
        ElMessage.success('实例创建成功')
        instanceDialogVisible.value = false
      } catch (error) {
        ElMessage.error('创建实例失败：' + (error as Error).message)
      }
    }
  })
}

// 删除游戏
const handleDelete = (row: GameCard) => {
  ElMessageBox.confirm("确定要删除该游戏吗？", "提示", {
    type: "warning",
  }).then(() => {
    const index = gameCards.value.findIndex((item) => item.id === row.id);
    if (index > -1) {
      gameCards.value.splice(index, 1);
      total.value--;
      ElMessage.success("删除成功");
    }
  });
};

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate((valid) => {
    if (valid) {
      if (dialogType.value === 'create') {
        // 创建新游戏
        const newGame: GameCard = {
          id: `game_${Date.now()}`,
          name: form.value.name,
          platform: platforms.value.find((p: GamePlatform) => p.id === form.value.platformId)!,
          description: form.value.description,
          coverImage: form.value.coverImage,
          status: form.value.status,
          createdAt: new Date().toISOString()
        }
        mockGameCards.push(newGame)
        ElMessage.success('添加成功')
      } else {
        // 更新游戏
        const index = gameCards.value.findIndex(item => item.id === form.value.id)
        if (index > -1) {
          gameCards.value[index] = {
            ...gameCards.value[index],
            name: form.value.name,
            platform: platforms.value.find((p: GamePlatform) => p.id === form.value.platformId)!,
            description: form.value.description,
            coverImage: form.value.coverImage,
            status: form.value.status
          }
          ElMessage.success('更新成功')
        }
      }
      dialogVisible.value = false
      getGameCardList()
    }
  })
}

// 封面图片上传
const handleCoverSuccess = (response: any) => {
  form.value.coverImage = response.url
}

const beforeCoverUpload = (file: File) => {
  const isImage = file.type.startsWith('image/')
  const isLt2M = file.size / 1024 / 1024 < 2

  if (!isImage) {
    ElMessage.error('只能上传图片文件!')
    return false
  }
  if (!isLt2M) {
    ElMessage.error('图片大小不能超过 2MB!')
    return false
  }
  return true
}

// 查看详情
const handleViewDetail = (row: GameCard) => {
  router.push(`/card/detail/${row.id}`)
}

// 初始化数据
onMounted(() => {
  getGameCardList();
});
</script>

<style scoped>
.game-cards-container {
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

.cover-uploader {
  border: 1px dashed #d9d9d9;
  border-radius: 6px;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  width: 178px;
  height: 178px;
}

.cover-uploader:hover {
  border-color: #409EFF;
}

.cover-uploader-icon {
  font-size: 28px;
  color: #8c939d;
  width: 178px;
  height: 178px;
  text-align: center;
  line-height: 178px;
}

.cover {
  width: 178px;
  height: 178px;
  display: block;
  object-fit: cover;
}
</style>
