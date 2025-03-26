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
var testPlatform = models.GamePlatform{
	ID:       "test-platform-1",
	Name:     "Test Platform",
	Version:  "1.0.0",
	Type:     "test-type",
	OS:       "test-os",
	Image:    "test-image",
	Bin:      "test-bin",
	Features: []string{"feature1", "feature2"},
	Config:   map[string]string{"key": "value"},
	Files:    []models.PlatformFile{{ID: "file1", Type: "bin", URL: "http://test.com/file1"}},
	Installer: []models.PlatformInstaller{
		{
			Command: "install",
			Move:    &models.InstallerMove{Src: "src", Dst: "dst"},
			Chmodx:  "file",
			Extract: &models.InstallerExtract{File: "archive", Dst: "dst"},
		},
	},
	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),
}

// TestPlatformStore_New 测试创建平台存储
func TestPlatformStore_New(t *testing.T) {
	// 测试正常创建
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	defer store.Cleanup()

	// 测试无效文件路径
	store, err = NewPlatformStore("/invalid/path/test.yaml")
	assert.Error(t, err)
	assert.Nil(t, store)
}

// TestPlatformStore_List 测试获取所有平台
func TestPlatformStore_List(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试空列表
	platforms, err := store.List()
	assert.NoError(t, err)
	assert.Empty(t, platforms)

	// 添加测试数据
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试获取列表
	platforms, err = store.List()
	assert.NoError(t, err)
	assert.Len(t, platforms, 1)
	assert.Equal(t, testPlatform.ID, platforms[0].ID)
}

// TestPlatformStore_Get 测试获取指定平台
func TestPlatformStore_Get(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试获取不存在的平台
	platform, err := store.Get("non-existent")
	assert.Error(t, err)
	assert.Empty(t, platform)

	// 添加测试数据
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试获取存在的平台
	platform, err = store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Equal(t, testPlatform.ID, platform.ID)
	assert.Equal(t, testPlatform.Name, platform.Name)
}

// TestPlatformStore_Add 测试添加平台
func TestPlatformStore_Add(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试添加平台
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 验证添加成功
	platform, err := store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Equal(t, testPlatform.ID, platform.ID)

	// 测试添加重复ID
	err = store.Add(testPlatform)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "平台ID已存在")
}

// TestPlatformStore_Update 测试更新平台
func TestPlatformStore_Update(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试更新不存在的平台
	err = store.Update(testPlatform)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "平台不存在")

	// 添加测试数据
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试更新平台
	updatedPlatform := testPlatform
	updatedPlatform.Name = "Updated Platform"
	err = store.Update(updatedPlatform)
	assert.NoError(t, err)

	// 验证更新成功
	platform, err := store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Platform", platform.Name)
}

// TestPlatformStore_Delete 测试删除平台
func TestPlatformStore_Delete(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 测试删除不存在的平台
	err = store.Delete("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "平台不存在")

	// 添加测试数据
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试删除平台
	err = store.Delete(testPlatform.ID)
	assert.NoError(t, err)

	// 验证删除成功
	_, err = store.Get(testPlatform.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "平台不存在")
}

// TestPlatformStore_Cleanup 测试清理文件
func TestPlatformStore_Cleanup(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)

	// 添加测试数据
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试清理文件
	err = store.Cleanup()
	assert.NoError(t, err)

	// 验证文件已删除
	_, err = os.Stat(tmpFile)
	assert.True(t, os.IsNotExist(err))
}

// TestPlatformStore 测试平台存储
func TestPlatformStore(t *testing.T) {
	// 创建存储实例
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	testPlatform := models.GamePlatform{
		ID:        "test-platform-1",
		Name:      "Test Platform",
		Type:      "emulator",
		Version:   "1.0.0",
		OS:        "linux",
		Image:     "test-image:latest",
		Bin:       "/usr/local/bin/test",
		Features:  []string{"feature1", "feature2"},
		Config:    map[string]string{"key1": "value1"},
		Files:     []models.PlatformFile{},
		Installer: []models.PlatformInstaller{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 测试添加平台
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试获取平台
	platform, err := store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Equal(t, testPlatform.ID, platform.ID)
	assert.Equal(t, testPlatform.Name, platform.Name)

	// 测试更新平台
	testPlatform.Name = "Updated Platform"
	err = store.Update(testPlatform)
	assert.NoError(t, err)

	// 验证更新
	platform, err = store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Platform", platform.Name)

	// 测试删除平台
	err = store.Delete(testPlatform.ID)
	assert.NoError(t, err)

	// 验证删除
	_, err = store.Get(testPlatform.ID)
	assert.Error(t, err)
}

// TestPlatformStoreVersionManagement 测试平台版本管理
func TestPlatformStoreVersionManagement(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	platformID := "test-platform-1"
	platform := models.GamePlatform{
		ID:        platformID,
		Name:      "Test Platform",
		Type:      "emulator",
		Version:   "1.0.0",
		OS:        "linux",
		Image:     "test-image:1.0.0",
		Bin:       "/usr/local/bin/test",
		Features:  []string{"feature1", "feature2"},
		Config:    map[string]string{"key1": "value1"},
		Files:     []models.PlatformFile{},
		Installer: []models.PlatformInstaller{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加初始版本
	err = store.Add(platform)
	assert.NoError(t, err)

	// 更新到新版本
	platform.Version = "1.1.0"
	platform.Image = "test-image:1.1.0"
	platform.Features = append(platform.Features, "feature3")
	platform.Config["key2"] = "value2"
	err = store.Update(platform)
	assert.NoError(t, err)

	// 验证版本更新
	updated, err := store.Get(platformID)
	assert.NoError(t, err)
	assert.Equal(t, "1.1.0", updated.Version)
	assert.Equal(t, "test-image:1.1.0", updated.Image)
	assert.Contains(t, updated.Features, "feature3")
	assert.Equal(t, "value2", updated.Config["key2"])

	// 清理测试数据
	err = store.Delete(platformID)
	assert.NoError(t, err)
}

// TestPlatformStoreConfigManagement 测试平台配置管理
func TestPlatformStoreConfigManagement(t *testing.T) {
	tmpFile := utils.CreateTempTestFile(t)
	store, err := NewPlatformStore(tmpFile)
	assert.NoError(t, err)
	defer store.Cleanup()

	// 创建测试数据
	testPlatform := models.GamePlatform{
		ID:        "test-platform-1",
		Name:      "Test Platform",
		Type:      "emulator",
		Version:   "1.0.0",
		OS:        "linux",
		Image:     "test-image:latest",
		Bin:       "/usr/local/bin/test",
		Features:  []string{"feature1", "feature2"},
		Config:    map[string]string{"key1": "value1"},
		Files:     []models.PlatformFile{},
		Installer: []models.PlatformInstaller{},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 添加平台
	err = store.Add(testPlatform)
	assert.NoError(t, err)

	// 测试更新配置
	testPlatform.Config["key2"] = "value2"
	err = store.Update(testPlatform)
	assert.NoError(t, err)

	// 验证配置更新
	platform, err := store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Equal(t, "value2", platform.Config["key2"])

	// 测试更新特性
	testPlatform.Features = append(testPlatform.Features, "feature3")
	err = store.Update(testPlatform)
	assert.NoError(t, err)

	// 验证特性更新
	platform, err = store.Get(testPlatform.ID)
	assert.NoError(t, err)
	assert.Contains(t, platform.Features, "feature3")

	// 清理测试数据
	err = store.Delete(testPlatform.ID)
	assert.NoError(t, err)
}
