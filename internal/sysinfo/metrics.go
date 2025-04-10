package sysinfo

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// IgnorePath 检查挂载点是否应该被忽略
func IgnorePath(path string) bool {
	// 忽略特殊文件系统和系统挂载点
	return path == "" ||
		strings.HasPrefix(path, "/proc") ||
		strings.HasPrefix(path, "/sys") ||
		strings.HasPrefix(path, "/dev") ||
		strings.HasPrefix(path, "/run") ||
		strings.HasPrefix(path, "/tmp") ||
		strings.Contains(path, "overlay")
}

// MetricsCollector 系统指标采集器
type MetricsCollector struct {
	logger           *logrus.Logger
	lastCPUTime      cpu.TimesStat
	lastNetworkStats map[string]net.IOCountersStat
}

// NewMetricsCollector 创建新的系统指标采集器
func NewMetricsCollector(logger *logrus.Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:           logger,
		lastNetworkStats: make(map[string]net.IOCountersStat),
	}
}

// GetMetricsInfo 获取系统指标信息
func (c *MetricsCollector) GetMetricsInfo(hardwareInfo *models.HardwareInfo) (models.MetricsInfo, error) {
	var metricsInfo models.MetricsInfo

	// 检查硬件信息是否有效
	if hardwareInfo == nil {
		return metricsInfo, fmt.Errorf("硬件信息未提供")
	}

	// 采集CPU使用情况
	if err := c.collectCPUMetrics(hardwareInfo, &metricsInfo); err != nil {
		c.logger.Error("采集CPU指标失败: %v", err)
		return metricsInfo, fmt.Errorf("采集CPU指标失败: %v", err)
	}

	// 采集内存使用情况
	if err := c.collectMemoryMetrics(hardwareInfo, &metricsInfo); err != nil {
		c.logger.Error("采集内存指标失败: %v", err)
		return metricsInfo, fmt.Errorf("采集内存指标失败: %v", err)
	}

	// 采集GPU使用情况
	if err := c.collectGPUMetrics(hardwareInfo, &metricsInfo); err != nil {
		// GPU可能不存在，这不是严重错误
		c.logger.Warn("采集GPU指标失败: %v", err)
	}

	// 采集存储设备使用情况
	if err := c.collectStorageMetrics(hardwareInfo, &metricsInfo); err != nil {
		c.logger.Error("采集存储指标失败: %v", err)
		return metricsInfo, fmt.Errorf("采集存储指标失败: %v", err)
	}

	// 采集网络使用情况
	if err := c.collectNetworkMetrics(&metricsInfo); err != nil {
		c.logger.Error("采集网络指标失败: %v", err)
		return metricsInfo, fmt.Errorf("采集网络指标失败: %v", err)
	}

	return metricsInfo, nil
}

// collectCPUMetrics 采集CPU使用情况
func (c *MetricsCollector) collectCPUMetrics(hardwareInfo *models.HardwareInfo, metricsInfo *models.MetricsInfo) error {
	// 获取CPU使用率
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return fmt.Errorf("获取CPU使用率失败: %v", err)
	}

	// 获取CPU信息
	if hardwareInfo != nil && len(hardwareInfo.CPUs) > 0 {
		metricsInfo.CPUs = append(metricsInfo.CPUs, models.CPUMetrics{
			Model:   hardwareInfo.CPUs[0].Model,
			Cores:   hardwareInfo.CPUs[0].Cores,
			Threads: hardwareInfo.CPUs[0].Threads,
			Usage:   percent[0],
		})
	}

	return nil
}

// collectMemoryMetrics 采集内存使用情况
func (c *MetricsCollector) collectMemoryMetrics(hardwareInfo *models.HardwareInfo, metricsInfo *models.MetricsInfo) error {
	// 获取内存信息
	mem, err := mem.VirtualMemory()
	if err != nil {
		return fmt.Errorf("获取内存信息失败: %v", err)
	}

	metricsInfo.Memory = models.MemoryMetrics{
		Total:     int64(mem.Total),
		Available: int64(mem.Available),
		Used:      int64(mem.Used),
		Usage:     mem.UsedPercent,
	}

	return nil
}

