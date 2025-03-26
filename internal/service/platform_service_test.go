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
	Files:     []models.PlatformFile{},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

func TestListPlatforms(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name           string
		params         PlatformListParams
		setup          func()
		expectedResult PlatformListResult
		expectedError  error
	}{
		{
			name: "成功获取平台列表",
			params: PlatformListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedResult: PlatformListResult{
				Total: 1,
				Items: []models.GamePlatform{testPlatform},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: PlatformListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedResult: PlatformListResult{},
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.ListPlatforms(tt.params)
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

func TestGetPlatform(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name           string
		platformID     string
		setup          func()
		expectedResult *models.GamePlatform
		expectedError  error
	}{
		{
			name:       "成功获取平台",
			platformID: "test-platform-1",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedResult: &testPlatform,
			expectedError:  nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedResult: nil,
			expectedError:  nil,
		},
		{
			name:       "存储层返回错误",
			platformID: "test-platform-1",
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
			result, err := service.GetPlatform(tt.platformID)
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

func TestCreatePlatform(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name          string
		platform      models.GamePlatform
		setup         func()
		expectedID    string
		expectedError error
	}{
		{
			name:     "成功创建平台",
			platform: testPlatform,
			setup: func() {
				// 不需要特殊设置
			},
			expectedID:    testPlatform.ID,
			expectedError: nil,
		},
		{
			name:     "存储层返回错误",
			platform: testPlatform,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
			},
			expectedID:    "",
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			id, err := service.CreatePlatform(tt.platform)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, id)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestUpdatePlatform(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name          string
		platformID    string
		platform      models.GamePlatform
		setup         func()
		expectedError error
	}{
		{
			name:       "成功更新平台",
			platformID: "test-platform-1",
			platform:   testPlatform,
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			platform:   testPlatform,
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: assert.AnError,
		},
		{
			name:       "存储层返回错误",
			platformID: "test-platform-1",
			platform:   testPlatform,
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
			err := service.UpdatePlatform(tt.platformID, tt.platform)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeletePlatform(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name          string
		platformID    string
		setup         func()
		expectedError error
	}{
		{
			name:       "成功删除平台",
			platformID: "test-platform-1",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "存储层返回错误",
			platformID: "test-platform-1",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
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
			err := service.DeletePlatform(tt.platformID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetPlatformAccess(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name          string
		platformID    string
		setup         func()
		expectedError error
	}{
		{
			name:       "成功获取访问链接",
			platformID: "test-platform-1",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: assert.AnError,
		},
		{
			name:       "存储层返回错误",
			platformID: "test-platform-1",
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
			result, err := service.GetPlatformAccess(tt.platformID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, result)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, result.Link)
			assert.NotEmpty(t, result.ExpiresAt)
		})
	}
}

func TestRefreshPlatformAccess(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	platformStore, err := store.NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer platformStore.Cleanup()

	service := NewPlatformService(platformStore)

	tests := []struct {
		name          string
		platformID    string
		setup         func()
		expectedError error
	}{
		{
			name:       "成功刷新访问链接",
			platformID: "test-platform-1",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent-platform",
			setup: func() {
				err := platformStore.Add(testPlatform)
				assert.NoError(t, err)
			},
			expectedError: assert.AnError,
		},
		{
			name:       "存储层返回错误",
			platformID: "test-platform-1",
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
			result, err := service.RefreshPlatformAccess(tt.platformID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, result)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, result.Link)
			assert.NotEmpty(t, result.ExpiresAt)
		})
	}
}
