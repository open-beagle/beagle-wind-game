<template>
  <div class="node-detail-container">
    <el-card v-loading="loading" class="detail-card">
      <template #header>
        <div class="card-header">
          <span class="card-title">节点详情</span>
          <div class="header-actions">
            <template v-if="isEditing">
              <el-button type="primary" @click="handleSubmit">保存</el-button>
              <el-button @click="cancelEdit">取消</el-button>
            </template>
            <template v-else>
              <el-button type="primary" @click="startEdit">编辑</el-button>
              <el-button type="danger" @click="handleDelete">删除</el-button>
            </template>
          </div>
        </div>
      </template>

      <div v-if="node" class="detail-content">
        <!-- 基本信息 -->
        <div class="detail-section">
          <div class="section-header">
            <h3>基本信息</h3>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">节点ID</div>
              <div class="item-value">{{ node.id }}</div>
            </div>
            <div class="detail-item">
              <div class="item-label">节点名称</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.alias"
                  placeholder="请输入节点名称"
                />
                <span v-else>{{ node.alias }}</span>
              </div>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">型号</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.model"
                  placeholder="请输入节点型号"
                />
                <span v-else>{{ node.model || "-" }}</span>
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">类型</div>
              <div class="item-value">
                <el-select
                  v-if="isEditing"
                  v-model="form.type"
                  placeholder="请选择节点类型"
                >
                  <el-option label="物理节点" :value="GameNodeType.Physical" />
                  <el-option label="虚拟节点" :value="GameNodeType.Virtual" />
                  <el-option label="容器节点" :value="GameNodeType.Container" />
                </el-select>
                <span v-else>{{ node.type }}</span>
              </div>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">区域</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.location"
                  placeholder="请输入区域"
                />
                <span v-else>{{ node.location || "-" }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 硬件配置 -->
        <div class="detail-section">
          <div class="section-header">
            <h3>硬件配置</h3>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">CPU</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.hardware.CPU"
                  placeholder="请输入CPU信息"
                />
                <span v-else>{{ node.hardware?.CPU || "-" }}</span>
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">内存</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.hardware.RAM"
                  placeholder="请输入内存信息"
                />
                <span v-else>{{ node.hardware?.RAM || "-" }}</span>
              </div>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">GPU</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.hardware.GPU"
                  placeholder="请输入GPU信息"
                />
                <span v-else>{{ node.hardware?.GPU || "-" }}</span>
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">存储</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.hardware.Storage"
                  placeholder="请输入存储信息"
                />
                <span v-else>{{ node.hardware?.Storage || "-" }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 系统配置 -->
        <div class="detail-section">
          <div class="section-header">
            <h3>系统配置</h3>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">操作系统类型</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.system.os_type"
                  placeholder="请输入操作系统类型"
                />
                <span v-else>{{ node.system?.os_type || "-" }}</span>
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">操作系统版本</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.system.os_version"
                  placeholder="请输入操作系统版本"
                />
                <span v-else>{{ node.system?.os_version || "-" }}</span>
              </div>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">GPU驱动版本</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.system.gpu_driver"
                  placeholder="请输入GPU驱动版本"
                />
                <span v-else>{{ node.system?.gpu_driver || "-" }}</span>
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">CUDA版本</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.system.cuda_version"
                  placeholder="请输入CUDA版本"
                />
                <span v-else>{{ node.system?.cuda_version || "-" }}</span>
              </div>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">IP地址</div>
              <div class="item-value">
                <el-input
                  v-if="isEditing"
                  v-model="form.system.ip_address"
                  placeholder="请输入IP地址"
                />
                <span v-else>{{ node.system?.ip_address || "-" }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- 标签 -->
        <div class="detail-section">
          <div class="section-header">
            <h3>标签</h3>
          </div>
          <div class="detail-row">
            <div class="detail-item full-width">
              <div class="item-value">
                <div class="tags-container">
                  <template v-if="isEditing">
                    <div class="tags-edit">
                      <div
                        v-for="(value, key) in form.labels"
                        :key="key"
                        class="tag-item-edit"
                      >
                        <el-input
                          v-model="labelKeys[key]"
                          placeholder="键"
                          class="tag-input"
                          @change="updateLabels"
                        >
                          <template #prepend>键</template>
                        </el-input>
                        <el-input
                          v-model="form.labels[key]"
                          placeholder="值"
                          class="tag-input"
                        >
                          <template #prepend>值</template>
                        </el-input>
                        <el-button
                          type="danger"
                          circle
                          @click="removeLabel(key)"
                        >
                          <el-icon><Delete /></el-icon>
                        </el-button>
                      </div>
                    </div>
                    <el-button type="primary" @click="addLabel"
                      >添加标签</el-button
                    >
                  </template>
                  <template v-else>
                    <template
                      v-if="node.labels && Object.keys(node.labels).length > 0"
                    >
                      <el-tag
                        v-for="(value, key) in node.labels"
                        :key="key"
                        class="tag-item"
                        size="small"
                      >
                        {{ key }}: {{ value }}
                      </el-tag>
                    </template>
                    <span v-else class="no-tags">无标签信息</span>
                  </template>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 时间信息 -->
        <div class="detail-section">
          <div class="section-header">
            <h3>时间信息</h3>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">创建时间</div>
              <div class="item-value">{{ formatTime(node.createdAt) }}</div>
            </div>
            <div class="detail-item">
              <div class="item-label">更新时间</div>
              <div class="item-value">{{ formatTime(node.updatedAt) }}</div>
            </div>
          </div>
        </div>

        <!-- 运行状态 -->
        <div class="detail-section" v-if="!isEditing">
          <div class="section-header">
            <h3>运行状态</h3>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">状态</div>
              <div class="item-value">
                <el-tag
                  :type="getStatusType(node.status?.state || '')"
                  class="status-tag"
                >
                  {{ getStatusText(node.status?.state || "") }}
                </el-tag>
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">在线状态</div>
              <div class="item-value">
                <el-tag
                  :type="node.status?.online ? 'success' : 'info'"
                  class="status-tag"
                >
                  {{ node.status?.online ? "在线" : "离线" }}
                </el-tag>
              </div>
            </div>
          </div>
          <div class="detail-row">
            <div class="detail-item">
              <div class="item-label">最后在线时间</div>
              <div class="item-value">
                {{ formatTime(node.status?.lastOnline) }}
              </div>
            </div>
            <div class="detail-item">
              <div class="item-label">状态更新时间</div>
              <div class="item-value">
                {{ formatTime(node.status?.updatedAt) }}
              </div>
            </div>
          </div>

          <!-- 硬件资源 -->
          <div class="section-divider">
            <div class="section-title">硬件资源</div>
          </div>
          <div v-if="node?.status?.resource?.hardware" class="resource-section">
            <!-- CPU信息 -->
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">CPU型号</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.cpu.model }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">CPU核心数</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.cpu.cores }} 核
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">CPU线程数</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.cpu.threads }} 线程
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">CPU频率</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.cpu.frequency }} GHz
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">CPU温度</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.cpu.temperature }}°C
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">CPU使用率</div>
                <div class="item-value">
                  <div class="progress-container">
                    <el-progress
                      :percentage="node.status.resource.hardware.cpu.usage"
                      :color="
                        getResourceColor(
                          node.status.resource.hardware.cpu.usage,
                          100
                        )
                      "
                      :stroke-width="16"
                      :text-inside="true"
                    >
                      <template #default>
                        <span
                          >{{ node.status.resource.hardware.cpu.usage }}%</span
                        >
                      </template>
                    </el-progress>
                  </div>
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">CPU缓存</div>
                <div class="item-value">
                  {{ formatStorage(node.status.resource.hardware.cpu.cache) }}
                </div>
              </div>
            </div>

            <!-- 内存信息 -->
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">内存类型</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.memory.type }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">内存频率</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.memory.frequency }} MHz
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">内存通道数</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.memory.channels }} 通道
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">内存使用率</div>
                <div class="item-value">
                  <div class="progress-container">
                    <el-progress
                      :percentage="node.status.resource.hardware.memory.usage"
                      :color="
                        getResourceColor(
                          node.status.resource.hardware.memory.usage,
                          100
                        )
                      "
                      :stroke-width="16"
                      :text-inside="true"
                    >
                      <template #default>
                        <span
                          >{{
                            formatMemory(
                              node.status.resource.hardware.memory.used
                            )
                          }}/{{
                            formatMemory(
                              node.status.resource.hardware.memory.total
                            )
                          }}</span
                        >
                      </template>
                    </el-progress>
                  </div>
                </div>
              </div>
            </div>

            <!-- GPU信息 -->
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">GPU型号</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.gpu.model }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">CUDA核心数</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.gpu.cudaCores }} 核
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">GPU温度</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.gpu.temperature }}°C
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">GPU功耗</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.gpu.power }}W
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">GPU使用率</div>
                <div class="item-value">
                  <div class="progress-container">
                    <el-progress
                      :percentage="node.status.resource.hardware.gpu.usage"
                      :color="
                        getResourceColor(
                          node.status.resource.hardware.gpu.usage,
                          100
                        )
                      "
                      :stroke-width="16"
                      :text-inside="true"
                    >
                      <template #default>
                        <span
                          >{{ node.status.resource.hardware.gpu.usage }}%</span
                        >
                      </template>
                    </el-progress>
                  </div>
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">显存使用率</div>
                <div class="item-value">
                  <div class="progress-container">
                    <el-progress
                      :percentage="
                        node.status.resource.hardware.gpu.memoryUsage
                      "
                      :color="
                        getResourceColor(
                          node.status.resource.hardware.gpu.memoryUsage,
                          100
                        )
                      "
                      :stroke-width="16"
                      :text-inside="true"
                    >
                      <template #default>
                        <span
                          >{{
                            formatMemory(
                              node.status.resource.hardware.gpu.memoryUsed
                            )
                          }}/{{
                            formatMemory(
                              node.status.resource.hardware.gpu.memoryTotal
                            )
                          }}</span
                        >
                      </template>
                    </el-progress>
                  </div>
                </div>
              </div>
            </div>

            <!-- 磁盘信息 -->
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">磁盘型号</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.disk.model }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">磁盘类型</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.disk.type }}
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">接口类型</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.disk.interface }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">磁盘使用率</div>
                <div class="item-value">
                  <div class="progress-container">
                    <el-progress
                      :percentage="node.status.resource.hardware.disk.usage"
                      :color="
                        getResourceColor(
                          node.status.resource.hardware.disk.usage,
                          100
                        )
                      "
                      :stroke-width="16"
                      :text-inside="true"
                    >
                      <template #default>
                        <span
                          >{{
                            formatStorage(
                              node.status.resource.hardware.disk.used
                            )
                          }}/{{
                            formatStorage(
                              node.status.resource.hardware.disk.capacity
                            )
                          }}</span
                        >
                      </template>
                    </el-progress>
                  </div>
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">读取速度</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.disk.readSpeed }} MB/s
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">写入速度</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.disk.writeSpeed }} MB/s
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">IOPS</div>
                <div class="item-value">
                  {{ node.status.resource.hardware.disk.iops }}
                </div>
              </div>
            </div>
          </div>

          <!-- 软件信息 -->
          <div class="section-divider">
            <div class="section-title">软件信息</div>
          </div>
          <div v-if="node?.status?.resource?.software" class="software-section">
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">操作系统</div>
                <div class="item-value">
                  {{ node.status.resource.software.osDistribution }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">系统版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.osVersion }}
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">系统架构</div>
                <div class="item-value">
                  {{ node.status.resource.software.osArchitecture }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">内核版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.kernelVersion }}
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">GPU驱动版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.gpuDriverVersion }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">CUDA版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.cudaVersion }}
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">Docker版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.dockerVersion }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">Containerd版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.containerdVersion }}
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">Runc版本</div>
                <div class="item-value">
                  {{ node.status.resource.software.runcVersion }}
                </div>
              </div>
            </div>
          </div>

          <!-- 网络信息 -->
          <div class="section-divider">
            <div class="section-title">网络信息</div>
          </div>
          <div v-if="node?.status?.resource?.network" class="network-section">
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">网络带宽</div>
                <div class="item-value">
                  {{ node.status.resource.network.bandwidth }} Mbps
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">网络延迟</div>
                <div class="item-value">
                  {{ node.status.resource.network.latency }} ms
                </div>
              </div>
            </div>
            <div class="detail-row">
              <div class="detail-item">
                <div class="item-label">连接数</div>
                <div class="item-value">
                  {{ node.status.resource.network.connections }}
                </div>
              </div>
              <div class="detail-item">
                <div class="item-label">丢包率</div>
                <div class="item-value">
                  {{ node.status.resource.network.packetLoss }}%
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import type { FormInstance, FormRules } from "element-plus";
import {
  GameNode,
  GameNodeType,
  GameNodeState,
  GPUInfo,
  Metric,
} from "@/types/GameNode";
import { getNodeDetail, updateNode, deleteNode } from "@/services/nodeService";

