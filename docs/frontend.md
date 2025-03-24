# 前端开发文档

## 当前开发状态

### 1. 项目初始化

- [x] 使用 Vite 创建 Vue3 + TypeScript 项目
- [x] 安装并配置核心依赖
  - Element Plus
  - Pinia
  - Vue Router
  - Axios

### 2. 项目结构

```text
frontend/
├── src/
│   ├── api/          # API 请求配置
│   ├── assets/       # 静态资源
│   ├── components/   # 公共组件
│   ├── router/       # 路由配置
│   ├── stores/       # 状态管理
│   ├── utils/        # 工具函数
│   ├── views/        # 页面组件
│   ├── App.vue       # 根组件
│   └── main.ts       # 入口文件
```

### 3. 已实现功能

- [x] 基础框架搭建
  - [x] 路由配置
  - [x] 状态管理
  - [x] API 请求封装
  - [x] 布局组件
- [x] 页面组件
  - [x] 登录页面
  - [x] 主布局
    - [x] 菜单组件
    - [x] 顶部导航
    - [x] 响应式布局
  - [x] 首页仪表盘
  - [x] 游戏节点管理
    - [x] 节点列表展示
    - [x] 节点添加/编辑
    - [x] 节点删除
    - [x] 分页功能
    - [x] 节点详情
    - [ ] 节点监控（开发中）
  - [x] 游戏平台管理
    - [x] 平台列表展示
    - [x] 平台配置
    - [x] 平台状态显示
    - [x] 平台详情
    - [ ] 平台状态详情（开发中）
  - [x] 游戏卡片管理
    - [x] 卡片列表展示
    - [x] 卡片添加/编辑
    - [x] 卡片删除
    - [x] 分页功能
    - [x] 卡片详情
  - [x] 游戏实例管理
    - [x] 实例列表展示
    - [x] 实例创建/编辑
    - [x] 实例删除
    - [x] 实例控制（启动/停止/重启）
    - [x] 分页功能
    - [x] 实例详情
    - [ ] 实例监控（开发中）
    - [ ] 实例配置（开发中）
- [x] 示例数据
  - [x] 游戏节点数据
  - [x] 游戏平台数据
  - [x] 游戏卡片数据
  - [x] 游戏实例数据
  - [x] 仪表盘数据

### 4. 详情页开发

详情页开发进度和规范请参考 [详情页开发文档](detail-pages.md)

### 5. 待开发功能

#### 5.1 页面组件

- [x] 游戏节点管理
  - [x] 节点列表
  - [x] 节点详情
  - [ ] 节点监控
  - [x] 节点配置
- [x] 游戏平台管理
  - [x] 平台列表
  - [x] 平台配置
  - [x] 平台状态
  - [x] 平台详情
  - [ ] 平台状态详情
- [x] 游戏卡片管理
  - [x] 卡片列表
  - [x] 卡片配置
  - [x] 卡片详情
- [x] 游戏实例管理
  - [x] 实例列表
  - [x] 实例控制
  - [x] 实例详情
  - [ ] 实例监控
  - [ ] 实例配置

#### 5.2 功能增强

- [ ] 用户认证
  - [ ] 登录状态管理
  - [ ] 路由守卫
  - [ ] 权限控制
- [ ] 数据交互
  - [ ] API 接口对接
  - [ ] 数据加载状态
  - [ ] 错误处理
- [ ] 用户体验
  - [ ] 主题配置
  - [ ] 国际化支持
  - [ ] 响应式适配
  - [ ] 动画效果

#### 5.3 开发规范

- [ ] 代码规范
  - [ ] ESLint 配置
  - [ ] Prettier 配置
  - [ ] TypeScript 类型定义
- [ ] 组件规范
  - [ ] 组件文档
  - [ ] 组件测试
  - [ ] 组件复用
- [ ] 构建优化
  - [ ] 打包优化
  - [ ] 性能优化
  - [ ] 缓存策略

### 6. UI 规范

#### 6.1 表格样式规范

所有管理页面的表格样式应遵循以下规范：

1. 表格基础样式

   - 移除 `border` 属性
   - 使用 `width: 100%` 样式
   - 使用 `v-loading` 指令显示加载状态

2. 操作按钮样式

   - 使用 `link` 类型的按钮
   - 移除 `el-button-group` 包裹
   - 按钮顺序：编辑/操作 -> 创建实例（如果有）-> 删除
   - 按钮颜色规范：
     - 编辑/操作：primary
     - 创建实例：success
     - 删除：danger
     - 其他操作：根据操作类型选择合适颜色

3. 分页样式

   - 使用 `pagination` 类名
   - 右对齐布局
   - 移除 `jumper` 功能
   - 分页大小选项：[10, 20, 50, 100]

