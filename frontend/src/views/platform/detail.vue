<template>
  <div class="platform-detail-container">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">
          <span>平台详情</span>
          <div class="header-actions">
            <el-button v-if="!isEditing" type="primary" @click="handleEdit"
              >编辑</el-button
            >
            <el-button v-if="isEditing" type="success" @click="handleSave"
              >保存</el-button
            >
            <el-button v-if="isEditing" type="info" @click="handleCancel"
              >取消</el-button
            >
            <el-button type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <!-- API错误状态显示 -->
      <api-error-state
        v-if="error"
        :title="error.title"
        :message="error.message"
        :detail="error.detail"
        :loading="loading"
        @retry="handleRetry"
        @back="handleBack"
      />

      <!-- 平台详情内容 -->
      <div v-else-if="platform" class="detail-content">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="平台ID">{{
            platform.id
          }}</el-descriptions-item>
          <el-descriptions-item label="平台名称">
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.name" />
            </template>
            <template v-else>
              {{ platform.name }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="操作系统">
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.os">
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
              <el-select v-model="safeEditForm.status">
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
              <el-input v-model="safeEditForm.version" />
            </template>
            <template v-else>
              {{ platform.version }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">{{
            platform.created_at
          }}</el-descriptions-item>
          <el-descriptions-item label="更新时间">{{
            platform.updated_at
          }}</el-descriptions-item>
        </el-descriptions>

        <div class="section-title">远程访问</div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="访问状态">
            <template v-if="platform.status === 'active'">
              <el-tag type="success">运行中</el-tag>
              <span class="access-status">可以远程访问平台</span>
            </template>
            <template v-else>
              <el-tag type="info">已停止</el-tag>
              <span class="access-status">需要先启动平台才能远程访问</span>
            </template>
          </el-descriptions-item>
          <el-descriptions-item
            label="访问链接"
            v-if="platform.status === 'active'"
          >
            <div class="access-link-container">
              <el-input
                v-model="accessLink"
                readonly
                placeholder="平台启动中，链接生成中..."
                class="access-link-input"
              >
                <template #append>
                  <el-button @click="copyAccessLink">复制</el-button>
                </template>
              </el-input>
              <el-button
                type="primary"
                @click="openAccessLink"
                :disabled="!accessLink"
              >
                <el-icon><Link /></el-icon>
                访问平台
              </el-button>
              <el-button
                type="warning"
                @click="refreshAccessLink"
                :disabled="!accessLink"
              >
                <el-icon><Refresh /></el-icon>
                刷新链接
              </el-button>
            </div>
            <div class="access-info">
              <el-alert
                title="链接有效期为24小时，平台停止后链接将失效"
                type="info"
                :closable="false"
                show-icon
              />
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">运行环境</div>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Docker镜像">
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.image" />
            </template>
            <template v-else>
              {{ platform.image }}
            </template>
          </el-descriptions-item>
        </el-descriptions>

        <el-descriptions :column="1" border>
          <el-descriptions-item label="启动路径">
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.bin" />
            </template>
            <template v-else>
              {{ platform.bin }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="资源文件">
            <template v-if="isEditing">
              <div
                v-for="(file, index) in safeEditForm.files"
                :key="file.id"
                class="file-item"
              >
                <el-input
                  v-model="file.type"
                  placeholder="文件类型"
                  style="width: 150px"
                />
                <el-input v-model="file.url" placeholder="文件URL" />
                <el-button type="danger" @click="removeFile(index)"
                  >删除</el-button
                >
              </div>
              <el-button type="primary" @click="addFile">添加文件</el-button>
            </template>
            <template v-else>
              <div
                v-for="file in platform.files"
                :key="file.id"
                class="file-item"
              >
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
              <div
                v-for="(feature, index) in safeEditForm.features"
                :key="index"
                class="feature-item"
              >
                <el-input v-model="safeEditForm.features[index]" />
                <el-button type="danger" @click="removeFeature(index)"
                  >删除</el-button
                >
              </div>
              <el-button type="primary" @click="addFeature">添加特性</el-button>
            </template>
            <template v-else>
              <div
                v-for="(feature, index) in platform.features"
                :key="index"
                class="feature-item"
              >
                {{ feature }}
              </div>
            </template>
          </el-descriptions-item>
        </el-descriptions>

        <div class="section-title">配置信息</div>
        <el-descriptions :column="1" border>
          <el-descriptions-item label="Wine版本" v-if="platform.config?.wine">
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.config.wine" />
            </template>
            <template v-else>
              {{ platform.config.wine }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="DXVK版本" v-if="platform.config?.dxvk">
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.config.dxvk" />
            </template>
            <template v-else>
              {{ platform.config.dxvk }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="VKD3D版本" v-if="platform.config?.vkd3d">
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.config.vkd3d" />
            </template>
            <template v-else>
              {{ platform.config.vkd3d }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item
            label="Python版本"
            v-if="platform.config?.python"
          >
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.config.python" />
            </template>
            <template v-else>
              {{ platform.config.python }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item
            label="Proton版本"
            v-if="platform.config?.proton"
          >
            <template v-if="isEditing">
              <el-input v-model="safeEditForm.config.proton" />
            </template>
            <template v-else>
              {{ platform.config.proton }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item
            label="着色器缓存"
            v-if="platform.config?.['shader-cache']"
          >
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config['shader-cache']">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config["shader-cache"] }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item
            label="远程游戏"
            v-if="platform.config?.['remote-play']"
          >
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config['remote-play']">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config["remote-play"] }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="广播" v-if="platform.config?.broadcast">
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config.broadcast">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.broadcast }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="运行模式" v-if="platform.config?.mode">
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config.mode">
                <el-option label="主机模式" value="docked" />
                <el-option label="掌机模式" value="handheld" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.mode }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item
            label="分辨率"
            v-if="platform.config?.resolution"
          >
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config.resolution">
                <el-option label="1080p" value="1080p" />
                <el-option label="720p" value="720p" />
                <el-option label="480p" value="480p" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.resolution }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="WiFi" v-if="platform.config?.wifi">
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config.wifi">
                <el-option label="启用" value="enabled" />
                <el-option label="禁用" value="disabled" />
              </el-select>
            </template>
            <template v-else>
              {{ platform.config.wifi }}
            </template>
          </el-descriptions-item>
          <el-descriptions-item label="蓝牙" v-if="platform.config?.bluetooth">
            <template v-if="isEditing">
              <el-select v-model="safeEditForm.config.bluetooth">
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
        <div>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="安装命令">
              <template v-if="isEditing">
                <div
                  v-for="(command, index) in safeEditForm.installer || []"
                  :key="index"
                  class="installer-item"
                >
                  <el-select
                    v-model="(command as any).type"
                    placeholder="选择命令类型"
                    @change="(val: CommandType) => handleCommandTypeChange(command, val)"
                  >
                    <el-option label="命令" value="command" />
                    <el-option label="移动" value="move" />
                    <el-option label="权限" value="chmodx" />
                    <el-option label="解压" value="extract" />
                  </el-select>
                  <template v-if="(command as any).type === 'command'">
                    <el-input
                      v-model="command.command"
                      placeholder="输入命令"
                      class="command-input"
                    />
                  </template>
                  <template v-if="(command as any).type === 'move'">
                    <div class="command-inputs">
                      <el-input
                        :model-value="command.move?.src || ''"
                        @update:model-value="(val: string) => { 
                          command.move = command.move || {src: '', dst: ''};
                          command.move.src = val;
                        }"
                        placeholder="源文件"
                      />
                      <el-input
                        :model-value="command.move?.dst || ''"
                        @update:model-value="(val: string) => { 
                          command.move = command.move || {src: '', dst: ''};
                          command.move.dst = val;
                        }"
                        placeholder="目标路径"
                      />
                    </div>
                  </template>
                  <template v-if="(command as any).type === 'chmodx'">
                    <el-input
                      v-model="command.chmodx"
                      placeholder="文件路径"
                      class="command-input"
                    />
                  </template>
                  <template v-if="(command as any).type === 'extract'">
                    <div class="command-inputs">
                      <el-input
                        :model-value="command.extract?.file || ''"
                        @update:model-value="(val: string) => { 
                          command.extract = command.extract || {file: '', dst: ''};
                          command.extract.file = val;
                        }"
                        placeholder="压缩文件"
                      />
                      <el-input
                        :model-value="command.extract?.dst || ''"
                        @update:model-value="(val: string) => { 
                          command.extract = command.extract || {file: '', dst: ''};
                          command.extract.dst = val;
                        }"
                        placeholder="解压目标"
                      />
                    </div>
                  </template>
                  <el-button
                    type="danger"
                    @click="removeInstallerCommand(index)"
                    >删除</el-button
                  >
                </div>
                <el-button type="primary" @click="addInstallerCommand"
                  >添加命令</el-button
                >
              </template>
              <template v-else>
                <div v-if="platform.installer && platform.installer.length > 0">
                  <div
                    v-for="(command, index) in platform.installer"
                    :key="index"
                    class="installer-item"
                  >
                    <template v-if="command.command">
                      <el-tag type="info">命令</el-tag>
                      <span class="command-text">{{ command.command }}</span>
                    </template>
                    <template v-if="command.move">
                      <el-tag type="success">移动</el-tag>
                      <span class="command-text"
                        >从 {{ command.move.src }} 到
                        {{ command.move.dst }}</span
                      >
                    </template>
                    <template v-if="command.chmodx">
                      <el-tag type="warning">权限</el-tag>
                      <span class="command-text"
                        >设置 {{ command.chmodx }} 可执行权限</span
                      >
                    </template>
                    <template v-if="command.extract">
                      <el-tag type="danger">解压</el-tag>
                      <span class="command-text"
                        >解压 {{ command.extract.file }} 到
                        {{ command.extract.dst }}</span
                      >
                    </template>
                  </div>
                </div>
                <div v-else class="empty-installer">
                  <el-empty description="暂无安装命令" />
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
import { ref, onMounted, computed } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import ApiErrorState from '@/components/ApiErrorState.vue';
import type {
  GamePlatform,
  GamePlatformFile,
  GamePlatformInstallerCommand,
} from "@/types";
import { Link, Refresh } from "@element-plus/icons-vue";
import {
  platformService,
  getPlatformDetail,
  updatePlatform,
  deletePlatform,
  getPlatformAccess,
  refreshPlatformAccess,
} from "@/services/platformService";

