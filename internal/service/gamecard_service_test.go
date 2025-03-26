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

// mockGameCardErrorStore 模拟返回错误的游戏卡存储实现
type mockGameCardErrorStore struct{}

func (s *mockGameCardErrorStore) List() ([]models.GameCard, error) {
	return nil, assert.AnError
}

func (s *mockGameCardErrorStore) Get(id string) (models.GameCard, error) {
	return models.GameCard{}, assert.AnError
}

func (s *mockGameCardErrorStore) Add(card models.GameCard) error {
	return assert.AnError
}

func (s *mockGameCardErrorStore) Update(card models.GameCard) error {
	return assert.AnError
}

func (s *mockGameCardErrorStore) Delete(id string) error {
	return assert.AnError
}

func (s *mockGameCardErrorStore) Cleanup() error {
	return nil
}

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

func TestGameCardService_List(t *testing.T) {
	tests := []struct {
		name           string
		params         GameCardListParams
		store          store.GameCardStore
		expectedResult *GameCardListResult
		expectedError  error
	}{
		{
			name: "成功获取游戏卡列表",
			params: GameCardListParams{
				Page:     1,
				PageSize: 20,
			},
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testCard)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &GameCardListResult{
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
			store:          &mockGameCardErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameCardService(tt.store)
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

func TestGameCardService_Get(t *testing.T) {
	tests := []struct {
		name           string
		cardID         string
		store          store.GameCardStore
		expectedResult *models.GameCard
		expectedError  error
	}{
		{
			name:   "成功获取游戏卡",
			cardID: "test-card-1",
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testCard)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: &testCard,
			expectedError:  nil,
		},
		{
			name:   "游戏卡不存在",
			cardID: "non-existent-card",
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedResult: nil,
			expectedError:  fmt.Errorf("卡片不存在: non-existent-card"),
		},
		{
			name:           "存储层返回错误",
			cardID:         "test-card-1",
			store:          &mockGameCardErrorStore{},
			expectedResult: nil,
			expectedError:  assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameCardService(tt.store)
			result, err := service.Get(tt.cardID)
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

func TestGameCardService_Create(t *testing.T) {
	tests := []struct {
		name          string
		card          models.GameCard
		store         store.GameCardStore
		expectedID    string
		expectedError error
	}{
		{
			name: "成功创建游戏卡",
			card: testCard,
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedID:    testCard.ID,
			expectedError: nil,
		},
		{
			name: "卡片ID已存在",
			card: testCard,
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testCard)
				assert.NoError(t, err)
				return store
			}(),
			expectedID:    "",
			expectedError: fmt.Errorf("卡片ID已存在: %s", testCard.ID),
		},
		{
			name:          "存储层返回错误",
			card:          testCard,
			store:         &mockGameCardErrorStore{},
			expectedID:    "",
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameCardService(tt.store)
			id, err := service.Create(tt.card)
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

func TestGameCardService_Update(t *testing.T) {
	updatedCard := testCard
	updatedCard.Name = "Updated Card"

	tests := []struct {
		name          string
		cardID        string
		card          models.GameCard
		store         store.GameCardStore
		expectedError error
	}{
		{
			name:   "成功更新游戏卡",
			cardID: "test-card-1",
			card:   updatedCard,
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testCard)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:   "游戏卡不存在",
			cardID: "non-existent-card",
			card:   updatedCard,
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("卡片不存在: non-existent-card"),
		},
		{
			name:          "存储层返回错误",
			cardID:        "test-card-1",
			card:          updatedCard,
			store:         &mockGameCardErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameCardService(tt.store)
			err := service.Update(tt.cardID, tt.card)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGameCardService_Delete(t *testing.T) {
	tests := []struct {
		name          string
		cardID        string
		store         store.GameCardStore
		expectedError error
	}{
		{
			name:   "成功删除游戏卡",
			cardID: "test-card-1",
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				err = store.Add(testCard)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: nil,
		},
		{
			name:   "游戏卡不存在",
			cardID: "non-existent-card",
			store: func() store.GameCardStore {
				tmpFile := utils.CreateTempTestFile(t)
				store, err := store.NewGameCardStore(tmpFile)
				assert.NoError(t, err)
				return store
			}(),
			expectedError: fmt.Errorf("卡片不存在: non-existent-card"),
		},
		{
			name:          "存储层返回错误",
			cardID:        "test-card-1",
			store:         &mockGameCardErrorStore{},
			expectedError: assert.AnError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewGameCardService(tt.store)
			err := service.Delete(tt.cardID)
			if tt.expectedError != nil {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