// collectGPUMetrics 收集GPU监控指标
func (c *MetricsCollector) collectGPUMetrics(hardwareInfo *models.HardwareInfo, metricsInfo *models.MetricsInfo) error {
	// 使用我们的工具函数查找nvidia-smi
	nvidiaSmiPath, err := findCommand("nvidia-smi")
	if err != nil {
		return fmt.Errorf("未找到nvidia-smi命令: %v", err)
	}

	// 执行nvidia-smi命令获取监控指标
	cmd := exec.Command(nvidiaSmiPath, "--query-gpu=index,name,memory.total,memory.used,memory.free,utilization.gpu", "--format=csv,noheader,nounits")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行nvidia-smi命令失败: %v, 输出: %s", err, string(out))
	}

	// 解析输出
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// 解析CSV格式的输出
		fields := strings.Split(line, ",")
		if len(fields) < 6 {
			continue
		}

		// 清理字段
		for i := range fields {
			fields[i] = strings.TrimSpace(fields[i])
		}

		// 转换指标值
		memoryTotal, _ := strconv.ParseUint(fields[2], 10, 64)
		memoryUsed, _ := strconv.ParseUint(fields[3], 10, 64)
		memoryFree, _ := strconv.ParseUint(fields[4], 10, 64)

		// 改进GPU使用率的解析
		gpuUtilStr := strings.TrimSpace(fields[5])
		gpuUtil, err := strconv.ParseFloat(gpuUtilStr, 64)
		if err != nil {
			c.logger.Warn("解析GPU使用率失败: %v, 原始值: %s", err, gpuUtilStr)
			gpuUtil = 0
		}

		// 计算实际的显存使用率
		memoryUsage := 0.0
		if memoryTotal > 0 {
			memoryUsage = float64(memoryUsed) / float64(memoryTotal) * 100.0
		}

		// 创建GPU指标对象
		gpuMetrics := models.GPUMetrics{
			Model:       fields[1],
			MemoryTotal: int64(memoryTotal), // 保持MB单位
			GPUUsage:    gpuUtil,
			MemoryUsed:  int64(memoryUsed), // 保持MB单位
			MemoryFree:  int64(memoryFree), // 保持MB单位
			MemoryUsage: memoryUsage,       // 使用实际计算的显存使用率
		}

		// 添加到指标信息
		metricsInfo.GPUs = append(metricsInfo.GPUs, gpuMetrics)
	}

	return nil
}

// collectStorageMetrics 采集存储设备指标
func (c *MetricsCollector) collectStorageMetrics(hardwareInfo *models.HardwareInfo, metricsInfo *models.MetricsInfo) error {
	// 遍历硬件信息中的存储设备
	for _, storage := range hardwareInfo.Storages {
		// 跳过没有挂载点的设备
		if storage.Path == "" {
			continue
		}

		// 使用df命令获取存储设备使用情况
		cmd := exec.Command("df", "-B1", storage.Path) // 使用1字节为单位
		out, err := cmd.Output()
		if err != nil {
			c.logger.Warn("获取存储设备 %s 使用情况失败: %v", storage.Path, err)
			continue
		}

		// 解析df输出
		lines := strings.Split(string(out), "\n")
		if len(lines) < 2 {
			continue
		}

		// 跳过标题行，解析数据行
		fields := strings.Fields(lines[1])
		if len(fields) < 6 {
			continue
		}

		// 解析容量信息（以字节为单位）
		total, _ := strconv.ParseInt(fields[1], 10, 64)
		used, _ := strconv.ParseInt(fields[2], 10, 64)
		free, _ := strconv.ParseInt(fields[3], 10, 64)

		// 计算使用率
		usage := float64(0)
		if total > 0 {
			usage = float64(used) / float64(total) * 100
		}

		// 添加到指标
		metricsInfo.Storages = append(metricsInfo.Storages, models.StorageMetrics{
			Path:     storage.Path,
			Type:     storage.Type,
			Model:    storage.Model,
			Capacity: total,
			Used:     used,
			Free:     free,
			Usage:    usage,
		})
	}

	return nil
}