const route = useRoute();
const router = useRouter();
const loading = ref(false);
const node = ref<GameNode | null>(null);
const isEditing = ref(false);

interface EditableGameNode {
  id: string;
  alias: string;
  model: string;
  type: GameNodeType;
  location: string;
  labels: Record<string, string>;
  hardware: {
    CPU: string;
    RAM: string;
    GPU: string;
    Storage: string;
  };
  system: {
    os_type: string;
    os_version: string;
    gpu_driver: string;
    cuda_version: string;
    ip_address: string;
  };
}

const defaultForm: EditableGameNode = {
  id: "",
  alias: "",
  model: "",
  type: GameNodeType.Physical,
  location: "",
  labels: {},
  hardware: {
    CPU: "",
    RAM: "",
    GPU: "",
    Storage: "",
  },
  system: {
    os_type: "",
    os_version: "",
    gpu_driver: "",
    cuda_version: "",
    ip_address: "",
  },
};

const form = ref<EditableGameNode>(defaultForm);
const originalForm = ref<EditableGameNode>(defaultForm);

// 添加新的响应式变量用于存储标签的键
const labelKeys = ref<Record<string, string>>({});

const rules: FormRules = {
  alias: [
    { required: true, message: "请输入节点名称", trigger: "blur" },
    { min: 2, max: 50, message: "长度在 2 到 50 个字符", trigger: "blur" },
  ],
  type: [{ required: true, message: "请选择节点类型", trigger: "change" }],
  model: [{ required: true, message: "请输入节点型号", trigger: "blur" }],
  location: [{ required: true, message: "请输入位置信息", trigger: "blur" }],
};