// 定义API错误类型
interface ApiError {
  title: string;
  message: string;
  detail?: string;
}

// 定义安装命令的类型
type CommandType = "command" | "move" | "chmodx" | "extract";
interface InstallerCommandWithType extends GamePlatformInstallerCommand {
  type: CommandType;
}

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const platform = ref<GamePlatform | null>(null);
const isEditing = ref(false);
const editForm = ref<GamePlatform | null>(null);
const error = ref<ApiError | null>(null);
const accessError = ref<ApiError | null>(null);
const accessLoading = ref(false);
const accessLink = ref("");

// 编辑中表单的计算属性，避免空值检查
const safeEditForm = computed(() => {
  if (isEditing.value && editForm.value) {
    return editForm.value;
  }
  
  // 如果不在编辑状态，使用 platform 的数据
  return platform.value || {
    id: "",
    name: "",
    type: "",
    version: "",
    image: "",
    bin: "",
    data: "",
    os: "Linux",
    status: "inactive",
    files: [],
    features: [],
    config: {},
    installer: [],
    created_at: "",
    updated_at: ""
  } as GamePlatform;
});

// 获取平台详情
const fetchPlatformDetail = async () => {
  loading.value = true;
  error.value = null;
  
  try {
    const id = route.params.id as string;
    console.log(`[平台详情] 正在获取平台数据: platformId=${id}`);
    platform.value = await getPlatformDetail(id);
    
    if (!platform.value) {
      error.value = {
        title: '未找到平台',
        message: '无法找到指定的平台数据',
        detail: `平台ID: ${id}`
      };
    }
  } catch (err: any) {
    console.error("[平台详情] 加载数据失败:", err);
    error.value = {
      title: '加载失败',
      message: '无法获取平台数据',
      detail: err.message || '未知错误'
    };
  } finally {
    loading.value = false;
  }
};

