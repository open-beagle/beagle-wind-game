package store

import (
	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockInstanceStore 模拟实例存储
type MockInstanceStore struct {
	mock.Mock
}

func (m *MockInstanceStore) List() ([]models.GameInstance, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GameInstance), args.Error(1)
}

func (m *MockInstanceStore) Get(id string) (models.GameInstance, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return models.GameInstance{}, args.Error(1)
	}
	return args.Get(0).(models.GameInstance), args.Error(1)
}

func (m *MockInstanceStore) Add(instance models.GameInstance) error {
	args := m.Called(instance)
	return args.Error(0)
}

func (m *MockInstanceStore) Update(instance models.GameInstance) error {
	args := m.Called(instance)
	return args.Error(0)
}

func (m *MockInstanceStore) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInstanceStore) FindByNodeID(nodeID string) ([]models.GameInstance, error) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GameInstance), args.Error(1)
}

func (m *MockInstanceStore) FindByCardID(cardID string) ([]models.GameInstance, error) {
	args := m.Called(cardID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GameInstance), args.Error(1)
}

// MockNodeStore 模拟节点存储
type MockNodeStore struct {
	mock.Mock
}

func (m *MockNodeStore) List() ([]models.GameNode, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GameNode), args.Error(1)
}

func (m *MockNodeStore) Get(id string) (models.GameNode, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return models.GameNode{}, args.Error(1)
	}
	return args.Get(0).(models.GameNode), args.Error(1)
}

func (m *MockNodeStore) Add(node models.GameNode) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockNodeStore) Update(node models.GameNode) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockNodeStore) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockGameCardStore 模拟游戏卡片存储
type MockGameCardStore struct {
	mock.Mock
}

func (m *MockGameCardStore) List() ([]models.GameCard, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GameCard), args.Error(1)
}

func (m *MockGameCardStore) Get(id string) (models.GameCard, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return models.GameCard{}, args.Error(1)
	}
	return args.Get(0).(models.GameCard), args.Error(1)
}

func (m *MockGameCardStore) Add(card models.GameCard) error {
	args := m.Called(card)
	return args.Error(0)
}

func (m *MockGameCardStore) Update(card models.GameCard) error {
	args := m.Called(card)
	return args.Error(0)
}

func (m *MockGameCardStore) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// MockPlatformStore 模拟平台存储
type MockPlatformStore struct {
	mock.Mock
}

func (m *MockPlatformStore) List() ([]models.GamePlatform, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.GamePlatform), args.Error(1)
}

func (m *MockPlatformStore) Get(id string) (models.GamePlatform, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return models.GamePlatform{}, args.Error(1)
	}
	return args.Get(0).(models.GamePlatform), args.Error(1)
}

func (m *MockPlatformStore) Add(platform models.GamePlatform) error {
	args := m.Called(platform)
	return args.Error(0)
}

func (m *MockPlatformStore) Update(platform models.GamePlatform) error {
	args := m.Called(platform)
	return args.Error(0)
}

func (m *MockPlatformStore) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
