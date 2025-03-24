<template>
  <div class="platform-container">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>游戏平台管理</span>
          <el-button type="primary" @click="handleAdd">添加平台</el-button>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="platformList"
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
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag type="info">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="版本" width="120" />
        <el-table-column label="特性" min-width="200">
          <template #default="{ row }">
            <el-tag
              v-for="feature in row.features"
              :key="feature"
              class="feature-tag"
              size="small"
            >
              {{ feature }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="配置" min-width="200">
          <template #default="{ row }">
            <el-tooltip
              :content="JSON.stringify(row.config, null, 2)"
              placement="top"
              :show-after="500"
            >
              <el-button link type="primary">查看配置</el-button>
            </el-tooltip>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
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

    <!-- 添加/编辑平台对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'add' ? '添加平台' : '编辑平台'"
      width="600px"
    >
      <el-form ref="formRef" :model="form" :rules="rules" label-width="100px">
        <el-form-item label="平台名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入平台名称" />
        </el-form-item>
        <el-form-item label="平台类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择平台类型">
            <el-option label="游戏平台" value="gaming" />
            <el-option label="模拟器" value="emulator" />
          </el-select>
        </el-form-item>
        <el-form-item label="版本" prop="version">
          <el-input v-model="form.version" placeholder="请输入版本号" />
        </el-form-item>
        <el-form-item label="镜像" prop="image">
          <el-input v-model="form.image" placeholder="请输入容器镜像地址" />
        </el-form-item>
        <el-form-item label="启动路径" prop="bin">
          <el-input v-model="form.bin" placeholder="请输入启动路径" />
        </el-form-item>
        <el-form-item label="操作系统" prop="os">
          <el-input v-model="form.os" placeholder="请输入操作系统" />
        </el-form-item>
        <el-form-item label="特性" prop="features">
          <el-tag
            v-for="(feature, index) in form.features"
            :key="index"
            closable
            class="feature-tag"
            @close="handleRemoveFeature(index)"
          >
            {{ feature }}
          </el-tag>
          <el-input
            v-if="featureInputVisible"
            ref="featureInputRef"
            v-model="featureInputValue"
            class="feature-input"
            size="small"
            @keyup.enter="handleAddFeature"
            @blur="handleAddFeature"
          />
          <el-button
            v-else
            class="button-new-tag"
            size="small"
            @click="showFeatureInput"
          >
            + 添加特性
          </el-button>
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
import { ref, reactive, nextTick, onMounted } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import type { FormInstance } from "element-plus";
import { mockGamePlatforms } from "@/mocks/data/GamePlatform";
import type {
  GamePlatform,
  GamePlatformQuery,
  GamePlatformForm,
} from "@/types/GamePlatform";
import { useRouter } from "vue-router";

const loading = ref(false);
const platformList = ref<GamePlatform[]>([]);
const total = ref(0);
const query = ref<GamePlatformQuery>({
  page: 1,
  pageSize: 10,
});

// 对话框相关
const dialogVisible = ref(false);
const dialogType = ref<"add" | "edit">("add");
const formRef = ref<FormInstance>();
const form = reactive<GamePlatformForm>({
  name: "",
  type: "gaming",
  version: "",
  image: "",
  bin: "",
  os: "Linux",
  files: [],
  features: [],
  config: {},
  status: "inactive"
});

// 特性输入相关
const featureInputVisible = ref(false);
const featureInputValue = ref("");
const featureInputRef = ref<HTMLInputElement>();

// 表单验证规则
const rules = {
  name: [{ required: true, message: "请输入平台名称", trigger: "blur" }],
  type: [{ required: true, message: "请选择平台类型", trigger: "change" }],
  version: [{ required: true, message: "请输入版本号", trigger: "blur" }],
  image: [{ required: true, message: "请输入容器镜像地址", trigger: "blur" }],
  bin: [{ required: true, message: "请输入启动路径", trigger: "blur" }],
  data: [{ required: true, message: "请输入数据目录", trigger: "blur" }],
};

const router = useRouter();

// 获取平台列表
const getPlatformList = async () => {
  loading.value = true;
  try {
    // 模拟 API 请求延迟
    await new Promise((resolve) => setTimeout(resolve, 300));

    // 模拟数据处理
    const start = (query.value.page - 1) * query.value.pageSize;
    const end = start + query.value.pageSize;
    platformList.value = mockGamePlatforms.slice(start, end);
    total.value = mockGamePlatforms.length;
  } catch (error) {
    ElMessage.error("加载数据失败");
    console.error("加载数据失败:", error);
  } finally {
    loading.value = false;
  }
};

// 分页处理
const handleSizeChange = (val: number) => {
  query.value.pageSize = val;
  getPlatformList();
};

const handleCurrentChange = (val: number) => {
  query.value.page = val;
  getPlatformList();
};

// 添加平台
const handleAdd = () => {
  dialogType.value = "add";
  dialogVisible.value = true;
  Object.assign(form, {
    name: "",
    type: "gaming",
    version: "",
    image: "",
    bin: "",
    data: "",
    os: "Linux",
    files: [],
    features: [],
    config: {},
    status: "inactive"
  });
};

// 编辑平台
const handleEdit = (row: GamePlatform) => {
  dialogType.value = "edit";
  dialogVisible.value = true;
  Object.assign(form, {
    id: row.id,
    name: row.name,
    type: row.type,
    version: row.version,
    image: row.image,
    bin: row.bin,
    os: row.os,
    files: [...row.files],
    features: [...row.features],
    config: { ...row.config },
    status: row.status
  });
};

// 删除平台
const handleDelete = (row: GamePlatform) => {
  ElMessageBox.confirm("确定要删除该平台吗？", "提示", {
    type: "warning",
  }).then(() => {
    const index = platformList.value.findIndex((item) => item.id === row.id);
    if (index > -1) {
      platformList.value.splice(index, 1);
      total.value--;
      ElMessage.success("删除成功");
    }
  });
};

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return;

  await formRef.value.validate((valid) => {
    if (valid) {
      if (dialogType.value === "add") {
        // 创建新平台
        const newPlatform: GamePlatform = {
          id: String(Date.now()),
          name: form.name || "",
          type: form.type || "gaming",
          version: form.version || "",
          image: form.image || "",
          bin: form.bin || "",
          os: form.os || "Linux",
          files: form.files || [],
          features: form.features || [],
          config: form.config || {},
          status: form.status || "inactive",
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
        };
        platformList.value.push(newPlatform);
        total.value++;
        ElMessage.success("添加成功");
      } else {
        const index = platformList.value.findIndex(
          (item) => item.id === form.id
        );
        if (index > -1) {
          // 更新时保留原有的created_at字段
          platformList.value[index] = {
            ...platformList.value[index],
            name: form.name || "",
            type: form.type || "gaming",
            version: form.version || "",
            image: form.image || "",
            bin: form.bin || "",
            os: form.os || "Linux",
            files: form.files || [],
            features: form.features || [],
            config: form.config || {},
            status: form.status || "inactive",
            updated_at: new Date().toISOString(),
          };
          ElMessage.success("更新成功");
        }
      }
      dialogVisible.value = false;
    }
  });
};

// 显示特性输入框
const showFeatureInput = () => {
  featureInputVisible.value = true;
  nextTick(() => {
    featureInputRef.value?.focus();
  });
};

// 添加特性
const handleAddFeature = () => {
  if (featureInputValue.value) {
    if (!form.features) {
      form.features = [];
    }
    form.features.push(featureInputValue.value);
  }
  featureInputVisible.value = false;
  featureInputValue.value = "";
};

// 移除特性
const handleRemoveFeature = (index: number) => {
  if (!form.features) {
    form.features = [];
  }
  form.features.splice(index, 1);
};

// 查看详情
const handleViewDetail = (row: GamePlatform) => {
  router.push(`/platforms/detail/${row.id}`);
};

// 初始化数据
onMounted(() => {
  getPlatformList();
});
</script>

<style scoped>
.platform-container {
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

.feature-tag {
  margin-right: 4px;
  margin-bottom: 4px;
}

.feature-input {
  width: 100px;
  margin-right: 4px;
  vertical-align: bottom;
}

.button-new-tag {
  margin-left: 4px;
  height: 24px;
  padding-top: 0;
  padding-bottom: 0;
}
</style>
