package service

import (
	"os"
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
	"github.com/stretchr/testify/assert"
)

// 测试数据
var testNode = models.GameNode{
	ID:       "test-node-1",
	Name:     "Test Node",
	Model:    "test-model",
	Type:     models.GameNodeTypePhysical,
	Location: "test-location",
	Hardware: map[string]string{"cpu": "2", "memory": "4"},
	Network:  map[string]string{"ip": "127.0.0.1"},
	Labels:   map[string]string{"region": "test"},
	Status: models.GameNodeStatus{
		State:      models.GameNodeStateOnline,
		Online:     true,
		LastOnline: time.Now(),
		UpdatedAt:  time.Now(),
		Resources:  map[string]string{"cpu": "2", "memory": "4"},
		Metrics:    map[string]interface{}{"cpu_usage": 0.5, "latency": 100},
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func TestListNodes(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name           string
		params         NodeListParams
		setup          func()
		expectedResult NodeListResult
		expectedError  error
	}{
		{
			name: "成功获取节点列表",
			params: NodeListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedResult: NodeListResult{
				Total: 1,
				Items: []models.GameNode{testNode},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: NodeListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedResult: NodeListResult{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.ListNodes(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.Total, result.Total)
			assert.Equal(t, len(tt.expectedResult.Items), len(result.Items))
			for i, expected := range tt.expectedResult.Items {
				assert.Equal(t, expected.ID, result.Items[i].ID)
				assert.Equal(t, expected.Name, result.Items[i].Name)
				assert.Equal(t, expected.Type, result.Items[i].Type)
			}
		})
	}
}

func TestGetNode(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name           string
		nodeID         string
		setup          func()
		expectedResult *models.GameNode
		expectedError  error
	}{
		{
			name:   "成功获取节点",
			nodeID: "test-node-1",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedResult: &testNode,
			expectedError:  nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.GetNode(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
			if tt.expectedResult == nil {
				assert.Nil(t, result)
				return
			}
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.Name, result.Name)
			assert.Equal(t, tt.expectedResult.Type, result.Type)
		})
	}
}

func TestCreateNode(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		node          models.GameNode
		setup         func()
		expectedError error
	}{
		{
			name: "成功创建节点",
			node: testNode,
			setup: func() {
				// 不需要特殊设置
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			node: testNode,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.CreateNode(tt.node)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNode(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		nodeID        string
		node          models.GameNode
		setup         func()
		expectedError error
	}{
		{
			name:   "成功更新节点",
			nodeID: "test-node-1",
			node:   testNode,
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			node:   testNode,
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: assert.AnError,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			node:   testNode,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateNode(tt.nodeID, tt.node)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeleteNode(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		nodeID        string
		setup         func()
		expectedError error
	}{
		{
			name:   "成功删除节点",
			nodeID: "test-node-1",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.DeleteNode(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeStatus(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		nodeID        string
		status        string
		setup         func()
		expectedError error
	}{
		{
			name:   "成功更新节点状态",
			nodeID: "test-node-1",
			status: "offline",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			status: "offline",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			status: "offline",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateNodeStatus(tt.nodeID, tt.status)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeMetrics(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		nodeID        string
		metrics       map[string]interface{}
		setup         func()
		expectedError error
	}{
		{
			name:   "成功更新节点指标",
			nodeID: "test-node-1",
			metrics: map[string]interface{}{
				"cpu_usage": 0.8,
				"memory":    "4GB",
			},
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			metrics: map[string]interface{}{
				"cpu_usage": 0.8,
			},
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: assert.AnError,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			metrics: map[string]interface{}{
				"cpu_usage": 0.8,
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateNodeMetrics(tt.nodeID, tt.metrics)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeResources(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		nodeID        string
		resources     map[string]interface{}
		setup         func()
		expectedError error
	}{
		{
			name:   "成功更新节点资源",
			nodeID: "test-node-1",
			resources: map[string]interface{}{
				"cpu":    "4",
				"memory": "8GB",
			},
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			resources: map[string]interface{}{
				"cpu": "4",
			},
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: assert.AnError,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			resources: map[string]interface{}{
				"cpu": "4",
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateNodeResources(tt.nodeID, tt.resources)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestUpdateNodeOnlineStatus(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	nodeStore, err := store.NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer nodeStore.Cleanup()

	service := NewNodeService(nodeStore)

	tests := []struct {
		name          string
		nodeID        string
		online        bool
		setup         func()
		expectedError error
	}{
		{
			name:   "成功更新节点在线状态",
			nodeID: "test-node-1",
			online: false,
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			online: false,
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			online: false,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateNodeOnlineStatus(tt.nodeID, tt.online)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