4. 列宽规范

   - ID 列：120px
   - 名称列：150px
   - 状态列：100px
   - 时间列：180px
   - 操作列：200-300px（根据按钮数量调整）
   - 其他列：根据内容长度调整

5. 状态标签样式

   - 使用 `el-tag` 组件
   - 颜色规范：
     - 成功/在线/运行中：success
     - 警告/维护中：warning
     - 错误：danger
     - 其他：info

6. 容器样式
   - 使用 `el-card` 组件作为容器
   - 内边距：20px
   - 头部使用 `card-header` 类名
   - 标题和操作按钮两端对齐

#### 6.2 表单样式规范

1. 对话框

   - 宽度：500px
   - 标题：根据操作类型显示"添加/编辑 xxx"
   - 底部按钮：取消（默认）、确定（primary）

2. 表单项

   - 标签宽度：100px
   - 必填项使用红色星号标记
   - 验证规则统一使用 trigger: "blur"

3. 输入控件
   - 文本输入：使用 `el-input`
   - 数字输入：使用 `el-input-number`
   - 选择器：使用 `el-select`
   - 文本域：使用 `el-input` 的 textarea 类型

## 开发计划

### 第一阶段：基础功能完善

1. 完成所有页面组件的基础实现
2. 实现用户认证和权限控制
3. 对接后端 API 接口
4. 添加基础错误处理

### 第二阶段：功能增强

1. 添加主题配置功能
2. 实现国际化支持
3. 优化响应式布局
4. 添加动画效果

### 第三阶段：性能优化

1. 优化打包配置
2. 实现组件懒加载
3. 优化数据缓存
4. 添加性能监控

## 技术栈

- 核心框架：Vue 3
- 开发语言：TypeScript
- 构建工具：Vite
- UI 框架：Element Plus
- 状态管理：Pinia
- 路由管理：Vue Router
- HTTP 客户端：Axios

## 开发规范

### 1. 命名规范

- 组件名：PascalCase
- 文件名：kebab-case
- 变量名：camelCase
- 常量名：UPPER_CASE

### 2. 目录规范

- 页面组件放在 views 目录
- 公共组件放在 components 目录
- API 接口放在 api 目录
- 工具函数放在 utils 目录

### 3. 组件规范

- 使用 Composition API
- 使用 TypeScript 类型注解
- 组件属性使用 Props 类型定义
- 事件使用 Emit 类型定义

### 4. API 错误处理规范

#### 4.1 错误状态组件

所有涉及 API 调用的页面必须使用统一的错误状态组件（`ApiErrorState.vue`）来展示错误信息。错误状态组件应包含：

1. 错误标题
2. 错误信息
3. 详细说明（可选）
4. 操作按钮（重试、返回等）

```html
<api-error-state
  :title="error.title"
  :message="error.message"
  :detail="error.detail"
  :loading="loading"
  @retry="handleRetry"
  @back="handleBack"
/>
```

#### 4.2 错误数据结构

统一使用以下错误数据结构：

```typescript
interface ApiError {
  title: string; // 错误标题
  message: string; // 错误信息
  detail?: string; // 详细说明（可选）
}
```

#### 4.3 错误处理流程

1. **API 调用前：**

   - 清空现有错误状态
   - 设置加载状态

   ```typescript
   error.value = null;
   loading.value = true;
   ```

2. **API 调用时：**

   - 使用 try-catch 包裹所有 API 调用
   - 在 catch 中统一处理错误

   ```typescript
   try {
     const result = await apiCall();
     // 处理成功响应
   } catch (err: any) {
     error.value = {
       title: "操作失败",
       message: "无法完成请求",
       detail: err.message || "未知错误",
     };
   } finally {
     loading.value = false;
   }
   ```

3. **API 调用后：**

   - 验证响应数据完整性
   - 处理空值情况

   ```typescript
   if (!result) {
     error.value = {
       title: "数据异常",
       message: "返回数据为空",
       detail: "请检查API响应",
     };
     return;
   }
   ```

#### 4.4 错误展示规则

1. **页面级错误：**

   - 使用全屏错误状态组件
   - 提供返回和重试操作

   ```vue
   <template>
     <api-error-state
       v-if="error"
       v-bind="error"
       @retry="handleRetry"
       @back="handleBack"
     />
     <div v-else>
       <!-- 正常内容 -->
     </div>
   </template>
   ```

