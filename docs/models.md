# 数据模型设计

## 设计要点

### 1. 游戏节点（GameNode）

- 定位：平台的核心资源管理单元
- 职责：为游戏实例提供运行环境
- 管理：支持平台级别的节点管理
- 特性：支持动态加入和删除节点
- 关联：与游戏实例是一对多关系

### 2. 游戏卡片（GameCard）

- 定位：游戏数据库的核心实体
- 标识：
  - 游戏卡片 ID：平台主键，格式为"平台 ID-发行日期"（如"lutris-20200917"）
  - 平台 ID：外键，关联到特定游戏平台
  - Slug 名：命令行标识符，用于启动游戏（如 Lutris 平台）
    - 示例：`hades`（哈迪斯）、`EldenRing`（艾尔登法环）、`MetalSlug3`（合金弹头 3）
    - 用于生成启动命令：`lutris lutris:rungame/<slug_name>`
- 约束：
  - 一个游戏只能属于一个平台
  - 游戏卡片 ID 必须遵循"平台 ID-发行日期"的格式
  - 发行日期格式为"YYYYMMDD"
- 关系：与游戏平台是多对一关系

### 2.1 用户游戏状态（UserGameStatus）

- 定位：记录用户对特定游戏的运行状态
- 关联：
  - 用户 ID：关联到用户表
  - 游戏卡片 ID：关联到游戏卡片表
- 状态信息：
  - 最后运行时间：用户上次启动游戏的时间
  - 累计运行时长：用户玩该游戏的总时长（分钟）
  - 运行次数：用户启动该游戏的次数
  - 最后存档时间：用户最后一次保存游戏的时间
  - 游戏进度：用户当前游戏进度（可选）
- 统计信息：
  - 平均每次运行时长
  - 最近一周运行频率
  - 游戏完成度（如果有）
- 时间信息：
  - 创建时间：首次运行时间
  - 更新时间：最后状态更新时间

## 核心模型

### 1. GameNode（游戏节点）

- 基本信息：ID、名称、型号
- 硬件信息：硬件配置、网络状态、地理位置
- 状态信息：运行状态、资源使用、在线状态
- 时间信息：创建时间、更新时间、最后在线时间
- 管理信息：节点组、标签、备注

### 2. GameCard（游戏卡片）

- 标识信息：
  - ID：游戏卡片 ID（平台主键，格式：平台 ID-发行日期）
  - PlatformID：平台 ID（外键）
  - SlugName：命令行标识符
- 基本信息：名称、排序名
- 游戏信息：类型、描述、封面、分类、发行日期、标签
- 资源信息：游戏文件、更新包、补丁包
- 配置信息：运行参数、游戏设置、权限控制
- 时间信息：创建时间、更新时间

### 3. GamePlatform（游戏平台）

- 定位：系统运行的静态数据，存储在 `config` 目录中
- 设计原因：
  - 静态性：
    - 平台信息（如 Lutris、Steam、Switch）是相对固定的
    - 平台的版本、类型、特性等配置变化不频繁
    - 平台的安装步骤和配置规则是预定义的
  - 配置性：
    - 平台需要特定的容器镜像（`image`）
    - 需要特定的启动路径（`bin`）
    - 需要特定的数据目录（`data`）
    - 需要特定的文件系统（`files`）
  - 可维护性：
    - 集中管理所有平台配置
    - 便于版本控制和更新
    - 便于平台特性的统一管理
  - 运行时依赖：
    - 游戏卡片（GameCard）需要引用平台配置
    - 游戏实例（GameInstance）需要平台配置来创建运行环境
    - 但平台本身不需要在运行时动态修改
- 优势：
  - 配置集中：所有平台配置统一管理
  - 版本可控：平台配置的变更可以追踪
  - 部署简单：平台配置随系统一起部署
  - 维护方便：修改平台配置不需要数据库操作
- 基本信息：ID、名称、版本、类型
- 运行信息：容器镜像、启动路径、数据目录
- 功能信息：平台特性、平台配置
- 文件信息：平台文件（AppImage、密钥、固件）
- 游戏配置：可执行文件
- 安装步骤：移动文件、添加权限、解压文件、执行命令
- 时间信息：创建时间、更新时间

### 数据同步规范

#### 1. 数据源定义

1. 配置数据源（`config` 目录）：

   - 存储系统静态配置
   - 包含平台、游戏卡片等基础数据
   - 使用 YAML 格式
   - 作为数据源和版本控制

2. 示例数据源（`data` 目录）：

   - 存储示例数据
   - 用于开发和测试
   - 使用 YAML 格式
   - 作为数据模板

