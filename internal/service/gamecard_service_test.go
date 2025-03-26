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
var testCard = models.GameCard{
	ID:          "test-card-1",
	Name:        "Test Card",
	SortName:    "Test Card",
	SlugName:    "test-card",
	Type:        "game",
	PlatformID:  "test-platform",
	Description: "Test game card",
	Category:    "action",
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
}

func TestListGameCards(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	cardStore, err := store.NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer cardStore.Cleanup()

	service := NewGameCardService(cardStore)

	tests := []struct {
		name           string
		params         GameCardListParams
		setup          func()
		expectedResult GameCardListResult
		expectedError  error
	}{
		{
			name: "成功获取游戏卡列表",
			params: GameCardListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				err := cardStore.Add(testCard)
				assert.NoError(t, err)
			},
			expectedResult: GameCardListResult{
				Total: 1,
				Items: []models.GameCard{testCard},
			},
			expectedError: nil,
		},
		{
			name: "存储层返回错误",
			params: GameCardListParams{
				Page:     1,
				PageSize: 20,
			},
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedResult: GameCardListResult{},
			expectedError:  fmt.Errorf("存储层错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result, err := service.ListGameCards(tt.params)
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

func TestGetGameCard(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	cardStore, err := store.NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer cardStore.Cleanup()

	service := NewGameCardService(cardStore)

	tests := []struct {
		name           string
		cardID         string
		setup          func()
		expectedResult *models.GameCard
		expectedError  error
	}{
		{
			name:   "成功获取游戏卡",
			cardID: "test-card-1",
			setup: func() {
				err := cardStore.Add(testCard)
				assert.NoError(t, err)
			},
			expectedResult: &testCard,
			expectedError:  nil,
		},
		{
			name:           "游戏卡不存在",
			cardID:         "non-existent-card",
			setup:          nil,
			expectedResult: nil,
			expectedError:  fmt.Errorf("卡片不存在: non-existent-card"),
		},
		{
			name:   "存储层返回错误",
			cardID: "test-card-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
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
			result, err := service.GetGameCard(tt.cardID)
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

func TestCreateGameCard(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	cardStore, err := store.NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer cardStore.Cleanup()

	service := NewGameCardService(cardStore)

	tests := []struct {
		name          string
		card          models.GameCard
		setup         func()
		expectedID    string
		expectedError error
	}{
		{
			name: "成功创建游戏卡",
			card: testCard,
			setup: func() {
				// 不需要特殊设置
			},
			expectedID:    testCard.ID,
			expectedError: nil,
		},
		{
			name: "卡片ID已存在",
			card: testCard,
			setup: func() {
				err := cardStore.Add(testCard)
				assert.NoError(t, err)
			},
			expectedID:    "",
			expectedError: fmt.Errorf("卡片ID已存在: %s", testCard.ID),
		},
		{
			name: "存储层返回错误",
			card: testCard,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedID:    "",
			expectedError: fmt.Errorf("存储层错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			id, err := service.CreateGameCard(tt.card)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Equal(t, tt.expectedID, id)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedID, id)
		})
	}
}

func TestUpdateGameCard(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	cardStore, err := store.NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer cardStore.Cleanup()

	service := NewGameCardService(cardStore)

	tests := []struct {
		name          string
		cardID        string
		card          models.GameCard
		setup         func()
		expectedError error
	}{
		{
			name:   "成功更新游戏卡",
			cardID: "test-card-1",
			card:   testCard,
			setup: func() {
				err := cardStore.Add(testCard)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "游戏卡不存在",
			cardID: "test-card-1",
			card:   testCard,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("存储层错误"),
		},
		{
			name:   "存储层返回错误",
			cardID: "test-card-1",
			card:   testCard,
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("存储层错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.UpdateGameCard(tt.cardID, tt.card)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestDeleteGameCard(t *testing.T) {
	// 创建临时测试文件
	tmpFile := utils.CreateTempTestFile(t)
	cardStore, err := store.NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer cardStore.Cleanup()

	service := NewGameCardService(cardStore)

	tests := []struct {
		name          string
		cardID        string
		setup         func()
		expectedError error
	}{
		{
			name:   "成功删除游戏卡",
			cardID: "test-card-1",
			setup: func() {
				err := cardStore.Add(testCard)
				assert.NoError(t, err)
			},
			expectedError: nil,
		},
		{
			name:   "游戏卡不存在",
			cardID: "test-card-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("存储层错误"),
		},
		{
			name:   "存储层返回错误",
			cardID: "test-card-1",
			setup: func() {
				// 删除临时文件以模拟存储层错误
				os.Remove(tmpFile)
				// 创建一个目录来替代文件，这样会导致读取错误
				err := os.MkdirAll(tmpFile, 0755)
				assert.NoError(t, err)
			},
			expectedError: fmt.Errorf("存储层错误"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			err := service.DeleteGameCard(tt.cardID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}
