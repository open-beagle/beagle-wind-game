package service

import (
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListPlatforms(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		params        PlatformListParams
		mockPlatforms []models.GamePlatform
		mockError     error
		expectedTotal int
		expectedItems int
		expectedError error
	}{
		{
			name: "正常获取平台列表",
			params: PlatformListParams{
				Page:     1,
				PageSize: 10,
			},
			mockPlatforms: []models.GamePlatform{
				{
					ID:        "platform-1",
					Name:      "测试平台1",
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
				},
				{
					ID:        "platform-2",
					Name:      "测试平台2",
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
				},
			},
			mockError:     nil,
			expectedTotal: 2,
			expectedItems: 2,
			expectedError: nil,
		},
		{
			name: "存储层错误",
			params: PlatformListParams{
				Page:     1,
				PageSize: 10,
			},
			mockPlatforms: nil,
			mockError:     assert.AnError,
			expectedTotal: 0,
			expectedItems: 0,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("List").Return(tt.mockPlatforms, tt.mockError).Once()

			result, err := service.ListPlatforms(tt.params)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedTotal, result.Total)
			assert.Len(t, result.Items, tt.expectedItems)

			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetPlatform(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		platformID    string
		mockPlatform  models.GamePlatform
		mockError     error
		expectedError error
	}{
		{
			name:       "正常获取平台",
			platformID: "platform-1",
			mockPlatform: models.GamePlatform{
				ID:        "platform-1",
				Name:      "测试平台",
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
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "平台不存在",
			platformID:    "non-existent",
			mockPlatform:  models.GamePlatform{},
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("Get", tt.platformID).Return(tt.mockPlatform, tt.mockError).Once()

			platform, err := service.GetPlatform(tt.platformID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.mockPlatform, platform)

			mockStore.AssertExpectations(t)
		})
	}
}

func TestCreatePlatform(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		platform      models.GamePlatform
		mockError     error
		expectedError error
	}{
		{
			name: "正常创建平台",
			platform: models.GamePlatform{
				ID:       "platform-1",
				Name:     "测试平台",
				Version:  "1.0.0",
				Type:     "gaming",
				OS:       "Linux",
				Image:    "test-image:1.0",
				Bin:      "/usr/bin/test",
				Features: []string{"feature1", "feature2"},
				Config:   map[string]string{"key1": "value1"},
				Files:    []models.PlatformFile{},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name: "存储层错误",
			platform: models.GamePlatform{
				ID:       "platform-1",
				Name:     "测试平台",
				Version:  "1.0.0",
				Type:     "gaming",
				OS:       "Linux",
				Image:    "test-image:1.0",
				Bin:      "/usr/bin/test",
				Features: []string{"feature1", "feature2"},
				Config:   map[string]string{"key1": "value1"},
				Files:    []models.PlatformFile{},
			},
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("Add", mock.AnythingOfType("models.GamePlatform")).Return(tt.mockError).Once()

			id, err := service.CreatePlatform(tt.platform)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Empty(t, id)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.platform.ID, id)

			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdatePlatform(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		platformID    string
		platform      models.GamePlatform
		mockError     error
		expectedError error
	}{
		{
			name:       "正常更新平台",
			platformID: "platform-1",
			platform: models.GamePlatform{
				ID:       "platform-1",
				Name:     "更新后的平台",
				Version:  "1.0.0",
				Type:     "gaming",
				OS:       "Linux",
				Image:    "test-image:1.0",
				Bin:      "/usr/bin/test",
				Features: []string{"feature1", "feature2"},
				Config:   map[string]string{"key1": "value1"},
				Files:    []models.PlatformFile{},
			},
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:       "平台不存在",
			platformID: "non-existent",
			platform: models.GamePlatform{
				ID:       "non-existent",
				Name:     "更新后的平台",
				Version:  "1.0.0",
				Type:     "gaming",
				OS:       "Linux",
				Image:    "test-image:1.0",
				Bin:      "/usr/bin/test",
				Features: []string{"feature1", "feature2"},
				Config:   map[string]string{"key1": "value1"},
				Files:    []models.PlatformFile{},
			},
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("Get", tt.platformID).Return(tt.platform, tt.mockError).Once()
			if tt.mockError == nil {
				mockStore.On("Update", mock.AnythingOfType("models.GamePlatform")).Return(nil).Once()
			}

			err := service.UpdatePlatform(tt.platformID, tt.platform)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.NoError(t, err)

			mockStore.AssertExpectations(t)
		})
	}
}

func TestDeletePlatform(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		platformID    string
		mockError     error
		expectedError error
	}{
		{
			name:          "正常删除平台",
			platformID:    "platform-1",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "平台不存在",
			platformID:    "non-existent",
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("Get", tt.platformID).Return(models.GamePlatform{}, tt.mockError).Once()
			if tt.mockError == nil {
				mockStore.On("Delete", tt.platformID).Return(nil).Once()
			}

			err := service.DeletePlatform(tt.platformID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.NoError(t, err)

			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetPlatformAccess(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		platformID    string
		mockError     error
		expectedError error
	}{
		{
			name:          "正常获取访问链接",
			platformID:    "platform-1",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "平台不存在",
			platformID:    "non-existent",
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("Get", tt.platformID).Return(models.GamePlatform{}, tt.mockError).Once()

			result, err := service.GetPlatformAccess(tt.platformID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result.Link)
			assert.True(t, time.Now().Before(result.ExpiresAt))

			mockStore.AssertExpectations(t)
		})
	}
}

func TestRefreshPlatformAccess(t *testing.T) {
	mockStore := new(store.MockPlatformStore)
	service := NewPlatformService(mockStore)

	tests := []struct {
		name          string
		platformID    string
		mockError     error
		expectedError error
	}{
		{
			name:          "正常刷新访问链接",
			platformID:    "platform-1",
			mockError:     nil,
			expectedError: nil,
		},
		{
			name:          "平台不存在",
			platformID:    "non-existent",
			mockError:     assert.AnError,
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore.On("Get", tt.platformID).Return(models.GamePlatform{}, tt.mockError).Once()

			result, err := service.RefreshPlatformAccess(tt.platformID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, result.Link)
			assert.True(t, time.Now().Before(result.ExpiresAt))

			mockStore.AssertExpectations(t)
		})
	}
}