// 获取节点详情
const fetchNodeDetail = async () => {
  loading.value = true;
  try {
    const id = route.params.id as string;
    const result = await getNodeDetail(id);

    if (result) {
      node.value = result;
      form.value = {
        id: result.id,
        alias: result.alias,
        model: result.model,
        type: result.type,
        location: result.location,
        labels: result.labels || {},
        hardware: {
          CPU: result.hardware?.CPU || "",
          RAM: result.hardware?.RAM || "",
          GPU: result.hardware?.GPU || "",
          Storage: result.hardware?.Storage || "",
        },
        system: {
          os_type: result.system?.os_type || "",
          os_version: result.system?.os_version || "",
          gpu_driver: result.system?.gpu_driver || "",
          cuda_version: result.system?.cuda_version || "",
          ip_address: result.system?.ip_address || "",
        },
      };
      originalForm.value = JSON.parse(JSON.stringify(form.value));
    } else {
      ElMessage.error("未找到节点数据");
      node.value = null;
    }
  } catch (error) {
    ElMessage.error("加载数据失败");
    console.error("加载数据失败:", error);
  } finally {
    loading.value = false;
  }
};

// 格式化时间显示
const formatTime = (time?: string): string => {
  if (!time) return "-";
  return new Date(time).toLocaleString();
};

