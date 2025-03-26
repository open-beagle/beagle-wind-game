package store

import (
	"os"
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
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

// TestGameCardStore_New 测试创建存储实例
func TestGameCardStore_New(t *testing.T) {
	// 测试正常创建
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	defer store.Cleanup()

	// 测试无效文件路径
	store, err = NewGameCardStore("/invalid/path/test.yaml")
	assert.Error(t, err)
	assert.Nil(t, store)
}

// TestGameCardStore_List 测试获取所有卡片
func TestGameCardStore_List(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试空列表
	cards, err := store.List()
	assert.NoError(t, err)
	assert.Empty(t, cards)

	// 添加测试数据
	err = store.Add(testCard)
	assert.NoError(t, err)

	// 测试获取列表
	cards, err = store.List()
	assert.NoError(t, err)
	assert.Len(t, cards, 1)
	assert.Equal(t, testCard.ID, cards[0].ID)
}

// TestGameCardStore_Get 测试获取指定卡片
func TestGameCardStore_Get(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试获取不存在的卡片
	card, err := store.Get("non-existent")
	assert.Error(t, err)
	assert.Empty(t, card)

	// 添加测试数据
	err = store.Add(testCard)
	assert.NoError(t, err)

	// 测试获取存在的卡片
	card, err = store.Get(testCard.ID)
	assert.NoError(t, err)
	assert.Equal(t, testCard.ID, card.ID)
	assert.Equal(t, testCard.Name, card.Name)
}

// TestGameCardStore_Add 测试添加卡片
func TestGameCardStore_Add(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试添加卡片
	err = store.Add(testCard)
	assert.NoError(t, err)

	// 验证添加成功
	card, err := store.Get(testCard.ID)
	assert.NoError(t, err)
	assert.Equal(t, testCard.ID, card.ID)

	// 测试添加重复ID
	err = store.Add(testCard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "卡片ID已存在")
}

// TestGameCardStore_Update 测试更新卡片
func TestGameCardStore_Update(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试更新不存在的卡片
	err = store.Update(testCard)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "卡片不存在")

	// 添加测试数据
	err = store.Add(testCard)
	assert.NoError(t, err)

	// 测试更新卡片
	updatedCard := testCard
	updatedCard.Name = "Updated Card"
	err = store.Update(updatedCard)
	assert.NoError(t, err)

	// 验证更新成功
	card, err := store.Get(testCard.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Card", card.Name)
}

// TestGameCardStore_Delete 测试删除卡片
func TestGameCardStore_Delete(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试删除不存在的卡片
	err = store.Delete("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "卡片不存在")

	// 添加测试数据
	err = store.Add(testCard)
	assert.NoError(t, err)

	// 测试删除卡片
	err = store.Delete(testCard.ID)
	assert.NoError(t, err)

	// 验证删除成功
	_, err = store.Get(testCard.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "卡片不存在")
}

// TestGameCardStore_Cleanup 测试清理文件
func TestGameCardStore_Cleanup(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameCardStore(tmpFile)
	assert.NoError(t, err)

	// 添加测试数据
	err = store.Add(testCard)
	assert.NoError(t, err)

	// 测试清理文件
	err = store.Cleanup()
	assert.NoError(t, err)

	// 验证文件已删除
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err))
}
