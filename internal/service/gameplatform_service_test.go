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

// mockGamePlatformErrorStore 模拟返回错误的游戏平台存储实现
type mockGamePlatformErrorStore struct{}

func (s *mockGamePlatformErrorStore) List() ([]models.GamePlatform, error) {
	return nil, assert.AnError
}

func (s *mockGamePlatformErrorStore) Get(id string) (models.GamePlatform, error) {
	return models.GamePlatform{}, assert.AnError
}

func (s *mockGamePlatformErrorStore) Add(platform models.GamePlatform) error {
	return assert.AnError
}

func (s *mockGamePlatformErrorStore) Update(platform models.GamePlatform) error {
	return assert.AnError
}

func (s *mockGamePlatformErrorStore) Delete(id string) error {
	return assert.AnError
}

func (s *mockGamePlatformErrorStore) Cleanup() error {
	return nil
}

// 测试数据
var testPlatform = models.GamePlatform{
	ID:        "test-platform-1",
	Name:      "Test Platform",
	Version:   "1.0.0",
	Type:      "gaming",
	OS:        "Linux",
	Image:     "test-image:1.0",
	Bin:       "/usr/bin/test",
	Features:  []string{"feature1", "feature2"},
	Config:    map[string]string{"key1": "value1"},
	Files:     []models.GamePlatformFile{},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func TestGamePlatformService_List(t *testing.T) {
	tests := []struct {
		name           string
		params         GamePlatformListParams
		store          store.GamePlatformStore
		expectedResult *GamePlatformListResult
		expectedError  error
	}{
		{
			name: "成功获取平台列表",
			params: GamePlatformListParams{
				Page:     1,
				PageSize: 20,
			},
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &GamePlatformListResult{
				Total: 1,
				Items: []models.GamePlatform{testPlatform},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: GamePlatformListParams{
				Page:     1,
				PageSize: 20,
			},
			store:          &mockGamePlatformErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			result, err := service.List(tt.params)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, result)
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

func TestGamePlatformService_Get(t *testing.T) {
	tests := []struct {
		name           string
		platformID     string
		store          store.GamePlatformStore
		expectedResult *models.GamePlatform
		expectedError  error
	}{
		{
			name:       "成功获取平台",
			platformID: "test-platform-1",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &testPlatform,
			expectedError:  nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
		{
			name:           "存储层返回错误",
			platformID:     "test-platform-1",
			store:          &mockGamePlatformErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			result, err := service.Get(tt.platformID)
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

func TestGamePlatformService_Create(t *testing.T) {
	tests := []struct {
		name          string
		platform      models.GamePlatform
		store         store.GamePlatformStore
		expectedID    string
		expectedError error
	}{
		{
			name:     "成功创建平台",
			platform: testPlatform,
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedID:    testPlatform.ID,
			expectedError: nil,
		},
		{
			name:     "平台ID已存在",
			platform: testPlatform,
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedID:    "",
			expectedError: fmt.Errorf("平台ID已存在: %s", testPlatform.ID),
		},
		{
			name:          "存储层返回错误",
			platform:      testPlatform,
			store:         &mockGamePlatformErrorStore{},
			expectedID:    "",
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			id, err := service.Create(tt.platform)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, "", id)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestGamePlatformService_Update(t *testing.T) {
	updatedPlatform := testPlatform
	updatedPlatform.Name = "Updated Platform"

	tests := []struct {
		name          string
		platformID    string
		platform      models.GamePlatform
		store         store.GamePlatformStore
		expectedError error
	}{
		{
			name:       "成功更新平台",
			platformID: "test-platform-1",
			platform:   updatedPlatform,
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			platform:   updatedPlatform,
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: assert.AnError,
		},
		{
			name:          "存储层返回错误",
			platformID:    "test-platform-1",
			platform:      updatedPlatform,
			store:         &mockGamePlatformErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			err := service.Update(tt.platformID, tt.platform)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGamePlatformService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		platformID    string
		store         store.GamePlatformStore
		expectedError error
	}{
		{
			name:       "成功删除平台",
			platformID: "test-platform-1",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: assert.AnError,
		},
		{
			name:          "存储层返回错误",
			platformID:    "test-platform-1",
			store:         &mockGamePlatformErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			err := service.Delete(tt.platformID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGamePlatformService_GetAccess(t *testing.T) {
	tests := []struct {
		name          string
		platformID    string
		store         store.GamePlatformStore
		expectedError error
		checkResult   func(t *testing.T, result GamePlatformAccessResult)
	}{
		{
			name:       "成功获取平台访问链接",
			platformID: "test-platform-1",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
			checkResult: func(t *testing.T, result GamePlatformAccessResult) {
				assert.NotEmpty(t, result.Link)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: assert.AnError,
			checkResult:   nil,
		},
		{
			name:          "存储层返回错误",
			platformID:    "test-platform-1",
			store:         &mockGamePlatformErrorStore{},
			expectedError: assert.AnError,
			checkResult:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			result, err := service.GetAccess(tt.platformID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}

func TestGamePlatformService_RefreshAccess(t *testing.T) {
	tests := []struct {
		name          string
		platformID    string
		store         store.GamePlatformStore
		expectedError error
		checkResult   func(t *testing.T, result GamePlatformAccessResult)
	}{
		{
			name:       "成功刷新平台访问链接",
			platformID: "test-platform-1",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testPlatform)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
			checkResult: func(t *testing.T, result GamePlatformAccessResult) {
				assert.NotEmpty(t, result.Link)
				assert.True(t, result.ExpiresAt.After(time.Now()))
			},
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			store: func() store.GamePlatformStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGamePlatformStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: assert.AnError,
			checkResult:   nil,
		},
		{
			name:          "存储层返回错误",
			platformID:    "test-platform-1",
			store:         &mockGamePlatformErrorStore{},
			expectedError: assert.AnError,
			checkResult:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGamePlatformService(tt.store)
			result, err := service.RefreshAccess(tt.platformID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, result)
			}
		})
	}
}
