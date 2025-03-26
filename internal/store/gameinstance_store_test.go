package store

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
	"github.com/stretchr/testify/assert"
)

// 测试数据
var testInstance = models.GameInstance{
	ID:          "test-instance-1",
	NodeID:      "test-node-1",
	PlatformID:  "test-platform-1",
	CardID:      "test-card-1",
	Status:      "stopped",
	Resources:   `{"cpu": "2", "memory": "4"}`,
	Performance: `{"cpu_usage": "0.5", "latency": 100}`,
	SaveData:    `{}`,
	Config:      `{"resolution": "1920x1080", "fps": 60}`,
	Backup:      `{}`,
	CreatedAt:   time.Now(),
	UpdatedAt:   time.Now(),
	StartedAt:   time.Now(),
	StoppedAt:   time.Now(),
}

// TestGameInstanceStore 测试实例存储
func TestGameInstanceStore(t *testing.T) {
	// 创建存储实例
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	resources := map[string]string{"cpu": "2", "memory": "4"}
	resourcesJSON, _ := json.Marshal(resources)

	config := map[string]string{"port": "8080"}
	configJSON, _ := json.Marshal(config)

	testInstance := models.GameInstance{
		ID:          "test-instance-1",
		NodeID:      "test-node-1",
		PlatformID:  "test-platform-1",
		CardID:      "test-card-1",
		Status:      "running",
		Resources:   string(resourcesJSON),
		Performance: "{}",
		SaveData:    "{}",
		Config:      string(configJSON),
		Backup:      "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		StartedAt:   time.Now(),
		StoppedAt:   time.Time{},
	}

	// 测试添加实例
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试获取实例
	instance, err := store.Get(testInstance.ID)
	assert.NoError(t, err)
	assert.Equal(t, testInstance.ID, instance.ID)
	assert.Equal(t, testInstance.Status, instance.Status)

	// 测试更新实例
	testInstance.Status = "stopped"
	err = store.Update(testInstance)
	assert.NoError(t, err)

	// 验证更新
	instance, err = store.Get(testInstance.ID)
	assert.NoError(t, err)
	assert.Equal(t, "stopped", instance.Status)

	// 测试删除实例
	err = store.Delete(testInstance.ID)
	assert.NoError(t, err)

	// 验证删除
	_, err = store.Get(testInstance.ID)
	assert.Error(t, err)
}

// TestGameInstanceStoreFindByNodeID 测试按节点ID查找实例
func TestGameInstanceStoreFindByNodeID(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	nodeID := "test-node-1"
	instances := []models.GameInstance{
		{
			ID:          "instance-1",
			NodeID:      nodeID,
			PlatformID:  "platform-1",
			CardID:      "card-1",
			Status:      "running",
			Resources:   "{}",
			Performance: "{}",
			SaveData:    "{}",
			Config:      "{}",
			Backup:      "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			StartedAt:   time.Now(),
			StoppedAt:   time.Time{},
		},
		{
			ID:          "instance-2",
			NodeID:      nodeID,
			PlatformID:  "platform-2",
			CardID:      "card-2",
			Status:      "running",
			Resources:   "{}",
			Performance: "{}",
			SaveData:    "{}",
			Config:      "{}",
			Backup:      "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			StartedAt:   time.Now(),
			StoppedAt:   time.Time{},
		},
	}

	// 添加测试实例
	for _, instance := range instances {
		err := store.Add(instance)
		assert.NoError(t, err)
	}

	// 测试按节点ID查找
	found, err := store.FindByNodeID(nodeID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(found))

	// 验证找到的实例
	for _, instance := range found {
		assert.Equal(t, nodeID, instance.NodeID)
	}

	// 清理测试数据
	for _, instance := range instances {
		err := store.Delete(instance.ID)
		assert.NoError(t, err)
	}
}

// TestGameInstanceStoreFindByCardID 测试按卡片ID查找实例
func TestGameInstanceStoreFindByCardID(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	cardID := "test-card-1"
	instances := []models.GameInstance{
		{
			ID:          "instance-1",
			NodeID:      "node-1",
			PlatformID:  "platform-1",
			CardID:      cardID,
			Status:      "running",
			Resources:   "{}",
			Performance: "{}",
			SaveData:    "{}",
			Config:      "{}",
			Backup:      "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			StartedAt:   time.Now(),
			StoppedAt:   time.Time{},
		},
		{
			ID:          "instance-2",
			NodeID:      "node-2",
			PlatformID:  "platform-2",
			CardID:      cardID,
			Status:      "running",
			Resources:   "{}",
			Performance: "{}",
			SaveData:    "{}",
			Config:      "{}",
			Backup:      "{}",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			StartedAt:   time.Now(),
			StoppedAt:   time.Time{},
		},
	}

	// 添加测试实例
	for _, instance := range instances {
		err := store.Add(instance)
		assert.NoError(t, err)
	}

	// 测试按卡片ID查找
	found, err := store.FindByCardID(cardID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(found))

	// 验证找到的实例
	for _, instance := range found {
		assert.Equal(t, cardID, instance.CardID)
	}

	// 清理测试数据
	for _, instance := range instances {
		err := store.Delete(instance.ID)
		assert.NoError(t, err)
	}
}

