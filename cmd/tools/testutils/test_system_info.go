package main

import (
	"encoding/json"
	"fmt"

	"github.com/open-beagle/beagle-wind-game/internal/sysinfo"
)

func main() {
	// 创建系统信息收集器
	collector := sysinfo.NewSystemCollector(nil)

	// 获取系统信息
	info, err := collector.GetSystemInfo()
	if err != nil {
		fmt.Printf("获取系统信息失败: %v\n", err)
		return
	}

	// 打印系统信息
	fmt.Println("=== 系统信息 ===")
	fmt.Printf("操作系统: %s %s (%s)\n", info.OSDistribution, info.OSVersion, info.OSArchitecture)
	fmt.Printf("内核版本: %s\n", info.KernelVersion)
	fmt.Printf("GPU 驱动: %s\n", info.GPUDriverVersion)
	fmt.Printf("计算框架: %s\n", info.GPUComputeAPIVersion)

	// 打印容器运行时信息
	fmt.Println("\n=== 容器运行时 ===")
	fmt.Printf("Docker版本: %s\n", info.DockerVersion)
	fmt.Printf("Containerd版本: %s\n", info.ContainerdVersion)
	fmt.Printf("Runc版本: %s\n", info.RuncVersion)

	// 以JSON格式输出完整信息
	fmt.Println("\n=== 完整JSON信息 ===")
	jsonData, _ := json.MarshalIndent(info, "", "  ")
	fmt.Println(string(jsonData))
}
