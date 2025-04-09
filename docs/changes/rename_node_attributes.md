# 节点属性重命名变更

## 变更说明

将 GameNode 的属性进行重命名：

- `node_id` -> `id`
- `name` -> `alias`

## 影响范围

### Proto 文件

- [x] internal/proto/gamenode.proto
  - [x] RegisterRequest 消息
  - [x] HeartbeatRequest 消息
  - [x] MetricsReport 消息
  - [x] ResourceInfo 消息
  - [x] ExecuteRequest 消息
  - [x] PipelineStatusUpdate 消息
  - [x] Event 消息
- [x] 重新生成 proto 文件

### 模型文件

- [x] internal/models/gamenode.go
- [x] internal/models/gamenode_pipeline.go

### 服务文件

- [x] internal/gamenode/gamenode_server.go
- [x] internal/gamenode/gamenode_agent.go

## 变更进度

### 第一阶段：Proto 文件修改

1. [x] 修改 RegisterRequest 消息定义
2. [x] 修改其他相关消息定义
3. [x] 重新生成 proto 文件
4. [x] 验证生成的代码

### 第二阶段：模型文件修改

1. [x] 更新 GameNode 结构体
2. [x] 更新相关方法和函数
3. [x] 更新测试用例

### 第三阶段：服务文件修改

1. [x] 更新 GameNodeServer 实现
2. [x] 更新 GameNodeAgent 实现

### 第四阶段：测试和验证

1. [x] 验证功能完整性
2. [x] 检查向后兼容性
3. [x] 更新文档

## 注意事项

1. 保持向后兼容性
2. 确保所有引用都被正确更新
3. 维护完整的测试覆盖率
4. 记录所有变更

## 变更总结

1. Proto 文件修改

   - 修改了 RegisterRequest、HeartbeatRequest、MetricsReport、ResourceInfo、ExecuteRequest、PipelineStatusUpdate 和 Event 消息中的字段名称
   - 重新生成了 proto 文件并验证了生成的代码

2. 模型文件修改

   - 更新了 GameNode 结构体中的字段名称
   - 更新了 PipelineStatus 和 GameNodePipeline 结构体中的字段名称
   - 确保了所有相关方法和函数都使用了新的字段名称

3. 服务文件修改

   - 更新了 GameNodeServer 实现，使用新的字段名称
   - 更新了 GameNodeAgent 实现，使用新的字段名称
   - 确保了所有事件和日志相关的代码都使用了新的字段名称

4. 向后兼容性

   - 所有修改都保持了 API 的向后兼容性
   - 字段重命名不会影响现有的功能和数据

5. 文档更新
   - 创建了变更文档，记录了所有修改内容
   - 更新了相关注释和文档字符串