2. **局部功能错误：**

   - 在相应区域显示错误状态组件
   - 仅提供重试操作

   ```vue
   <template>
     <div class="section">
       <api-error-state
         v-if="sectionError"
         v-bind="sectionError"
         @retry="handleSectionRetry"
       >
         <template #actions>
           <el-button type="primary" @click="handleSectionRetry">
             重新加载
           </el-button>
         </template>
       </api-error-state>
       <div v-else>
         <!-- 区域内容 -->
       </div>
     </div>
   </template>
   ```

#### 4.5 错误提示规则

1. **标题文案规范：**

   - 404 错误：'未找到 xxx'
   - 网络错误：'网络连接失败'
   - 服务器错误：'服务器错误'
   - 权限错误：'无权访问'
   - 数据错误：'数据异常'

2. **消息文案规范：**

   - 简洁明了
   - 避免技术术语
   - 提供解决建议

3. **详情信息规范：**
   - 包含错误码
   - 包含具体错误信息
   - 适当包含调试信息

#### 4.6 错误恢复机制

1. **重试机制：**

   - 提供重试功能
   - 限制重试次数
   - 重试间隔递增

2. **降级处理：**
   - 提供备用数据
   - 降级展示方案
   - 离线模式支持

### 5. 样式规范

- 使用 SCSS 预处理器
- 使用 BEM 命名规范
- 组件样式使用 scoped
- 全局样式放在 assets 目录

## 注意事项

1. 代码提交前进行代码格式化
2. 保持组件单一职责
3. 合理使用 TypeScript 类型
4. 注意性能优化
5. 保持代码可维护性

## 数据源切换功能

为了方便前端开发和调试，系统实现了API数据和Mock数据的环境切换功能。数据源的选择在启动时完成，运行时不可更改。

### 架构设计

数据源切换功能基于以下几个核心组件：

1. **配置管理 (`config/index.ts`)**：
   - 通过环境变量控制是否使用Mock数据
   - 配置在应用启动时确定，运行时不可更改

2. **数据服务基类 (`services/dataService.ts`)**：
   - 提供统一的数据访问逻辑
   - 根据环境配置决定使用Mock数据还是真实API数据
   - 处理不同的API响应格式和错误情况

3. **业务服务层**：
   - 每种数据类型（节点、平台、卡片、实例）有独立的服务
   - 继承自数据服务基类，复用通用逻辑
   - 提供特定业务的数据处理逻辑

### 使用方法

#### 1. 在视图中使用服务

```typescript
import { getNodeList, deleteNode } from '@/services/nodeService';

// 获取数据列表
const fetchNodeList = async () => {
  try {
    const result = await getNodeList(params);
    // 处理结果...
  } catch (error) {
    // 处理错误...
  }
};
```

无需关心数据是来自Mock还是API，服务层会根据环境配置自动处理。

#### 2. 开发环境启动

使用提供的脚本启动开发环境：

```bash
# 使用API数据启动
bash ./scripts/dev.sh

# 使用Mock数据启动
bash ./scripts/dev.sh --mock
```

### 环境变量控制

通过环境变量控制数据源模式：

```env
# .env.local 或通过命令行设置
VITE_USE_MOCK=true      # 使用Mock数据
VITE_USE_MOCK=false     # 使用API数据
```

### 如何扩展

#### 添加新的服务

1. 创建新的服务类，继承自`DataService`
2. 实现对应的方法，分别处理Mock和API数据的情况
3. 导出服务实例和便捷方法

```typescript
import { DataService } from './dataService';
import { mockNewData } from '@/mocks/data/NewData';
import api from '@/api';

class NewDataService extends DataService {
  async getList(params?: any): Promise<{ list: any[], total: number }> {
    if (this.useMock()) {
      // 使用Mock数据
      await this.mockDelay();
      return { 
        list: mockNewData, 
        total: mockNewData.length 
      };
    } else {
      // 使用API数据
      try {
        const response = await api.newData.getList(params);
        return this.safelyExtractListData(response);
      } catch (error) {
        console.error('获取数据失败', error);
        return { list: [], total: 0 };
      }
    }
  }
  
  // 其他方法...
}

export const newDataService = new NewDataService();
export const getNewDataList = (params?: any) => newDataService.getList(params);
// 其他便捷方法导出...
```

### 优势

1. **开发效率**：在后端API尚未完成时，可使用Mock数据进行前端开发和测试
2. **代码整洁**：视图层代码无需关心数据来源，减少条件判断
3. **统一错误处理**：在服务层统一处理各种错误情况
4. **环境隔离**：开发环境和生产环境可以使用不同的数据源配置

### 注意事项

1. Mock数据应当尽可能模拟真实API的数据结构
2. 在生产环境部署前，确保环境变量设置为使用API数据源
3. 添加新的API端点时，应同步更新相应的Mock数据
