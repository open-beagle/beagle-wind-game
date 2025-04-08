package sysinfo

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// 常见命令路径列表，根据常见安装位置设置
	commonCommandPaths = []string{
		"/usr/bin",
		"/usr/local/bin",
		"/bin",
		"/sbin",
		"/usr/sbin",
		"/usr/local/sbin",
		"/opt/bin",
		"/usr/lib/wsl/lib",    // WSL中的NVIDIA命令
		"/usr/local/cuda/bin", // CUDA路径
		"/opt/rocm/bin",       // ROCm路径
	}
)

// findCommand 在多个路径下查找命令
func findCommand(name string) (string, error) {
	// 首先尝试使用exec.LookPath（依赖环境变量PATH）
	path, err := exec.LookPath(name)
	if err == nil {
		return path, nil
	}

	// 如果在PATH中找不到，尝试在常见路径中查找
	for _, dir := range commonCommandPaths {
		fullPath := filepath.Join(dir, name)
		if _, err := os.Stat(fullPath); err == nil {
			// 文件存在
			return fullPath, nil
		}
	}

	// 全局搜索一些关键命令（代价较高，仅用于关键命令）
	if name == "nvidia-smi" || name == "rocm-smi" || name == "intel_gpu_top" {
		foundPaths, _ := findCommandInSystem(name)
		if len(foundPaths) > 0 {
			return foundPaths[0], nil
		}
	}

	return "", fmt.Errorf("找不到命令: %s", name)
}

// findCommandInSystem 在常见系统目录中递归查找指定命令
func findCommandInSystem(name string) ([]string, error) {
	var results []string
	searchDirs := []string{"/usr", "/opt"}

	for _, dir := range searchDirs {
		// 检查该目录是否存在
		if _, err := os.Stat(dir); err != nil {
			continue
		}

		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			// 跳过错误
			if err != nil {
				return nil
			}

			// 只检查文件
			if !info.IsDir() && info.Name() == name {
				// 检查是否可执行
				if info.Mode()&0111 != 0 {
					results = append(results, path)
				}
			}

			// 如果结果已经足够，提前退出搜索
			if len(results) > 0 {
				return filepath.SkipDir
			}

			return nil
		})

		if err == nil && len(results) > 0 {
			break
		}
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("在系统中找不到命令: %s", name)
	}

	return results, nil
}

// execCommand 执行命令并返回输出
func execCommand(name string, args ...string) ([]byte, error) {
	cmdPath, err := findCommand(name)
	if err != nil {
		return nil, err
	}

	return exec.Command(cmdPath, args...).Output()
}

// checkCommand 检查命令是否可用
func checkCommand(name string) error {
	_, err := findCommand(name)
	if err != nil {
		return err
	}
	return nil
}

// getScriptPath 获取用于复杂操作的脚本路径
func getScriptPath(scriptName string) (string, error) {
	// 检查临时目录
	tempDir := os.TempDir()
	scriptPath := filepath.Join(tempDir, scriptName)

	// 创建脚本
	if scriptName == "gpu_detect.sh" {
		script := `#!/bin/bash
# 检测GPU信息的脚本
if command -v nvidia-smi &> /dev/null; then
    echo "NVIDIA:$(nvidia-smi --query-gpu=name,memory.total,pci.bus_id,uuid --format=csv,noheader)"
fi

if command -v rocm-smi &> /dev/null; then
    echo "AMD:$(rocm-smi --showproductname --showmeminfo vram --showbus --showhwid)"
fi

if command -v lspci &> /dev/null; then
    echo "INTEL:$(lspci | grep -i 'VGA\\|3D\\|Display' | grep -i 'Intel')"
fi

exit 0
`
		err := os.WriteFile(scriptPath, []byte(script), 0755)
		if err != nil {
			return "", err
		}
	}

	return scriptPath, nil
}

// runWithEnvPath 使用额外的PATH环境变量运行命令
func runWithEnvPath(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)

	// 获取当前环境变量
	env := os.Environ()
	path := getEnvValue(env, "PATH")

	// 添加我们的常见路径到PATH
	var additionalPaths []string
	for _, dir := range commonCommandPaths {
		if !strings.Contains(path, dir) {
			additionalPaths = append(additionalPaths, dir)
		}
	}

	if len(additionalPaths) > 0 {
		newPath := path + ":" + strings.Join(additionalPaths, ":")
		setEnvValue(env, "PATH", newPath)
		cmd.Env = env
	}

	return cmd.Output()
}

// getEnvValue 获取环境变量的值
func getEnvValue(env []string, key string) string {
	prefix := key + "="
	for _, envVar := range env {
		if strings.HasPrefix(envVar, prefix) {
			return envVar[len(prefix):]
		}
	}
	return ""
}

// setEnvValue 设置环境变量的值
func setEnvValue(env []string, key, value string) {
	prefix := key + "="
	for i, envVar := range env {
		if strings.HasPrefix(envVar, prefix) {
			env[i] = prefix + value
			return
		}
	}
	env = append(env, prefix+value)
}
