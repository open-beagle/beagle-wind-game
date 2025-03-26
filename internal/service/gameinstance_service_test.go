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

// mockGameInstanceErrorStore 模拟返回错误的游戏实例存储实现
type mockGameInstanceErrorStore struct{}

func (s *mockGameInstanceErrorStore) List() ([]models.GameInstance, error) {
	return nil, assert.AnError
}

func (s *mockGameInstanceErrorStore) Get(id string) (models.GameInstance, error) {
	return models.GameInstance{}, assert.AnError
}

func (s *mockGameInstanceErrorStore) Add(instance models.GameInstance) error {
	return assert.AnError
}

func (s *mockGameInstanceErrorStore) Update(instance models.GameInstance) error {
	return assert.AnError
}

func (s *mockGameInstanceErrorStore) Delete(id string) error {
	return assert.AnError
}

func (s *mockGameInstanceErrorStore) FindByCardID(cardID string) ([]models.GameInstance, error) {
	return nil, assert.AnError
}

func (s *mockGameInstanceErrorStore) FindByNodeID(nodeID string) ([]models.GameInstance, error) {
	return nil, assert.AnError
}

func (s *mockGameInstanceErrorStore) Cleanup() error {
	return nil
}

// 测试数据
var testInstance = models.GameInstance{
	ID:         "test-instance-1",
	NodeID:     "test-node-1",
	CardID:     "test-card-1",
	PlatformID: "test-platform-1",
	Status:     "stopped",
	CreatedAt:  time.Now(),
	UpdatedAt:  time.Now(),
}

func TestGameInstanceService_List(t *testing.T) {
	tests := []struct {
		name           string
		params         GameInstanceListParams
		store          store.GameInstanceStore
		expectedResult *InstanceListResult
		expectedError  error
	}{
		{
			name: "成功获取实例列表",
			params: GameInstanceListParams{
				Page:     1,
				PageSize: 20,
			},
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &InstanceListResult{
				Total: 1,
				Items: []models.GameInstance{testInstance},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: GameInstanceListParams{
				Page:     1,
				PageSize: 20,
			},
			store:          &mockGameInstanceErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
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
				assert.Equal(t, expected.NodeID, result.Items[i].NodeID)
				assert.Equal(t, expected.CardID, result.Items[i].CardID)
				assert.Equal(t, expected.Status, result.Items[i].Status)
			}
		})
	}
}

func TestGameInstanceService_Get(t *testing.T) {
	tests := []struct {
		name           string
		instanceID     string
		store          store.GameInstanceStore
		expectedResult models.GameInstance
		expectedError  error
	}{
		{
			name:       "成功获取实例",
			instanceID: "test-instance-1",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: testInstance,
			expectedError:  nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: models.GameInstance{},
			expectedError:  fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:           "存储层返回错误",
			instanceID:     "test-instance-1",
			store:          &mockGameInstanceErrorStore{},
			expectedResult: models.GameInstance{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
			result, err := service.Get(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedResult, result)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedResult.ID, result.ID)
			assert.Equal(t, tt.expectedResult.NodeID, result.NodeID)
			assert.Equal(t, tt.expectedResult.CardID, result.CardID)
			assert.Equal(t, tt.expectedResult.Status, result.Status)
		})
	}
}

func TestGameInstanceService_Start(t *testing.T) {
	tests := []struct {
		name          string
		instanceID    string
		store         store.GameInstanceStore
		expectedError error
	}{
		{
			name:       "成功启动实例",
			instanceID: "test-instance-1",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:          "存储层返回错误",
			instanceID:    "test-instance-1",
			store:         &mockGameInstanceErrorStore{},
			expectedError: assert.AnError,
		},
		{
			name:       "实例已在运行中",
			instanceID: "test-instance-1",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				runningInstance := testInstance
				runningInstance.Status = "running"
				err = store.Add(runningInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: ErrInstanceAlreadyRunning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
			err := service.Start(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameInstanceService_Stop(t *testing.T) {
	tests := []struct {
		name          string
		instanceID    string
		store         store.GameInstanceStore
		expectedError error
	}{
		{
			name:       "成功停止实例",
			instanceID: "test-instance-1",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				runningInstance := testInstance
				runningInstance.Status = "running"
				err = store.Add(runningInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:          "存储层返回错误",
			instanceID:    "test-instance-1",
			store:         &mockGameInstanceErrorStore{},
			expectedError: assert.AnError,
		},
		{
			name:       "实例已停止",
			instanceID: "test-instance-1",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("实例已停止"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
			err := service.Stop(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameInstanceService_Create(t *testing.T) {
	tests := []struct {
		name          string
		params        CreateInstanceParams
		store         store.GameInstanceStore
		expectedID    string
		expectedError error
	}{
		{
			name: "成功创建实例",
			params: CreateInstanceParams{
				NodeID:     "test-node-1",
				PlatformID: "test-platform-1",
				CardID:     "test-card-1",
				Config:     "{}",
			},
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedID:    "test-node-1-test-card-1",
			expectedError: nil,
		},
		{
			name: "实例ID已存在",
			params: CreateInstanceParams{
				NodeID:     "test-node-1",
				PlatformID: "test-platform-1",
				CardID:     "test-card-1",
				Config:     "{}",
			},
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				instance := testInstance
				instance.ID = "test-node-1-test-card-1"
				err = store.Add(instance)
				assert.NoError(t, err)
				return store
			}(),
			expectedID:    "",
			expectedError: fmt.Errorf("实例ID已存在: test-node-1-test-card-1"),
		},
		{
			name: "存储层返回错误",
			params: CreateInstanceParams{
				NodeID:     "test-node-1",
				PlatformID: "test-platform-1",
				CardID:     "test-card-1",
				Config:     "{}",
			},
			store:         &mockGameInstanceErrorStore{},
			expectedID:    "",
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
			id, err := service.Create(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedID, id)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestGameInstanceService_Update(t *testing.T) {
	updatedInstance := testInstance
	updatedInstance.Status = "running"

	tests := []struct {
		name          string
		instanceID    string
		instance      models.GameInstance
		store         store.GameInstanceStore
		expectedError error
	}{
		{
			name:       "成功更新实例",
			instanceID: "test-instance-1",
			instance:   updatedInstance,
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			instance:   updatedInstance,
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:          "存储层返回错误",
			instanceID:    "test-instance-1",
			instance:      updatedInstance,
			store:         &mockGameInstanceErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
			err := service.Update(tt.instanceID, tt.instance)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameInstanceService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		instanceID    string
		store         store.GameInstanceStore
		expectedError error
	}{
		{
			name:       "成功删除实例",
			instanceID: "test-instance-1",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testInstance)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:       "实例不存在",
			instanceID: "non-existent-instance",
			store: func() store.GameInstanceStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameInstanceStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("实例不存在: non-existent-instance"),
		},
		{
			name:          "存储层返回错误",
			instanceID:    "test-instance-1",
			store:         &mockGameInstanceErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameInstanceService(tt.store)
			err := service.Delete(tt.instanceID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
