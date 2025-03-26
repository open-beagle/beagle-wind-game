package service

import (
	"fmt"
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
	Model:    "Test Model",
	Type:     models.GameNodeTypePhysical,
	Location: "Test Location",
	Hardware: map[string]string{
		"cpu":    "Test CPU",
		"memory": "16GB",
		"gpu":    "Test GPU",
	},
	Network: map[string]string{
		"ip":       "192.168.1.1",
		"port":     "8080",
		"protocol": "TCP",
	},
	Labels: map[string]string{
		"key1": "value1",
	},
	Status: models.GameNodeStatus{
		State:      models.GameNodeStateOffline,
		Online:     false,
		LastOnline: time.Now(),
		UpdatedAt:  time.Now(),
		Resources:  map[string]string{},
		Metrics:    map[string]interface{}{},
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
			expectedError:  fmt.Errorf("存储层错误"),
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
				assert.Equal(t, tt.expectedError.Error(), err.Error())
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
			name:           "节点不存在",
			nodeID:         "non-existent-node",
			setup:          nil,
			expectedResult: nil,
			expectedError:  fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedResult: nil,
			expectedError:  fmt.Errorf("存储层错误"),
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
				assert.Equal(t, tt.expectedError.Error(), err.Error())
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
			name: "节点ID已存在",
			node: testNode,
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("节点ID已存在: %s", testNode.ID),
		},
		{
			name: "存储层返回错误",
			node: testNode,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: fmt.Errorf("存储层错误"),
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
				assert.Equal(t, tt.expectedError.Error(), err.Error())
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

	updatedNode := testNode
	updatedNode.Name = "Updated Node"

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
			node:   updatedNode,
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:          "节点不存在",
			nodeID:        "non-existent-node",
			node:          updatedNode,
			setup:         nil,
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			node:   updatedNode,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: fmt.Errorf("存储层错误"),
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
				assert.Equal(t, tt.expectedError.Error(), err.Error())
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
			name:          "节点不存在",
			nodeID:        "non-existent-node",
			setup:         nil,
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: fmt.Errorf("存储层错误"),
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
				assert.Equal(t, tt.expectedError.Error(), err.Error())
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
			status: string(models.GameNodeStateOnline),
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:          "节点不存在",
			nodeID:        "non-existent-node",
			status:        string(models.GameNodeStateOnline),
			setup:         nil,
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			status: string(models.GameNodeStateOnline),
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: fmt.Errorf("存储层错误"),
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
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetNodeAccess(t *testing.T) {
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
		checkResult   func(t *testing.T, result NodeAccessResult)
	}{
		{
			name:   "成功获取节点访问链接",
			nodeID: "test-node-1",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, result NodeAccessResult) {
				assert.NotEmpty(t, result.Link)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:          "节点不存在",
			nodeID:        "non-existent-node",
			setup:         nil,
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
			checkResult:   nil,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: fmt.Errorf("存储层错误"),
			checkResult:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.GetNodeAccess(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestRefreshNodeAccess(t *testing.T) {
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
		checkResult   func(t *testing.T, result NodeAccessResult)
	}{
		{
			name:   "成功刷新节点访问链接",
			nodeID: "test-node-1",
			setup: func() {
				err := nodeStore.Add(testNode)
				assert.NoError(t, err)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, result NodeAccessResult) {
				assert.NotEmpty(t, result.Link)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:          "节点不存在",
			nodeID:        "non-existent-node",
			setup:         nil,
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
			checkResult:   nil,
		},
		{
			name:   "存储层返回错误",
			nodeID: "test-node-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedError: fmt.Errorf("存储层错误"),
			checkResult:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.RefreshNodeAccess(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