// TestGameInstanceStoreStateTransition 测试实例状态转换
func TestGameInstanceStoreStateTransition(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试实例
	testInstance := models.GameInstance{
		ID:          "test-instance-1",
		NodeID:      "test-node-1",
		PlatformID:  "test-platform-1",
		CardID:      "test-card-1",
		Status:      "created",
		Resources:   "{}",
		Performance: "{}",
		SaveData:    "{}",
		Config:      "{}",
		Backup:      "{}",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		StartedAt:   time.Time{},
		StoppedAt:   time.Time{},
	}

	// 添加实例
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试状态转换
	testInstance.Status = "running"
	err = store.Update(testInstance)
	assert.NoError(t, err)

	// 验证状态更新
	instance, err := store.Get(testInstance.ID)
	assert.NoError(t, err)
	assert.Equal(t, "running", instance.Status)

	// 清理测试数据
	err = store.Delete(testInstance.ID)
	assert.NoError(t, err)
}

// TestGameInstanceStore_New 测试创建实例存储
func TestGameInstanceStore_New(t *testing.T) {
	// 测试正常创建
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	defer store.Cleanup()

	// 测试无效文件路径
	store, err = NewGameInstanceStore("/invalid/path/test.yaml")
	assert.Error(t, err)
	assert.Nil(t, store)
}

// TestGameInstanceStore_List 测试获取所有实例
func TestGameInstanceStore_List(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试空列表
	instances, err := store.List()
	assert.NoError(t, err)
	assert.Empty(t, instances)

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试获取列表
	instances, err = store.List()
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, testInstance.ID, instances[0].ID)
}

// TestGameInstanceStore_Get 测试获取指定实例
func TestGameInstanceStore_Get(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试获取不存在的实例
	instance, err := store.Get("non-existent")
	assert.Error(t, err)
	assert.Empty(t, instance)

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试获取存在的实例
	instance, err = store.Get(testInstance.ID)
	assert.NoError(t, err)
	assert.Equal(t, testInstance.ID, instance.ID)
	assert.Equal(t, testInstance.NodeID, instance.NodeID)
}

// TestGameInstanceStore_Add 测试添加实例
func TestGameInstanceStore_Add(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试添加实例
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 验证添加成功
	instance, err := store.Get(testInstance.ID)
	assert.NoError(t, err)
	assert.Equal(t, testInstance.ID, instance.ID)

	// 测试添加重复ID
	err = store.Add(testInstance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "实例ID已存在")
}

// TestGameInstanceStore_Update 测试更新实例
func TestGameInstanceStore_Update(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试更新不存在的实例
	err = store.Update(testInstance)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "实例不存在")

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试更新实例
	updatedInstance := testInstance
	updatedInstance.Status = "running"
	err = store.Update(updatedInstance)
	assert.NoError(t, err)

	// 验证更新成功
	instance, err := store.Get(testInstance.ID)
	assert.NoError(t, err)
	assert.Equal(t, "running", instance.Status)
}

// TestGameInstanceStore_Delete 测试删除实例
func TestGameInstanceStore_Delete(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试删除不存在的实例
	err = store.Delete("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "实例不存在")

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试删除实例
	err = store.Delete(testInstance.ID)
	assert.NoError(t, err)

	// 验证删除成功
	_, err = store.Get(testInstance.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "实例不存在")
}

// TestGameInstanceStore_FindByNodeID 测试按节点ID查找实例
func TestGameInstanceStore_FindByNodeID(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试查找不存在的节点ID
	instances, err := store.FindByNodeID("non-existent")
	assert.NoError(t, err)
	assert.Empty(t, instances)

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试查找存在的节点ID
	instances, err = store.FindByNodeID(testInstance.NodeID)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, testInstance.ID, instances[0].ID)
}

// TestGameInstanceStore_FindByCardID 测试按卡片ID查找实例
func TestGameInstanceStore_FindByCardID(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试查找不存在的卡片ID
	instances, err := store.FindByCardID("non-existent")
	assert.NoError(t, err)
	assert.Empty(t, instances)

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试查找存在的卡片ID
	instances, err = store.FindByCardID(testInstance.CardID)
	assert.NoError(t, err)
	assert.Len(t, instances, 1)
	assert.Equal(t, testInstance.ID, instances[0].ID)
}

// TestGameInstanceStore_Cleanup 测试清理文件
func TestGameInstanceStore_Cleanup(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewGameInstanceStore(tmpFile)
	assert.NoError(t, err)

	// 添加测试数据
	err = store.Add(testInstance)
	assert.NoError(t, err)

	// 测试清理文件
	err = store.Cleanup()
	assert.NoError(t, err)

	// 验证文件已删除
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err))
}