// collectNetworkMetrics 采集网络使用情况
func (c *MetricsCollector) collectNetworkMetrics(metricsInfo *models.MetricsInfo) error {
	// 读取/proc/net/dev获取网络接口流量信息
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		c.logger.Error("读取网络信息失败: %v", err)
		return fmt.Errorf("读取网络信息失败: %v", err)
	}

	lines := strings.Split(string(data), "\n")

	// 获取网络流量
	var totalRxBytes, totalTxBytes int64
	var connectionCount int32

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		// 处理接口名（去掉冒号）
		ifaceName := strings.TrimRight(fields[0], ":")

		// 跳过loopback和虚拟接口
		if ifaceName == "lo" || strings.HasPrefix(ifaceName, "vir") || strings.HasPrefix(ifaceName, "docker") || strings.HasPrefix(ifaceName, "br-") {
			continue
		}

		// 解析接收和发送字节数
		if len(fields) >= 10 {
			rxBytes, _ := strconv.ParseInt(fields[1], 10, 64)
			txBytes, _ := strconv.ParseInt(fields[9], 10, 64)

			totalRxBytes += rxBytes
			totalTxBytes += txBytes
		}
	}

	// 获取连接数
	out, err := exec.Command("ss", "-s").Output()
	if err == nil {
		// 解析连接数
		re := regexp.MustCompile(`TCP: +(\d+) \(estab \d+`)
		matches := re.FindStringSubmatch(string(out))
		if len(matches) > 1 {
			if val, err := strconv.ParseInt(matches[1], 10, 32); err == nil {
				connectionCount = int32(val)
			}
		}
	}

	// 计算流量（Mbps）
	// 这里简化处理，如果需要准确计算流量，应该进行两次采样并计算差值
	inboundTraffic := float64(totalRxBytes) / 1024.0 / 1024.0 * 8.0 / 10.0  // 假设数据是10秒的累积
	outboundTraffic := float64(totalTxBytes) / 1024.0 / 1024.0 * 8.0 / 10.0 // 假设数据是10秒的累积

	// 添加到指标信息 - 只添加有效值
	if inboundTraffic > 0 || outboundTraffic > 0 || connectionCount > 0 {
		metricsInfo.Network = models.NetworkMetrics{
			InboundTraffic:  inboundTraffic,
			OutboundTraffic: outboundTraffic,
			Connections:     connectionCount,
		}
	}

	return nil
}