3. Mock 数据（`frontend/src/mocks/data` 目录）：
   - 前端开发使用的模拟数据
   - 使用 TypeScript 格式
   - 自动从配置同步生成
   - 用于前端开发和测试

#### 2. 同步规则

1. 配置到 Mock 的同步：

   - 使用 `sync-mocks.ts` 脚本
   - 自动生成 TypeScript 类型定义
   - 保持数据结构一致性
   - 支持数据验证

2. Mock 到配置的同步：

   - 使用 `sync-mocks.ts` 脚本
   - 自动生成 YAML 配置
   - 保持数据格式规范
   - 支持数据验证

3. 数据验证规则：
   - 类型检查
   - 必填字段验证
   - 格式规范验证
   - 关联关系验证

#### 3. 同步流程

1. 开发流程：

   - 在 `config` 或 `data` 目录修改数据
   - 运行同步脚本更新 mock 数据
   - 前端开发使用 mock 数据
   - 提交代码时包含所有变更

2. 测试流程：

   - 使用 mock 数据进行单元测试
   - 使用示例数据进行集成测试
   - 验证数据同步的正确性
   - 确保数据一致性

3. 部署流程：
   - 构建时使用配置数据
   - 验证数据完整性
   - 确保数据格式正确
   - 保持数据版本一致

#### 4. 目录结构

```text
project/
├── config/                 # 配置数据源
│   ├── platforms.yaml     # 平台配置
│   └── cards.yaml        # 游戏卡片配置
├── data/                  # 示例数据源
│   ├── platforms/        # 平台示例
│   └── cards/           # 游戏卡片示例
└── frontend/
    └── src/
        └── mocks/
            └── data/     # Mock 数据
                ├── GamePlatform.ts
                └── GameCard.ts
```

#### 5. 注意事项

1. 数据一致性：

   - 保持各数据源格式一致
   - 确保数据同步及时
   - 避免手动修改 mock 数据
   - 定期验证数据完整性

2. 版本控制：

   - 配置数据需要版本控制
   - 记录数据变更历史
   - 支持数据回滚
   - 保持版本同步

3. 开发规范：

   - 遵循数据格式规范
   - 使用同步脚本更新数据
   - 保持代码整洁
   - 编写必要的注释

4. 测试要求：
   - 编写数据验证测试
   - 测试同步功能
   - 验证数据一致性
   - 确保测试覆盖

### 4. GameInstance（游戏实例）

- 关联信息：游戏节点 ID、平台 ID、游戏卡片 ID
- 状态信息：运行状态、资源占用、性能指标
- 数据信息：存档数据、实例配置、备份数据
- 时间信息：创建时间、更新时间、启动时间、停止时间

## 设计特点

1. 数据分离

   - 游戏卡片和游戏节点完全独立
   - 通过游戏实例建立运行时关联
   - 平台配置独立管理

2. 配置灵活

   - 使用 JSON/YAML 存储复杂配置
   - 支持平台特定的安装步骤
   - 可扩展的文件系统设计

3. 状态追踪

   - 完整的生命周期记录
   - 详细的资源使用监控
   - 完整的备份机制

4. 权限控制

   - 平台级别的权限管理
   - 游戏级别的权限控制
   - 实例级别的访问控制

## 数据关系

1. 静态关系

   - GameCard 和 GamePlatform 通过 platform_id 强关联
   - GameNode 由平台统一管理
   - GameCard 必须属于一个 GamePlatform

2. 动态关系

   - GameInstance 运行时关联 GameNode、GamePlatform 和 GameCard
   - 实例可以独立于游戏卡片存在
   - 节点可以动态加入或退出平台

## 设计意图

1. 模块化设计

   - 每个模型职责单一
   - 接口清晰，易于扩展
   - 配置灵活，易于定制

2. 数据隔离

   - 游戏卡片与游戏节点解耦
   - 平台配置独立管理
   - 实例数据独立存储

3. 状态管理

   - 完整的生命周期追踪
   - 详细的资源监控
   - 灵活的配置管理

4. 扩展性

   - 支持新的平台类型
   - 支持新的游戏类型
   - 支持新的功能特性

## 数据表设计

### 游戏平台用户状态表 (platform_user_states)

```sql
CREATE TABLE platform_user_states (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,           -- 用户ID
    platform_id TEXT NOT NULL,          -- 平台ID
    state_type TEXT NOT NULL,           -- 状态类型：system/user
    state_data TEXT NOT NULL,           -- 状态数据（JSON格式）
    version INTEGER NOT NULL,           -- 状态版本号
    created_at TIMESTAMP NOT NULL,      -- 创建时间
    updated_at TIMESTAMP NOT NULL,      -- 更新时间
    last_used_at TIMESTAMP,             -- 最后使用时间
    is_active BOOLEAN NOT NULL,         -- 是否激活
    UNIQUE(user_id, platform_id, state_type)
);

-- 索引
CREATE INDEX idx_platform_user_states_user_id ON platform_user_states(user_id);
CREATE INDEX idx_platform_user_states_platform_id ON platform_user_states(platform_id);
CREATE INDEX idx_platform_user_states_last_used_at ON platform_user_states(last_used_at);
```