// 状态类型
const statusTypeMap: Record<string, string> = {
  active: "success",
  maintenance: "warning",
  inactive: "info",
};

const getStatusType = (status: string) => statusTypeMap[status] || "info";

// 状态文本
const statusTextMap: Record<string, string> = {
  active: "正常",
  maintenance: "维护中",
  inactive: "停用",
};

const getStatusText = (status: string) => statusTextMap[status] || status;

// 安装命令相关方法
// 确定命令类型
const getCommandType = (command: GamePlatformInstallerCommand): CommandType => {
  if (command.command !== undefined) return "command";
  if (command.move !== undefined) return "move";
  if (command.chmodx !== undefined) return "chmodx";
  if (command.extract !== undefined) return "extract";
  return "command"; // 默认类型
};

// 根据类型初始化命令
const initializeCommand = (type: CommandType): GamePlatformInstallerCommand => {
  switch (type) {
    case "command":
      return { command: "" };
    case "move":
      return { move: { src: "", dst: "" } };
    case "chmodx":
      return { chmodx: "" };
    case "extract":
      return { extract: { file: "", dst: "" } };
    default:
      return {};
  }
};

// 处理命令类型变更
const handleCommandTypeChange = (
  command: GamePlatformInstallerCommand & { type?: CommandType },
  newType: CommandType
) => {
  const commandWithType = command as InstallerCommandWithType;
  // 保存旧类型
  const oldType = commandWithType.type;
  // 如果类型相同，不做处理
  if (oldType === newType) return;

  // 清空旧的命令内容
  if (oldType === "command") delete command.command;
  if (oldType === "move") delete command.move;
  if (oldType === "chmodx") delete command.chmodx;
  if (oldType === "extract") delete command.extract;

  // 设置新类型
  commandWithType.type = newType;

  // 初始化新类型对应的值
  const newCommand = initializeCommand(newType);
  Object.assign(command, newCommand);
};

