# 节点注册流程改进

## 变更说明

改进 GameNodeAgent 和 GameNodeServer 的节点注册流程，完善节点信息的采集和更新机制。

## 影响范围

### 1. GameNodeServer 注册逻辑改进

#### 1.1 注册流程优化

- [ ] 在 Register 方法中增加节点存在性检查
- [ ] 区分首次注册和重复注册的处理逻辑
- [ ] 完善节点信息的更新机制

#### 1.2 具体修改

```go
// Register 处理节点注册请求
func (s *GameNodeServer) Register(ctx context.Context, req *proto.RegisterRequest) (*proto.RegisterResponse, error) {
    // 1. 检查节点是否存在
    node, err := s.store.Get(req.Id)
    if err != nil {
        // 节点不存在，创建新节点
        node = &models.GameNode{
            ID:        req.Id,
            Alias:     req.Alias,
            Model:     req.Model,
            Type:      req.Type,
            Location:  req.Location,
            Hardware:  req.Hardware,
            Network:   req.Network,
            Labels:    req.Labels,
            Status:    &models.GameNodeStatus{
                State:      "online",
                Online:     true,
                LastOnline: time.Now(),
            },
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        }
        if err := s.store.Create(node); err != nil {
            return &proto.RegisterResponse{
                Success: false,
                Message: fmt.Sprintf("创建节点失败: %v", err),
            }, nil
        }
    } else {
        // 节点存在，更新状态
        node.Status.State = "online"
        node.Status.Online = true
        node.Status.LastOnline = time.Now()
        node.UpdatedAt = time.Now()
        
        if err := s.store.Update(node); err != nil {
            return &proto.RegisterResponse{
                Success: false,
                Message: fmt.Sprintf("更新节点失败: %v", err),
            }, nil
        }
    }
    
    return &proto.RegisterResponse{
        Success: true,
        Message: "节点注册成功",
    }, nil
}
```

### 2. GameNodeAgent 注册逻辑改进

#### 2.1 信息采集优化

- [ ] 增加硬件信息采集
- [ ] 增加系统信息采集
- [ ] 完善资源信息采集

#### 2.2 具体修改

```go
// collectNodeInfo 采集节点信息
func (a *GameNodeAgent) collectNodeInfo() (*proto.RegisterRequest, error) {
    // 1. 采集硬件信息
    hardware, err := a.collectHardwareInfo()
    if err != nil {
        return nil, fmt.Errorf("采集硬件信息失败: %v", err)
    }
    
    // 2. 采集网络信息
    network, err := a.collectNetworkInfo()
    if err != nil {
        return nil, fmt.Errorf("采集网络信息失败: %v", err)
    }
    
    // 3. 获取节点标签
    labels, err := a.getNodeLabels()
    if err != nil {
        return nil, fmt.Errorf("获取节点标签失败: %v", err)
    }
    
    return &proto.RegisterRequest{
        Id:       a.id,
        Alias:    a.config.Alias,
        Model:    a.config.Model,
        Type:     a.config.Type,
        Location: a.config.Location,
        Hardware: hardware,
        Network:  network,
        Labels:   labels,
    }, nil
}

// collectHardwareInfo 采集硬件信息
func (a *GameNodeAgent) collectHardwareInfo() (map[string]string, error) {
    info := make(map[string]string)
    
    // 采集 CPU 信息
    cpuInfo, err := a.collectCPUInfo()
    if err != nil {
        return nil, err
    }
    info["cpu_model"] = cpuInfo.Model
    info["cpu_cores"] = strconv.Itoa(cpuInfo.Cores)
    info["cpu_threads"] = strconv.Itoa(cpuInfo.Threads)
    
    // 采集内存信息
    memInfo, err := a.collectMemoryInfo()
    if err != nil {
        return nil, err
    }
    info["memory_total"] = strconv.FormatInt(memInfo.Total, 10)
    info["memory_available"] = strconv.FormatInt(memInfo.Available, 10)
    
    // 采集 GPU 信息
    gpuInfo, err := a.collectGPUInfo()
    if err != nil {
        return nil, err
    }
    info["gpu_model"] = gpuInfo.Model
    info["gpu_memory"] = strconv.FormatInt(gpuInfo.MemoryTotal, 10)
    
    return info, nil
}

// collectNetworkInfo 采集网络信息
func (a *GameNodeAgent) collectNetworkInfo() (map[string]string, error) {
    info := make(map[string]string)
    
    // 采集网络接口信息
    interfaces, err := net.Interfaces()
    if err != nil {
        return nil, err
    }
    
    for _, iface := range interfaces {
        if iface.Flags&net.FlagUp == 0 {
            continue
        }
        
        addrs, err := iface.Addrs()
        if err != nil {
            continue
        }
        
        for _, addr := range addrs {
            if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                info[iface.Name] = ipnet.IP.String()
            }
        }
    }
    
    return info, nil
}

// getNodeLabels 获取节点标签
func (a *GameNodeAgent) getNodeLabels() (map[string]string, error) {
    labels := make(map[string]string)
    
    // 从配置文件加载标签
    for k, v := range a.config.Labels {
        labels[k] = v
    }
    
    // 添加系统标签
    labels["os"] = runtime.GOOS
    labels["arch"] = runtime.GOARCH
    
    return labels, nil
}
```

### 3. Proto 文件更新

#### 3.1 消息定义更新

```protobuf
// RegisterRequest 注册请求
message RegisterRequest {
    string id = 1;
    string alias = 2;
    string model = 3;
    string type = 4;
    string location = 5;
    map<string, string> hardware = 6;
    map<string, string> system = 7;
    map<string, string> labels = 8;
}

// RegisterResponse 注册响应
message RegisterResponse {
    bool success = 1;
    string message = 2;
}
```

## 变更进度

### 第一阶段：Proto 文件修改

- [x] 更新 RegisterRequest 消息定义
- [x] 添加硬件信息相关消息定义
- [x] 添加系统信息相关消息定义
- [x] 重新生成 proto 文件

### 第二阶段：GameNodeServer 修改

- [x] 更新 Register 方法实现
- [x] 完善节点信息更新逻辑
- [x] 添加错误处理
- [x] 更新测试用例

### 第三阶段：GameNodeAgent 修改

- [x] 实现硬件信息采集
- [x] 实现系统信息采集
- [x] 实现资源信息采集
- [x] 更新注册流程
- [x] 更新测试用例

### 第四阶段：测试和验证

- [ ] 验证节点注册流程
- [ ] 验证信息采集准确性
- [ ] 验证状态更新正确性
- [ ] 更新文档

## 注意事项

1. 硬件信息采集需要考虑不同操作系统的兼容性
2. 系统信息采集需要处理权限问题
3. 资源信息采集需要考虑性能影响
4. 保持向后兼容性
5. 完善错误处理和日志记录

## 变更总结

1. 改进了节点注册流程，区分首次注册和重复注册
2. 完善了节点信息的采集机制
3. 优化了节点状态的更新逻辑
4. 增强了错误处理和日志记录
