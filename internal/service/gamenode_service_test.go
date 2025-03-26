package service

import (
	"fmt"
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

// mockGameNodeErrorStore 模拟存储层错误
type mockGameNodeErrorStore struct{}

func (m *mockGameNodeErrorStore) Load() error {
	return assert.AnError
}

func (m *mockGameNodeErrorStore) Save() error {
	return assert.AnError
}

func (m *mockGameNodeErrorStore) List() ([]models.GameNode, error) {
	return nil, assert.AnError
}

func (m *mockGameNodeErrorStore) Get(id string) (models.GameNode, error) {
	return models.GameNode{}, assert.AnError
}

func (m *mockGameNodeErrorStore) Add(node models.GameNode) error {
	return assert.AnError
}

func (m *mockGameNodeErrorStore) Update(node models.GameNode) error {
	return assert.AnError
}

func (m *mockGameNodeErrorStore) Delete(id string) error {
	return assert.AnError
}

func (m *mockGameNodeErrorStore) Cleanup() error {
	return assert.AnError
}

func TestGameNodeService_List(t *testing.T) {
	tests := []struct {
		name           string
		params         GameNodeListParams
		store          store.GameNodeStore
		expectedResult *GameNodeListResult
		expectedError  error
	}{
		{
			name: "成功获取节点列表",
			params: GameNodeListParams{
				Page:     1,
				PageSize: 20,
			},
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &GameNodeListResult{
				Total: 1,
				Items: []models.GameNode{testNode},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: GameNodeListParams{
				Page:     1,
				PageSize: 20,
			},
			store:          &mockGameNodeErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			result, err := service.List(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
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

func TestGameNodeService_Get(t *testing.T) {
	tests := []struct {
		name           string
		nodeID         string
		store          store.GameNodeStore
		expectedResult *models.GameNode
		expectedError  error
	}{
		{
			name:   "成功获取节点",
			nodeID: "test-node-1",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &testNode,
			expectedError:  nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: nil,
			expectedError:  fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:           "存储层返回错误",
			nodeID:         "test-node-1",
			store:          &mockGameNodeErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			result, err := service.Get(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
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

func TestGameNodeService_Create(t *testing.T) {
	tests := []struct {
		name          string
		node          models.GameNode
		store         store.GameNodeStore
		expectedError error
	}{
		{
			name: "成功创建节点",
			node: testNode,
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name: "节点ID已存在",
			node: testNode,
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("节点ID已存在: %s", testNode.ID),
		},
		{
			name:          "存储层返回错误",
			node:          testNode,
			store:         &mockGameNodeErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			err := service.Create(tt.node)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameNodeService_Update(t *testing.T) {
	updatedNode := testNode
	updatedNode.Name = "Updated Node"

	tests := []struct {
		name          string
		nodeID        string
		node          models.GameNode
		store         store.GameNodeStore
		expectedError error
	}{
		{
			name:   "成功更新节点",
			nodeID: "test-node-1",
			node:   updatedNode,
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			node:   updatedNode,
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:          "存储层返回错误",
			nodeID:        "test-node-1",
			node:          updatedNode,
			store:         &mockGameNodeErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			err := service.Update(tt.nodeID, tt.node)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameNodeService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		store         store.GameNodeStore
		expectedError error
	}{
		{
			name:   "成功删除节点",
			nodeID: "test-node-1",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:          "存储层返回错误",
			nodeID:        "test-node-1",
			store:         &mockGameNodeErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			err := service.Delete(tt.nodeID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameNodeService_UpdateStatus(t *testing.T) {
	newStatus := models.GameNodeStatus{
		State:      models.GameNodeStateOnline,
		Online:     true,
		LastOnline: time.Now(),
		UpdatedAt:  time.Now(),
		Resources:  map[string]string{},
		Metrics:    map[string]interface{}{},
	}

	tests := []struct {
		name          string
		nodeID        string
		status        models.GameNodeStatus
		store         store.GameNodeStore
		expectedError error
	}{
		{
			name:   "成功更新节点状态",
			nodeID: "test-node-1",
			status: newStatus,
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			status: newStatus,
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
		},
		{
			name:          "存储层返回错误",
			nodeID:        "test-node-1",
			status:        newStatus,
			store:         &mockGameNodeErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			err := service.UpdateStatusState(tt.nodeID, string(tt.status.State))
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameNodeService_GetAccess(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		store         store.GameNodeStore
		expectedError error
		checkResult   func(t *testing.T, result NodeAccessResult)
	}{
		{
			name:   "成功获取节点访问链接",
			nodeID: "test-node-1",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
			checkResult: func(t *testing.T, result NodeAccessResult) {
				assert.NotEmpty(t, result.Link)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
			checkResult:   nil,
		},
		{
			name:          "存储层返回错误",
			nodeID:        "test-node-1",
			store:         &mockGameNodeErrorStore{},
			expectedError: fmt.Errorf("存储层错误: %v", assert.AnError),
			checkResult:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			result, err := service.GetAccess(tt.nodeID)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestGameNodeService_RefreshAccess(t *testing.T) {
	tests := []struct {
		name          string
		nodeID        string
		store         store.GameNodeStore
		expectedError error
		checkResult   func(t *testing.T, result NodeAccessResult)
	}{
		{
			name:   "成功刷新节点访问链接",
			nodeID: "test-node-1",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testNode)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
			checkResult: func(t *testing.T, result NodeAccessResult) {
				assert.NotEmpty(t, result.Link)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:   "节点不存在",
			nodeID: "non-existent-node",
			store: func() store.GameNodeStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameNodeStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("节点不存在: non-existent-node"),
			checkResult:   nil,
		},
		{
			name:          "存储层返回错误",
			nodeID:        "test-node-1",
			store:         &mockGameNodeErrorStore{},
			expectedError: fmt.Errorf("存储层错误: %v", assert.AnError),
			checkResult:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameNodeService(tt.store)
			result, err := service.RefreshAccess(tt.nodeID)
			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