// 格式化内存
const formatMemory = (bytes?: number): string => {
  if (!bytes) return "-";
  const gb = bytes / (1024 * 1024 * 1024);
  return `${gb.toFixed(2)}GB`;
};

// 格式化存储
const formatStorage = (bytes?: number): string => {
  if (!bytes) return "-";
  const gb = bytes / (1024 * 1024 * 1024);
  return `${gb.toFixed(2)}GB`;
};

// 资源使用率颜色
const getResourceColor = (value: number, max: number): string => {
  const percentage = (value / max) * 100;
  if (percentage < 60) return "#67C23A";
  if (percentage < 80) return "#E6A23C";
  return "#F56C6C";
};

// 状态类型
const getStatusType = (state: GameNodeState): string => {
  switch (state) {
    case GameNodeState.Online:
      return "success";
    case GameNodeState.Offline:
      return "info";
    case GameNodeState.Maintenance:
      return "warning";
    case GameNodeState.Ready:
      return "success";
    case GameNodeState.Busy:
      return "warning";
    case GameNodeState.Error:
      return "danger";
    default:
      return "info";
  }
};

const getStatusText = (state: GameNodeState): string => {
  switch (state) {
    case GameNodeState.Online:
      return "在线";
    case GameNodeState.Offline:
      return "离线";
    case GameNodeState.Maintenance:
      return "维护中";
    case GameNodeState.Ready:
      return "就绪";
    case GameNodeState.Busy:
      return "忙碌";
    case GameNodeState.Error:
      return "错误";
    default:
      return "未知";
  }
};

