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
var testNode = models.GameNode{
	ID:       "test-node-1",
	Name:     "Test Node",
	Model:    "test-model",
	Type:     models.GameNodeTypePhysical,
	Location: "test-location",
	Hardware: map[string]string{
		"cpu":    "4",
		"memory": "8",
	},
	Network: map[string]string{
		"ip": "192.168.1.1",
	},
	Labels: map[string]string{
		"env": "test",
	},
	Status: models.GameNodeStatus{
		State:      models.GameNodeStateOffline,
		Online:     false,
		LastOnline: time.Now(),
		UpdatedAt:  time.Now(),
		Resources: map[string]string{
			"cpu_usage": "0.5",
		},
		Metrics: map[string]interface{}{
			"latency": 100,
		},
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

// TestNodeStore_New 测试创建节点存储
func TestNodeStore_New(t *testing.T) {
	// 测试正常创建
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	defer store.Cleanup()

	// 测试无效文件路径
	store, err = NewNodeStore("/invalid/path/test.yaml")
	assert.Error(t, err)
	assert.Nil(t, store)
}

// TestNodeStore_List 测试获取所有节点
func TestNodeStore_List(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试空列表
	nodes, err := store.List()
	assert.NoError(t, err)
	assert.Empty(t, nodes)

	// 添加测试数据
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试获取列表
	nodes, err = store.List()
	assert.NoError(t, err)
	assert.Len(t, nodes, 1)
	assert.Equal(t, testNode.ID, nodes[0].ID)
}

// TestNodeStore_Get 测试获取指定节点
func TestNodeStore_Get(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试获取不存在的节点
	node, err := store.Get("non-existent")
	assert.Error(t, err)
	assert.Empty(t, node)

	// 添加测试数据
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试获取存在的节点
	node, err = store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, testNode.ID, node.ID)
	assert.Equal(t, testNode.Name, node.Name)
}

// TestNodeStore_Add 测试添加节点
func TestNodeStore_Add(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试添加节点
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 验证添加成功
	node, err := store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, testNode.ID, node.ID)

	// 测试添加重复ID
	err = store.Add(testNode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点已存在")
}

// TestNodeStore_Update 测试更新节点
func TestNodeStore_Update(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试更新不存在的节点
	err = store.Update(testNode)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")

	// 添加测试数据
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试更新节点
	updatedNode := testNode
	updatedNode.Name = "Updated Node"
	err = store.Update(updatedNode)
	assert.NoError(t, err)

	// 验证更新成功
	node, err := store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Node", node.Name)
}

// TestNodeStore_Delete 测试删除节点
func TestNodeStore_Delete(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试删除不存在的节点
	err = store.Delete("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")

	// 添加测试数据
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试删除节点
	err = store.Delete(testNode.ID)
	assert.NoError(t, err)

	// 验证删除成功
	_, err = store.Get(testNode.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "节点不存在")
}

// TestNodeStore_Cleanup 测试清理文件
func TestNodeStore_Cleanup(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)

	// 添加测试数据
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试清理文件
	err = store.Cleanup()
	assert.NoError(t, err)

	// 验证文件已删除
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err), "文件应该被删除")
}

