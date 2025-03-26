package utils

import (
	"os"
	"testing"
)

// CreateTempTestFile 创建临时测试文件
func CreateTempTestFile(t *testing.T) string {
	tmpFile, err := os.CreateTemp("", "beagle-wind-game_test_*.yaml")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	return tmpFile.Name()
}
