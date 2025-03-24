package store

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Store 数据存储接口
type Store interface {
	// Get 获取指定ID的数据
	Get(id string, result interface{}) error
	// List 获取所有数据
	List(result interface{}) error
	// Save 保存数据
	Save(id string, data interface{}) error
	// Delete 删除数据
	Delete(id string) error
}

// YAMLStore YAML文件存储实现
type YAMLStore struct {
	dataDir   string        // 数据目录
	dataFile  string        // 数据文件名
	dataCache []interface{} // 缓存数据
	mu        sync.RWMutex
}

// NewYAMLStore 创建一个新的YAML文件存储
func NewYAMLStore(dataDir, dataFile string) (*YAMLStore, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	store := &YAMLStore{
		dataDir:  dataDir,
		dataFile: dataFile,
	}

	// 初始化数据文件
	err := store.initDataFile()
	if err != nil {
		return nil, err
	}

	return store, nil
}

// 获取数据文件完整路径
func (s *YAMLStore) getDataFilePath() string {
	return filepath.Join(s.dataDir, s.dataFile)
}

// 初始化数据文件
func (s *YAMLStore) initDataFile() error {
	filePath := s.getDataFilePath()

	// 判断文件是否存在
	_, err := os.Stat(filePath)
	if err == nil {
		// 文件存在，读取数据
		return s.loadData()
	}

	if os.IsNotExist(err) {
		// 文件不存在，创建空数据文件
		s.dataCache = []interface{}{}
		return s.saveData()
	}

	return fmt.Errorf("检查数据文件失败: %w", err)
}

// 加载数据
func (s *YAMLStore) loadData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := s.getDataFilePath()
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取数据文件失败: %w", err)
	}

	var items []interface{}
	err = yaml.Unmarshal(data, &items)
	if err != nil {
		return fmt.Errorf("解析YAML数据失败: %w", err)
	}

	s.dataCache = items
	return nil
}

// 保存数据
func (s *YAMLStore) saveData() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := s.getDataFilePath()
	data, err := yaml.Marshal(s.dataCache)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入数据文件失败: %w", err)
	}

	return nil
}

// Get 获取指定ID的数据
func (s *YAMLStore) Get(id string, result interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 在缓存中查找数据
	for _, item := range s.dataCache {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemID, ok := itemMap["id"]; ok && itemID == id {
				// 找到匹配的数据
				data, err := yaml.Marshal(item)
				if err != nil {
					return fmt.Errorf("序列化数据失败: %w", err)
				}

				err = yaml.Unmarshal(data, result)
				if err != nil {
					return fmt.Errorf("解析数据失败: %w", err)
				}

				return nil
			}
		}
	}

	return fmt.Errorf("数据不存在: %s", id)
}

// List 获取所有数据
func (s *YAMLStore) List(result interface{}) error {
	err := s.loadData() // 刷新数据
	if err != nil {
		return err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// 将缓存数据序列化后反序列化到结果
	data, err := yaml.Marshal(s.dataCache)
	if err != nil {
		return fmt.Errorf("序列化数据失败: %w", err)
	}

	err = yaml.Unmarshal(data, result)
	if err != nil {
		return fmt.Errorf("解析数据失败: %w", err)
	}

	return nil
}

// Save 保存数据
func (s *YAMLStore) Save(id string, data interface{}) error {
	// 先加载数据确保最新
	err := s.loadData()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查是否已存在相同ID的数据
	found := false
	for i, item := range s.dataCache {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemID, ok := itemMap["id"]; ok && itemID == id {
				// 更新现有数据
				s.dataCache[i] = data
				found = true
				break
			}
		}
	}

	// 如果不存在，添加新数据
	if !found {
		s.dataCache = append(s.dataCache, data)
	}

	// 保存回文件
	return s.saveData()
}

// Delete 删除数据
func (s *YAMLStore) Delete(id string) error {
	// 先加载数据确保最新
	err := s.loadData()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// 查找并删除数据
	found := false
	newCache := make([]interface{}, 0, len(s.dataCache))

	for _, item := range s.dataCache {
		if itemMap, ok := item.(map[string]interface{}); ok {
			if itemID, ok := itemMap["id"]; ok && itemID == id {
				found = true
				continue // 跳过这条记录
			}
		}
		newCache = append(newCache, item)
	}

	if !found {
		return fmt.Errorf("数据不存在: %s", id)
	}

	s.dataCache = newCache

	// 保存回文件
	return s.saveData()
}