// TestNodeStoreStatusManagement 测试节点状态管理
func TestNodeStoreStatusManagement(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	testNode := models.GameNode{
		ID:       "test-node-1",
		Name:     "Test Node",
		Model:    "test-model",
		Type:     models.GameNodeTypePhysical,
		Location: "test-location",
		Hardware: map[string]string{
			"cpu":    "4",
			"memory": "8",
		},
		Network: map[string]string{
			"ip": "192.168.1.1",
		},
		Labels: map[string]string{
			"env": "test",
		},
		Status: models.GameNodeStatus{
			State:      models.GameNodeStateOffline,
			Online:     false,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
			Resources: map[string]string{
				"cpu_usage": "0.5",
			},
			Metrics: map[string]interface{}{
				"latency": 100,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加节点
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试状态转换
	testNode.Status.State = models.GameNodeStateReady
	testNode.Status.Online = true
	err = store.Update(testNode)
	assert.NoError(t, err)

	// 验证状态更新
	node, err := store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.GameNodeStateReady, node.Status.State)
	assert.True(t, node.Status.Online)

	// 测试错误状态
	testNode.Status.State = models.GameNodeStateError
	err = store.Update(testNode)
	assert.NoError(t, err)

	// 验证错误状态
	node, err = store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, models.GameNodeStateError, node.Status.State)

	// 清理测试数据
	err = store.Delete(testNode.ID)
	assert.NoError(t, err)
}

// TestNodeStoreResourceManagement 测试节点资源管理
func TestNodeStoreResourceManagement(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	testNode := models.GameNode{
		ID:       "test-node-1",
		Name:     "Test Node",
		Model:    "test-model",
		Type:     models.GameNodeTypePhysical,
		Location: "test-location",
		Hardware: map[string]string{
			"cpu":    "4",
			"memory": "8",
		},
		Network: map[string]string{
			"ip": "192.168.1.1",
		},
		Labels: map[string]string{
			"env": "test",
		},
		Status: models.GameNodeStatus{
			State:      models.GameNodeStateReady,
			Online:     true,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
			Resources: map[string]string{
				"cpu_usage":    "0.3",
				"memory_usage": "0.4",
			},
			Metrics: map[string]interface{}{
				"latency": 100,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加节点
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试更新资源使用情况
	testNode.Status.Resources["cpu_usage"] = "0.7"
	testNode.Status.Resources["memory_usage"] = "0.8"
	err = store.Update(testNode)
	assert.NoError(t, err)

	// 验证资源更新
	node, err := store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, "0.7", node.Status.Resources["cpu_usage"])
	assert.Equal(t, "0.8", node.Status.Resources["memory_usage"])

	// 测试更新监控指标
	testNode.Status.Metrics["latency"] = 150
	testNode.Status.Metrics["packets"] = 2000
	err = store.Update(testNode)
	assert.NoError(t, err)

	// 验证指标更新
	node, err = store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, 150, node.Status.Metrics["latency"])
	assert.Equal(t, 2000, node.Status.Metrics["packets"])

	// 清理测试数据
	err = store.Delete(testNode.ID)
	assert.NoError(t, err)
}

// TestNodeStoreLabelManagement 测试节点标签管理
func TestNodeStoreLabelManagement(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewNodeStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	testNode := models.GameNode{
		ID:       "test-node-1",
		Name:     "Test Node",
		Model:    "test-model",
		Type:     models.GameNodeTypePhysical,
		Location: "test-location",
		Hardware: map[string]string{
			"cpu":    "4",
			"memory": "8",
		},
		Network: map[string]string{
			"ip": "192.168.1.1",
		},
		Labels: map[string]string{
			"env": "test",
		},
		Status: models.GameNodeStatus{
			State:      models.GameNodeStateOffline,
			Online:     false,
			LastOnline: time.Now(),
			UpdatedAt:  time.Now(),
			Resources: map[string]string{
				"cpu_usage": "0.5",
			},
			Metrics: map[string]interface{}{
				"latency": 100,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加节点
	err = store.Add(testNode)
	assert.NoError(t, err)

	// 测试添加标签
	testNode.Labels["region"] = "cn-east"
	err = store.Update(testNode)
	assert.NoError(t, err)

	// 验证标签添加
	node, err := store.Get(testNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, "cn-east", node.Labels["region"])

	// 测试删除标签
	delete(testNode.Labels, "region")
	err = store.Update(testNode)
	assert.NoError(t, err)

	// 验证标签删除
	node, err = store.Get(testNode.ID)
	assert.NoError(t, err)
	_, exists := node.Labels["region"]
	assert.False(t, exists)

	// 清理测试数据
	err = store.Delete(testNode.ID)
	assert.NoError(t, err)
}