// 开始编辑
const startEdit = () => {
  isEditing.value = true;
  originalForm.value = JSON.parse(JSON.stringify(form.value));
  labelKeys.value = {};
  Object.keys(form.value.labels).forEach((key) => {
    labelKeys.value[key] = key;
  });
};

// 取消编辑
const cancelEdit = () => {
  isEditing.value = false;
  form.value = JSON.parse(JSON.stringify(originalForm.value));
};

// 提交表单
const handleSubmit = async () => {
  try {
    const success = await updateNode(form.value.id, {
      alias: form.value.alias,
      model: form.value.model,
      type: form.value.type,
      location: form.value.location,
      labels: form.value.labels,
      hardware: form.value.hardware,
      system: form.value.system,
    });

    if (success) {
      ElMessage.success("更新成功");
      isEditing.value = false;
      await fetchNodeDetail();
    } else {
      ElMessage.error("更新失败");
    }
  } catch (error) {
    ElMessage.error("更新失败");
    console.error("更新失败:", error);
  }
};

// 删除节点
const handleDelete = () => {
  if (!node.value) return;

  ElMessageBox.confirm(
    `确定要删除节点 "${node.value.alias}" 吗？此操作不可恢复`,
    "删除确认",
    {
      confirmButtonText: "确定",
      cancelButtonText: "取消",
      type: "warning",
    }
  )
    .then(async () => {
      try {
        const success = await deleteNode(node.value?.id || "");
        if (success) {
          ElMessage.success("删除成功");
          router.push("/node");
        } else {
          ElMessage.error("删除失败");
        }
      } catch (error) {
        ElMessage.error("删除失败");
        console.error("删除失败:", error);
      }
    })
    .catch(() => {
      // 用户取消删除
    });
};

// 添加新标签
const addLabel = () => {
  const newKey = `key${Object.keys(form.value.labels).length + 1}`;
  form.value.labels[newKey] = "";
  labelKeys.value[newKey] = newKey;
};

// 删除标签
const removeLabel = (key: string) => {
  delete form.value.labels[key];
  delete labelKeys.value[key];
};

// 更新标签键
const updateLabels = () => {
  const newLabels: Record<string, string> = {};
  Object.entries(labelKeys.value).forEach(([oldKey, newKey]) => {
    if (form.value.labels[oldKey] !== undefined) {
      newLabels[newKey] = form.value.labels[oldKey];
    }
  });
  form.value.labels = newLabels;
};