// CollectStorageMetrics 收集存储设备指标
func CollectStorageMetrics() ([]models.StorageMetrics, error) {
	var storageMetrics []models.StorageMetrics

	// 创建日志器
	logger := utils.New("StorageMetrics")

	// 使用更简单的方法获取分区信息，不依赖gopsutil
	out, err := exec.Command("df", "-B1", "--output=source,target,fstype,size,used,avail,pcent").Output()
	if err != nil {
		logger.Error("执行df命令失败: %v", err)
		return nil, fmt.Errorf("执行df命令失败: %v", err)
	}

	// 最小有效存储大小 - 100MB
	const minValidStorageSize int64 = 100 * 1024 * 1024
	// 相似性阈值 - 5%
	const similarityThreshold float64 = 0.05

	// 使用map进行去重
	deviceMap := make(map[string]models.StorageMetrics)

	// 解析输出
	lines := strings.Split(string(out), "\n")
	for i, line := range lines {
		if i == 0 || len(strings.TrimSpace(line)) == 0 {
			continue // 跳过表头和空行
		}

		fields := strings.Fields(line)
		if len(fields) >= 7 {
			source := fields[0]
			mountpoint := fields[1]
			fstype := fields[2]

			// 跳过特定的挂载点
			if IgnorePath(mountpoint) {
				continue
			}

			size, _ := strconv.ParseInt(fields[3], 10, 64)

			// 跳过太小的设备
			if size < minValidStorageSize {
				continue
			}

			used, _ := strconv.ParseInt(fields[4], 10, 64)
			free, _ := strconv.ParseInt(fields[5], 10, 64)

			// 解析使用率
			usage := 0.0
			usageStr := fields[6]
			if strings.HasSuffix(usageStr, "%") {
				usageValue := strings.TrimSuffix(usageStr, "%")
				if val, err := strconv.ParseFloat(usageValue, 64); err == nil {
					usage = val
				}
			}

			// 确定设备类型
			deviceType := "Physical"
			if strings.Contains(fstype, "tmpfs") ||
				strings.Contains(fstype, "devtmpfs") ||
				strings.Contains(fstype, "overlay") {
				deviceType = "Virtual"
			}

			// 创建指标
			metrics := models.StorageMetrics{
				Path:     mountpoint,
				Usage:    usage,
				Used:     used,
				Free:     free,
				Capacity: size,
				Type:     deviceType,
				Model:    source, // 默认使用source作为Model
			}

			// WSL环境下的Windows驱动器检测
			isWslDrive := false
			if runtime.GOOS == "linux" &&
				strings.HasPrefix(mountpoint, "/mnt/") &&
				len(mountpoint) >= 6 {
				driveLetter := mountpoint[5:6]
				if strings.Contains("abcdefghijklmnopqrstuvwxyz", strings.ToLower(driveLetter)) {
					isWslDrive = true
					metrics.Type = "Virtual"

					// 设置更具描述性的类型
					if driveLetter == "c" {
						metrics.Model = "Windows System Drive"
					} else {
						metrics.Model = fmt.Sprintf("Windows %s Drive", strings.ToUpper(driveLetter))
					}
				}
			}

			// 设备去重逻辑
			deviceKey := ""
			if isWslDrive {
				driveLetter := mountpoint[5:6]
				deviceKey = fmt.Sprintf("wsl-%s", strings.ToUpper(driveLetter))
			} else {
				// 对于常规设备，使用容量范围作为键
				sizeGB := float64(metrics.Capacity) / (1024 * 1024 * 1024)
				// 舍入到最接近的10GB
				sizeRounded := math.Round(sizeGB/10) * 10
				deviceKey = fmt.Sprintf("%.0fGB-%s", sizeRounded, metrics.Type)
			}

			// 检查map中是否已存在类似设备
			if existing, exists := deviceMap[deviceKey]; exists {
				// 如果已存在类似设备，保留路径更短的（通常是主挂载点）
				// 或容量更大的设备
				if (len(metrics.Path) < len(existing.Path) &&
					metrics.Capacity >= int64(float64(existing.Capacity)*0.95)) ||
					metrics.Capacity > int64(float64(existing.Capacity)*1.05) {
					deviceMap[deviceKey] = metrics
				}
			} else {
				deviceMap[deviceKey] = metrics
			}
		}
	}

	// 将map中的设备转换为切片
	for _, device := range deviceMap {
		storageMetrics = append(storageMetrics, device)
	}

	// 按容量排序（降序）
	sort.Slice(storageMetrics, func(i, j int) bool {
		return storageMetrics[i].Capacity > storageMetrics[j].Capacity
	})

	return storageMetrics, nil
}

// getBasicCPUInfo 从/proc/cpuinfo获取基本CPU信息
func getBasicCPUInfo() ([]struct {
	Model   string
	Cores   int32
	Threads int32
}, error) {
	var result []struct {
		Model   string
		Cores   int32
		Threads int32
	}

	cpuinfoData, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return nil, err
	}

	cpuinfoLines := strings.Split(string(cpuinfoData), "\n")

	var cpuModel string
	var cpuCores int32
	var cpuThreads int32
	processorCount := 0

	for _, line := range cpuinfoLines {
		if strings.HasPrefix(line, "processor") {
			processorCount++
		} else if strings.HasPrefix(line, "model name") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				cpuModel = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "cpu cores") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if val, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32); err == nil {
					cpuCores = int32(val)
				}
			}
		}
	}

	// 设置线程数为处理器数
	cpuThreads = int32(processorCount)

	// 如果无法获取核心数，但有线程数，估算核心数
	if cpuCores == 0 && cpuThreads > 0 {
		// 假设超线程，核心数约为线程数的一半
		cpuCores = cpuThreads / 2
		if cpuCores == 0 {
			cpuCores = 1 // 至少有1个核心
		}
	}

	// 防止返回空值
	if cpuModel == "" {
		cpuModel = "Unknown CPU"
	}

	// 添加到结果
	result = append(result, struct {
		Model   string
		Cores   int32
		Threads int32
	}{
		Model:   cpuModel,
		Cores:   cpuCores,
		Threads: cpuThreads,
	})

	return result, nil
}