#### 状态类型说明

1. system 状态：

   - 系统配置信息
   - 环境变量设置
   - 依赖项版本
   - 系统文件路径

2. user 状态：
   - 用户登录信息
   - 用户偏好设置
   - 游戏配置
   - 控制器设置
   - 显示设置
   - 音频设置

#### 状态数据示例

1. Lutris system 状态：

```json
{
  "wine_version": "7.0",
  "dxvk_version": "2.0",
  "python_version": "3.8",
  "system_path": "/data/beagle-wind/games/lutris/system",
  "environment": {
    "WINEARCH": "win64",
    "WINEPREFIX": "/data/beagle-wind/games/lutris/prefix"
  }
}
```

2. Steam user 状态：

```json
{
  "steam_id": "76561198123456789",
  "login_state": "logged_in",
  "steam_play": {
    "enabled": true,
    "compatibility_tool": "Proton-7.0"
  },
  "display": {
    "resolution": "1920x1080",
    "refresh_rate": 60
  },
  "audio": {
    "output_device": "default",
    "volume": 100
  },
  "controller": {
    "enabled": true,
    "layout": "xbox"
  }
}
```

3. Nintendo Switch user 状态：

```json
{
  "yuzu_version": "1200",
  "keys_path": "/data/beagle-wind/games/switch/keys",
  "firmware_path": "/data/beagle-wind/games/switch/firmware",
  "display": {
    "resolution": "1280x720",
    "fullscreen": false
  },
  "audio": {
    "output_device": "default",
    "volume": 100
  },
  "controller": {
    "enabled": true,
    "layout": "pro"
  }
}
```

### 平台状态历史表 (platform_state_history)

```sql
CREATE TABLE platform_state_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,           -- 用户ID
    platform_id TEXT NOT NULL,          -- 平台ID
    state_type TEXT NOT NULL,           -- 状态类型
    state_data TEXT NOT NULL,           -- 状态数据
    version INTEGER NOT NULL,           -- 状态版本号
    created_at TIMESTAMP NOT NULL,      -- 创建时间
    created_by INTEGER NOT NULL,        -- 创建者ID
    reason TEXT,                        -- 变更原因
    UNIQUE(user_id, platform_id, state_type, version)
);

-- 索引
CREATE INDEX idx_platform_state_history_user_id ON platform_state_history(user_id);
CREATE INDEX idx_platform_state_history_platform_id ON platform_state_history(platform_id);
CREATE INDEX idx_platform_state_history_created_at ON platform_state_history(created_at);
```

#### 用途说明

1. 状态追踪：

   - 记录用户平台配置的变更历史
   - 支持配置回滚
   - 审计配置变更

2. 版本控制：

   - 管理配置版本
   - 支持增量更新
   - 配置冲突处理

3. 数据恢复：
   - 支持配置回滚
   - 数据备份恢复
   - 故障恢复

### 平台状态同步表 (platform_state_sync)

```sql
CREATE TABLE platform_state_sync (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,           -- 用户ID
    platform_id TEXT NOT NULL,          -- 平台ID
    state_type TEXT NOT NULL,           -- 状态类型
    sync_status TEXT NOT NULL,          -- 同步状态：pending/synced/failed
    sync_version INTEGER NOT NULL,      -- 同步版本号
    last_sync_at TIMESTAMP,             -- 最后同步时间
    error_message TEXT,                 -- 错误信息
    created_at TIMESTAMP NOT NULL,      -- 创建时间
    updated_at TIMESTAMP NOT NULL,      -- 更新时间
    UNIQUE(user_id, platform_id, state_type)
);

-- 索引
CREATE INDEX idx_platform_state_sync_user_id ON platform_state_sync(user_id);
CREATE INDEX idx_platform_state_sync_platform_id ON platform_state_sync(platform_id);
CREATE INDEX idx_platform_state_sync_status ON platform_state_sync(sync_status);
```

#### 用途说明

1. 同步管理：

   - 追踪配置同步状态
   - 处理同步冲突
   - 错误恢复

2. 状态监控：

   - 监控同步状态
   - 检测同步失败
   - 同步性能统计

3. 数据一致性：
   - 确保多设备同步
   - 维护数据一致性
   - 处理网络问题