onMounted(() => {
  fetchNodeDetail();
});
</script>

<style scoped>
.node-detail-container {
  padding: 12px;
  height: 100%;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: auto;
}

.detail-card {
  background-color: #fff;
  height: auto;
  min-height: min-content;
}

:deep(.el-card__body) {
  padding: 2px;
  height: auto;
  overflow: visible;
}

.detail-content {
  padding: 0;
  height: auto;
  overflow: visible;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 2px;
  border-bottom: none;
  height: 24px;
}

.card-title {
  font-size: 14px;
  font-weight: bold;
  color: #303133;
  margin: 0;
  line-height: 1;
}

.header-actions {
  display: flex;
  gap: 6px;
}

:deep(.el-button) {
  padding: 4px 12px;
  height: 24px;
  line-height: 1;
}

.detail-section {
  margin-top: 6px;
  padding: 6px;
  background-color: #fff;
  border-radius: 4px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 4px;
}

.section-header h3 {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin: 0;
}

.detail-row {
  display: flex;
  border-bottom: 1px solid #ebeef5;
  background-color: #fff;
  margin-bottom: 4px;
}

.detail-row:last-child {
  border-bottom: none;
  margin-bottom: 0;
}

.detail-item {
  display: flex;
  width: 50%;
  min-height: 28px;
}

.item-label {
  width: 120px;
  min-width: 120px;
  padding: 6px 10px;
  background-color: #f5f7fa;
  border-right: 1px solid #ebeef5;
  color: #606266;
  text-align: left;
  font-size: 14px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
}

.item-value {
  flex: 1;
  padding: 6px 10px;
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #303133;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.label-tag {
  margin: 2px;
}

.no-tags {
  color: #909399;
  font-style: italic;
}

.progress-container {
  width: 100%;
  margin: 0;
}

/* 进度条样式 */
:deep(.el-progress) {
  margin-bottom: 0;
  width: 100%;
}

:deep(.el-progress-bar__outer) {
  height: 16px !important;
  border-radius: 4px;
}

:deep(.el-progress-bar__inner) {
  border-radius: 4px;
}

:deep(.el-progress-bar__innerText) {
  font-size: 12px;
  line-height: 16px;
  color: #fff;
}

/* 标签样式 */
:deep(.el-tag) {
  display: inline-flex;
  align-items: center;
  height: 24px;
  font-size: 12px;
  padding: 0 8px;
}

.status-tag {
  min-width: 60px;
  text-align: center;
  justify-content: center;
}

/* IP地址样式 */
.ip-info {
  display: flex;
  align-items: center;
  gap: 5px;
}

.no-ip {
  color: #909399;
  font-style: italic;
}

/* 资源部分和指标部分样式 */
.resource-section,
.metrics-section {
  width: 100%;
}

.resource-section .detail-row,
.metrics-section .detail-row {
  width: 100%;
}

.resource-section .detail-item,
.metrics-section .detail-item {
  width: 100%;
}

.resource-section .item-value,
.metrics-section .item-value {
  flex: 1;
}

.section-divider {
  margin-top: 6px;
  margin-bottom: 4px;
}

.section-title {
  position: relative;
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  border-left: 4px solid #409eff;
  padding-left: 6px;
  line-height: 20px;
  margin: 0;
}

.system-form {
  max-width: 400px;
}

:deep(.el-input) {
  width: 100%;
}

:deep(.el-select) {
  width: 100%;
}

:deep(.el-textarea) {
  width: 100%;
}

.tags-edit {
  display: flex;
  flex-direction: column;
  gap: 8px;
  padding: 4px 0;
}

.tag-item-edit {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tag-input {
  width: 200px;
}

:deep(.el-input-group__prepend) {
  width: 35px;
  padding: 0 8px;
}

.tag-item {
  margin: 2px 4px 2px 0;
}
</style>