// 安装命令操作
const addInstallerCommand = () => {
  if (!editForm.value) return;
  editForm.value.installer = editForm.value.installer || [];

  // 创建一个新的命令对象，默认为命令类型
  const newCommand: InstallerCommandWithType = {
    type: "command",
    command: "",
  };

  editForm.value.installer.push(newCommand);
};

// 获取平台访问链接
const getAccessLink = async () => {
  if (!platform.value) return;
  
  if (platform.value.status === "active") {
    accessLoading.value = true;
    accessError.value = null;
    
    try {
      console.log(`[平台详情] 正在获取平台ID:${platform.value.id}的访问链接`);
      const link = await getPlatformAccess(platform.value.id);
      
      if (!link) {
        accessError.value = {
          title: '获取访问链接失败',
          message: '服务器返回空链接',
          detail: '可能是后端服务不可用'
        };
        return;
      }
      
      console.log(`[平台详情] 成功获取访问链接:`, { platformId: platform.value.id, link });
      accessLink.value = link;
    } catch (err: any) {
      console.error("[平台详情] 获取平台访问链接失败:", err);
      accessError.value = {
        title: '获取访问链接失败',
        message: err.response?.status === 404 ? '平台不存在或API路径错误' : '无法获取访问链接',
        detail: err.message || '未知错误'
      };
    } finally {
      accessLoading.value = false;
    }
  } else {
    console.log(`[平台详情] 平台未激活，不获取访问链接`);
    accessLink.value = "";
  }
};

// 复制访问链接
const copyAccessLink = () => {
  if (!accessLink.value) return;

  navigator.clipboard
    .writeText(accessLink.value)
    .then(() => {
      ElMessage.success("访问链接已复制到剪贴板");
    })
    .catch(() => {
      ElMessage.error("复制失败，请手动复制链接");
    });
};

// 打开访问链接
const openAccessLink = () => {
  if (!accessLink.value) return;
  window.open(accessLink.value, "_blank");
};

// 刷新访问链接
const refreshAccessLink = () => {
  ElMessageBox.confirm("刷新链接后，旧链接将立即失效，确定要刷新吗？", "提示", {
    type: "warning",
  })
    .then(async () => {
      if (!platform.value) return;
      try {
        const newLink = await refreshPlatformAccess(platform.value.id);
        accessLink.value = newLink;
        ElMessage.success("访问链接已刷新");
      } catch (error) {
        ElMessage.error("刷新链接失败");
        console.error("刷新链接失败:", error);
      }
    })
    .catch(() => {});
};

// 编辑平台
const handleEdit = () => {
  if (!platform.value) return;
  
  // 使用深拷贝确保不会直接修改原始数据
  const newEditForm = JSON.parse(JSON.stringify({
    ...platform.value,
    // 确保所有必要的字段都被初始化
    config: platform.value.config || {},
    files: platform.value.files || [],
    features: platform.value.features || [],
    installer: platform.value.installer || []
  }));

  // 为每个安装命令添加类型标识
  if (newEditForm.installer?.length) {
    newEditForm.installer = newEditForm.installer.map((command: GamePlatformInstallerCommand) => ({
      ...command,
      type: getCommandType(command)
    }));
  }

  editForm.value = newEditForm;
  isEditing.value = true;
};

// 取消编辑
const handleCancel = () => {
  isEditing.value = false;
  editForm.value = null;
};

// 保存编辑
const handleSave = async () => {
  if (!editForm.value || !platform.value) return;

  loading.value = true;
  try {
    const success = await updatePlatform(platform.value.id, editForm.value);
    if (success) {
      ElMessage.success("更新成功");
      // 重新获取平台数据
      await fetchPlatformDetail();
      isEditing.value = false;
      editForm.value = null;
    } else {
      ElMessage.error("更新失败");
    }
  } catch (error) {
    ElMessage.error("保存数据失败");
    console.error("保存数据失败:", error);
  } finally {
    loading.value = false;
  }
};

// 删除平台
const handleDelete = () => {
  ElMessageBox.confirm("确定要删除该平台吗？", "提示", {
    type: "warning",
  }).then(async () => {
    if (!platform.value) return;

    loading.value = true;
    try {
      const success = await deletePlatform(platform.value.id);
      if (success) {
        ElMessage.success("删除成功");
        router.push("/platform");
      } else {
        ElMessage.error("删除失败");
      }
    } catch (error) {
      ElMessage.error("删除失败");
      console.error("删除失败:", error);
    } finally {
      loading.value = false;
    }
  });
};

// 文件操作
const addFile = () => {
  if (!editForm.value) return;
  editForm.value.files.push({
    id: `file-${Date.now()}`,
    type: "",
    url: ""
  });
};

const removeFile = (index: number) => {
  if (!editForm.value) return;
  editForm.value.files.splice(index, 1);
};

// 特性操作
const addFeature = () => {
  if (!editForm.value) return;
  editForm.value.features.push("");
};

const removeFeature = (index: number) => {
  if (!editForm.value) return;
  editForm.value.features.splice(index, 1);
};

// 安装命令操作
const removeInstallerCommand = (index: number) => {
  if (!editForm.value?.installer) return;
  editForm.value.installer.splice(index, 1);
};

// 重试加载平台详情
const handleRetry = () => {
  fetchPlatformDetail();
};

// 返回平台列表
const handleBack = () => {
  router.push("/platform");
};

// 重试获取访问链接
const handleAccessRetry = () => {
  getAccessLink();
};

onMounted(() => {
  fetchPlatformDetail();
  // 获取平台访问链接
  setTimeout(() => {
    getAccessLink();
  }, 500);
});
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
  border-left: 4px solid #409eff;
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

.installer-item {
  display: flex;
  align-items: flex-start;
  margin-bottom: 12px;
  gap: 8px;
  flex-wrap: wrap;
}

.command-input {
  flex: 1;
  min-width: 250px;
}

.command-inputs {
  display: flex;
  flex-direction: column;
  flex: 1;
  gap: 8px;
  min-width: 250px;
}

.command-text {
  margin-left: 8px;
  color: #666;
  font-family: monospace;
  word-break: break-all;
}

.empty-installer {
  text-align: center;
  padding: 20px 0;
  color: #909399;
}

.access-status {
  margin-left: 10px;
  font-size: 14px;
  color: #606266;
}

.access-link-container {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.access-link-input {
  flex: 1;
  min-width: 300px;
}

.access-info {
  margin-top: 10px;
}
</style>
