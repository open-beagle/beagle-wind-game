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
	"syscall"

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

// MetricsCollector 指标信息采集器
type MetricsCollector struct {
	// 可能的配置选项
	options map[string]string
	// 日志框架
	logger utils.Logger
}

// NewMetricsCollector 创建新的指标信息采集器
func NewMetricsCollector(options map[string]string) *MetricsCollector {
	if options == nil {
		options = make(map[string]string)
	}

	// 创建日志器
	logger := utils.New("MetricsCollector")

	// 尝试查找常用GPU命令
	for _, cmd := range []string{"nvidia-smi", "rocm-smi", "intel_gpu_top"} {
		// 仅当用户没有指定自定义路径时自动查找
		optKey := strings.Replace(cmd, "-", "_", -1) + "_path"
		if _, exists := options[optKey]; !exists {
			if path, err := findCommand(cmd); err == nil {
				options[optKey] = path
			}
		}
	}

	return &MetricsCollector{
		options: options,
		logger:  logger,
	}
}

// GetMetricsInfo 获取系统指标信息
func (c *MetricsCollector) GetMetricsInfo() (models.MetricsInfo, error) {
	var metricsInfo models.MetricsInfo

	// 采集CPU使用情况
	if err := c.collectCPUMetrics(&metricsInfo); err != nil {
		c.logger.Error("采集CPU指标失败: %v", err)
		return metricsInfo, fmt.Errorf("采集CPU指标失败: %v", err)
	}

	// 采集内存使用情况
	if err := c.collectMemoryMetrics(&metricsInfo); err != nil {
		c.logger.Error("采集内存指标失败: %v", err)
		return metricsInfo, fmt.Errorf("采集内存指标失败: %v", err)
	}

	// 采集GPU使用情况
	if err := c.collectGPUMetrics(&metricsInfo); err != nil {
		// GPU可能不存在，这不是严重错误
		c.logger.Warn("采集GPU指标失败: %v", err)
	}

	// 采集存储设备使用情况
	if err := c.collectStorageMetrics(&metricsInfo); err != nil {
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
func (c *MetricsCollector) collectCPUMetrics(metricsInfo *models.MetricsInfo) error {
	// 读取/proc/stat获取CPU使用情况
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		c.logger.Error("读取CPU信息失败: %v", err)
		return fmt.Errorf("读取CPU信息失败: %v", err)
	}

	// 解析CPU使用率
	cpuUsage := getCPUUsage(string(data))

	// 获取CPU信息
	cpuInfo, err := getCPUInfo()
	if err != nil {
		c.logger.Error("获取CPU详细信息失败: %v", err)
		return fmt.Errorf("获取CPU详细信息失败: %v", err)
	}

	// 遍历CPU设备，添加指标
	for _, device := range cpuInfo {
		metricsInfo.CPUs = append(metricsInfo.CPUs, models.CPUMetrics{
			Model:   device.Model,
			Cores:   device.Cores,
			Threads: device.Threads,
			Usage:   cpuUsage,
		})
	}

	return nil
}

// getCPUInfo 获取CPU信息
func getCPUInfo() ([]struct {
	Model   string
	Cores   int32
	Threads int32
}, error) {
	var cpuDevices []struct {
		Model   string
		Cores   int32
		Threads int32
	}

	// 使用lscpu命令获取CPU信息
	out, err := exec.Command("lscpu").Output()
	if err != nil {
		return nil, fmt.Errorf("执行lscpu命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	var model string
	var cores int32
	var threads int32

	for _, line := range lines {
		if strings.Contains(line, "Model name:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				model = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "CPU(s):") && !strings.Contains(line, "NUMA") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if val, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32); err == nil {
					threads = int32(val)
				}
			}
		} else if strings.Contains(line, "Core(s) per socket:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if val, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32); err == nil {
					cores = int32(val)
				}
			}
		} else if strings.Contains(line, "Socket(s):") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if val, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 32); err == nil {
					// 确保核心数是每个插槽的核心数乘以插槽数
					if cores > 0 {
						cores = cores * int32(val)
					}
				}
			}
		}
	}

	// 如果无法获取核心数，但有线程数，估算核心数
	if cores == 0 && threads > 0 {
		// 假设超线程，核心数约为线程数的一半
		cores = threads / 2
		if cores == 0 {
			cores = 1 // 至少有1个核心
		}
	}

	// 如果无法获取线程数，但有核心数，估算线程数
	if threads == 0 && cores > 0 {
		// 假设每个核心有2个线程
		threads = cores * 2
	}

	// 防止返回空值
	if model == "" {
		// 尝试从/proc/cpuinfo获取
		cpuinfo, err := os.ReadFile("/proc/cpuinfo")
		if err == nil {
			lines := strings.Split(string(cpuinfo), "\n")
			for _, line := range lines {
				if strings.Contains(line, "model name") {
					parts := strings.Split(line, ":")
					if len(parts) > 1 {
						model = strings.TrimSpace(parts[1])
						break
					}
				}
			}
		}

		// 仍然没有获取到，使用一个通用名称
		if model == "" {
			model = "Unknown CPU"
		}
	}

	// 添加到CPU设备列表
	cpuDevices = append(cpuDevices, struct {
		Model   string
		Cores   int32
		Threads int32
	}{
		Model:   model,
		Cores:   cores,
		Threads: threads,
	})

	return cpuDevices, nil
}

// getCPUUsage 解析/proc/stat获取CPU使用率
func getCPUUsage(data string) float64 {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "cpu ") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				return 0.0
			}

			// 解析CPU时间片，按 user nice system idle iowait 等顺序
			user, _ := strconv.ParseInt(fields[1], 10, 64)
			nice, _ := strconv.ParseInt(fields[2], 10, 64)
			system, _ := strconv.ParseInt(fields[3], 10, 64)
			idle, _ := strconv.ParseInt(fields[4], 10, 64)

			// 计算CPU总时间和非空闲时间
			total := user + nice + system + idle
			nonIdle := user + nice + system

			// 计算使用率
			if total > 0 {
				return float64(nonIdle) / float64(total) * 100.0
			}
		}
	}
	return 0.0 // 默认返回0
}

// collectMemoryMetrics 采集内存使用情况
func (c *MetricsCollector) collectMemoryMetrics(metricsInfo *models.MetricsInfo) error {
	var memoryStat *models.MemoryMetrics
	var err error

	// 尝试从free命令获取内存信息
	memoryStat, err = c.getMemoryFromFree()
	if err != nil {
		// 尝试从/proc/meminfo获取
		memoryStat, err = c.getMemoryFromProcMeminfo()
		if err != nil {
			// 尝试从sysinfo获取
			memoryStat, err = c.getMemoryFromSysinfo()
			if err != nil {
				return fmt.Errorf("无法获取内存信息: %v", err)
			}
		}
	}

	// 只在获取到有效数据时设置内存指标
	if memoryStat != nil && memoryStat.Total > 0 {
		metricsInfo.Memory = *memoryStat
	}

	return nil
}

// isValidMemoryStat 检查内存统计信息是否有效
func (c *MetricsCollector) isValidMemoryStat(stat *models.MemoryMetrics) bool {
	if stat == nil {
		return false
	}

	// 检查Total是否有效（大于100MB）
	if stat.Total < 100*1024*1024 {
		return false
	}

	// 确保Used和Available不为零
	if stat.Used <= 0 || stat.Available <= 0 {
		return false
	}

	return true
}

// getMemoryFromFree 使用free命令获取内存信息
func (c *MetricsCollector) getMemoryFromFree() (*models.MemoryMetrics, error) {
	// 使用free命令获取内存使用情况
	out, err := exec.Command("free", "-b").Output()
	if err != nil {
		return nil, fmt.Errorf("执行free命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Mem:") {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}

			// 解析内存值
			total, totalErr := strconv.ParseInt(fields[1], 10, 64)
			used, usedErr := strconv.ParseInt(fields[2], 10, 64)
			available := int64(0)

			// 获取可用内存
			if len(fields) >= 7 {
				// 在Linux系统上，第7列是可用内存(available)
				available, _ = strconv.ParseInt(fields[6], 10, 64)
			} else if len(fields) >= 4 {
				// 如果没有available列，使用free列
				free, freeErr := strconv.ParseInt(fields[3], 10, 64)
				if freeErr == nil {
					available = free
				}
			}

			// 若available有误差，确保它不超过total
			if available > total {
				available = total
			}

			// 计算使用率
			var usage float64 = 0
			if totalErr == nil && total > 0 {
				if usedErr == nil {
					usage = float64(used) / float64(total) * 100
				} else if available > 0 {
					usage = float64(total-available) / float64(total) * 100
				}
			}

			// 确保有合理的值
			if totalErr != nil || total <= 0 {
				return nil, fmt.Errorf("无法从free命令获取有效的内存总量")
			}

			if used <= 0 && available > 0 {
				used = total - available
			}

			if available <= 0 && used > 0 {
				available = total - used
			}

			// 确保不返回零值
			if available <= 0 {
				available = total / 10 // 假设至少有10%的内存可用
			}

			if used <= 0 {
				used = total / 10 // 假设至少有10%的内存被使用
			}

			return &models.MemoryMetrics{
				Total:     total,
				Used:      used,
				Available: available,
				Usage:     usage,
			}, nil
		}
	}

	return nil, fmt.Errorf("无法从free命令输出中解析内存信息")
}

// getMemoryFromProcMeminfo 从/proc/meminfo文件读取内存信息
func (c *MetricsCollector) getMemoryFromProcMeminfo() (*models.MemoryMetrics, error) {
	// 读取/proc/meminfo文件
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("读取/proc/meminfo失败: %v", err)
	}

	var total, free, available, buffers, cached int64

	// 解析内存信息
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				total, _ = strconv.ParseInt(fields[1], 10, 64)
				total *= 1024 // 转换为字节（meminfo中的单位是KB）
			}
		} else if strings.HasPrefix(line, "MemFree:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				free, _ = strconv.ParseInt(fields[1], 10, 64)
				free *= 1024
			}
		} else if strings.HasPrefix(line, "MemAvailable:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				available, _ = strconv.ParseInt(fields[1], 10, 64)
				available *= 1024
			}
		} else if strings.HasPrefix(line, "Buffers:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				buffers, _ = strconv.ParseInt(fields[1], 10, 64)
				buffers *= 1024
			}
		} else if strings.HasPrefix(line, "Cached:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				cached, _ = strconv.ParseInt(fields[1], 10, 64)
				cached *= 1024
			}
		}
	}

	// 如果total为0，这是一个严重错误
	if total <= 0 {
		return nil, fmt.Errorf("无法从/proc/meminfo获取有效的内存总量")
	}

	// 如果MemAvailable字段不可用（老版本内核）
	if available <= 0 {
		available = free + buffers + cached
	}

	// 计算已用内存
	used := total - free

	// 如果available大于total，纠正它
	if available > total {
		available = total
	}

	// 确保不返回零值
	if available <= 0 {
		available = total / 10 // 假设至少有10%的内存可用
	}

	if used <= 0 {
		used = total / 10 // 假设至少有10%的内存被使用
	}

	// 计算使用率
	var usage float64 = 0
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return &models.MemoryMetrics{
		Total:     total,
		Used:      used,
		Available: available,
		Usage:     usage,
	}, nil
}

// getMemoryFromSysinfo 使用syscall.Sysinfo获取内存信息
func (c *MetricsCollector) getMemoryFromSysinfo() (*models.MemoryMetrics, error) {
	var info syscall.Sysinfo_t
	err := syscall.Sysinfo(&info)
	if err != nil {
		c.logger.Error("调用syscall.Sysinfo失败: %v", err)
		return nil, fmt.Errorf("调用syscall.Sysinfo失败: %v", err)
	}

	// 在一些系统上可能需要通过unit_multiplier来调整单位
	// 参见: https://man7.org/linux/man-pages/man2/sysinfo.2.html
	multiplier := uint64(info.Unit) // sysinfo unit in bytes
	if multiplier == 0 {
		multiplier = 1 // 默认为1字节
	}

	// 计算内存值（字节）
	total := uint64(info.Totalram) * multiplier
	free := uint64(info.Freeram) * multiplier
	// 注意：syscall.Sysinfo在较旧的Linux系统上不提供MemAvailable
	// 我们使用Freeram+Bufferram作为近似值
	available := (uint64(info.Freeram) + uint64(info.Bufferram)) * multiplier
	used := total - free

	// 如果available大于total，纠正它
	if available > total {
		available = total
	}

	// 确保不返回零值
	if available <= 0 {
		available = total / 10 // 假设至少有10%的内存可用
	}

	if used <= 0 {
		used = total / 10 // 假设至少有10%的内存被使用
	}

	// 计算使用率
	var usage float64 = 0
	if total > 0 {
		usage = float64(used) / float64(total) * 100
	}

	return &models.MemoryMetrics{
		Total:     int64(total),
		Used:      int64(used),
		Available: int64(available),
		Usage:     usage,
	}, nil
}

// collectGPUMetrics 采集GPU使用情况
func (c *MetricsCollector) collectGPUMetrics(metricsInfo *models.MetricsInfo) error {
	// 尝试NVIDIA GPU
	if err := c.collectNvidiaGPUMetrics(metricsInfo); err != nil {
		// 尝试AMD GPU
		if err := c.collectAMDGPUMetrics(metricsInfo); err != nil {
			// 尝试Intel GPU
			if err := c.collectIntelGPUMetrics(metricsInfo); err != nil {
				c.logger.Warn("未检测到GPU或无法获取GPU指标: %v", err)
				return fmt.Errorf("未检测到GPU或无法获取GPU指标: %v", err)
			}
		}
	}

	return nil
}

// collectNvidiaGPUMetrics 采集NVIDIA GPU使用情况
func (c *MetricsCollector) collectNvidiaGPUMetrics(metricsInfo *models.MetricsInfo) error {
	// 从配置中获取自定义路径
	var nvidiaSmiPath string
	var err error

	if customPath, ok := c.options["nvidia_smi_path"]; ok && customPath != "" {
		// 使用配置中指定的路径
		nvidiaSmiPath = customPath
	} else {
		// 使用通用的查找函数
		nvidiaSmiPath, err = findCommand("nvidia-smi")
		if err != nil {
			c.logger.Debug("未找到nvidia-smi命令: %v", err)
			return fmt.Errorf("未找到nvidia-smi命令: %v", err)
		}
	}

	// 执行nvidia-smi命令获取GPU使用情况，同时获取模型信息
	cmd := fmt.Sprintf("%s --query-gpu=name,utilization.gpu,utilization.memory,memory.total,memory.used,memory.free --format=csv,noheader,nounits", nvidiaSmiPath)
	out, err := execCommand("sh", "-c", cmd)
	if err != nil {
		c.logger.Error("执行nvidia-smi命令失败: %v", err)
		return fmt.Errorf("执行nvidia-smi命令失败: %v", err)
	}

	// 解析输出
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) >= 5 {
			// 创建GPU监控指标
			metric := models.GPUMetrics{}

			// 获取GPU型号
			if len(fields) >= 1 {
				metric.Model = strings.TrimSpace(fields[0])
			}

			// 获取GPU利用率
			if len(fields) >= 2 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64); err == nil {
					metric.Usage = val
				}
			}

			// 获取显存利用率
			if len(fields) >= 3 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64); err == nil {
					metric.MemoryUsage = val
				}
			}

			// 获取显存总量
			if len(fields) >= 4 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64); err == nil {
					metric.MemoryTotal = int64(val * 1024 * 1024) // 转换为字节
				}
			}

			// 获取已用显存
			if len(fields) >= 5 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64); err == nil {
					metric.MemoryUsed = int64(val * 1024 * 1024) // 转换为字节
				}
			}

			// 获取可用显存
			if len(fields) >= 6 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(fields[5]), 64); err == nil {
					metric.MemoryFree = int64(val * 1024 * 1024) // 转换为字节
				}
			}

			// 只添加有有效数据的GPU指标
			if metric.Model != "" || metric.Usage > 0 || metric.MemoryTotal > 0 {
				metricsInfo.GPUs = append(metricsInfo.GPUs, metric)
			}
		}
	}

	return nil
}

// collectAMDGPUMetrics 收集AMD GPU指标
func (c *MetricsCollector) collectAMDGPUMetrics(metricsInfo *models.MetricsInfo) error {
	// 从配置中获取自定义路径
	var rocmSmiPath string
	var err error

	if customPath, ok := c.options["rocm_smi_path"]; ok && customPath != "" {
		// 使用配置中指定的路径
		rocmSmiPath = customPath
	} else {
		// 使用通用的查找函数
		rocmSmiPath, err = findCommand("rocm-smi")
		if err != nil {
			c.logger.Debug("未找到rocm-smi命令: %v", err)
			return fmt.Errorf("未找到rocm-smi命令，请确保AMD ROCm驱动已正确安装或在配置中指定rocm_smi_path: %v", err)
		}
	}

	// 执行rocm-smi命令获取GPU使用情况
	_, err = execCommand(rocmSmiPath, "--showuse", "--showmemuse", "--showtemp")
	if err != nil {
		c.logger.Error("执行rocm-smi命令失败: %v", err)
		return fmt.Errorf("执行rocm-smi命令失败: %v", err)
	}

	// 解析AMD GPU指标
	// 这里简化处理，实际应根据rocm-smi输出格式进行解析
	// ...

	return nil
}

// collectIntelGPUMetrics 收集Intel GPU指标
func (c *MetricsCollector) collectIntelGPUMetrics(metricsInfo *models.MetricsInfo) error {
	// 从配置中获取自定义路径
	var intelGpuTopPath string
	var err error

	if customPath, ok := c.options["intel_gpu_top_path"]; ok && customPath != "" {
		// 使用配置中指定的路径
		intelGpuTopPath = customPath
	} else {
		// 使用通用的查找函数
		intelGpuTopPath, err = findCommand("intel_gpu_top")
		if err != nil {
			c.logger.Debug("未找到intel_gpu_top命令: %v", err)
			return fmt.Errorf("未找到intel_gpu_top命令，请确保Intel GPU工具已正确安装或在配置中指定intel_gpu_top_path: %v", err)
		}
	}

	// 执行intel_gpu_top命令获取GPU使用情况
	_, err = execCommand(intelGpuTopPath, "-o", "-J")
	if err != nil {
		c.logger.Error("执行intel_gpu_top命令失败: %v", err)
		return fmt.Errorf("执行intel_gpu_top命令失败: %v", err)
	}

	// 解析Intel GPU指标
	// 这里简化处理，实际应根据intel_gpu_top输出格式进行解析
	// ...

	return nil
}

// collectStorageMetrics 采集存储设备使用情况
func (c *MetricsCollector) collectStorageMetrics(metricsInfo *models.MetricsInfo) error {
	// 使用df命令获取存储设备使用情况
	out, err := exec.Command("df", "-B1", "--output=source,target,fstype,size,used,avail,pcent").Output()
	if err != nil {
		c.logger.Error("执行df命令失败: %v", err)
		return fmt.Errorf("执行df命令失败: %v", err)
	}

	// 定义最小有效存储大小（100MB），过滤掉太小的设备
	const minValidStorageSize int64 = 100 * 1024 * 1024

	// 用于存储有效设备的临时映射，以防重复
	// key格式: "路径" 例如: "/mnt/c"
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
			path := fields[1]
			fstype := fields[2]

			// 跳过特殊文件系统和小型临时文件系统
			if source == "tmpfs" || source == "devtmpfs" ||
				strings.HasPrefix(source, "/dev/loop") ||
				strings.HasPrefix(source, "overlay") ||
				strings.Contains(path, "/proc") ||
				strings.Contains(path, "/sys") ||
				strings.Contains(path, "/run") ||
				strings.Contains(path, "/dev") ||
				strings.Contains(path, "/tmp") {
				continue
			}

			size, _ := strconv.ParseInt(fields[3], 10, 64)

			// 跳过太小的存储设备（可能是临时挂载点）
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

			// 存储类型判断
			storageType := "HDD" // 默认为HDD
			model := source      // 默认使用设备路径作为型号

			// 尝试获取更详细的存储设备信息
			if strings.HasPrefix(source, "/dev/") {
				// 从设备名称判断类型
				deviceName := strings.TrimPrefix(source, "/dev/")
				if strings.HasPrefix(deviceName, "nvme") {
					storageType = "SSD"
					model = "NVMe SSD"
				} else if strings.HasPrefix(deviceName, "sd") {
					// 尝试判断是SSD还是HDD
					rotationalPath := fmt.Sprintf("/sys/block/%s/queue/rotational", deviceName)
					rotationalData, err := os.ReadFile(rotationalPath)
					if err == nil {
						rotational := strings.TrimSpace(string(rotationalData))
						if rotational == "0" {
							storageType = "SSD"
						}
					}

					// 尝试获取型号信息
					modelPath := fmt.Sprintf("/sys/block/%s/device/model", deviceName)
					modelData, err := os.ReadFile(modelPath)
					if err == nil {
						deviceModel := strings.TrimSpace(string(modelData))
						if deviceModel != "" {
							model = deviceModel
						}
					}
				}
			} else if strings.Contains(path, "/mnt/") && (fstype == "drvfs" || strings.Contains(fstype, "9p") || strings.Contains(fstype, "wslfs")) {
				// WSL环境下的Windows盘符挂载
				storageType = "Virtual"
				model = "Windows Drive"

				// 检测是Windows系统盘还是数据盘
				if strings.Contains(path, "/mnt/c") {
					model = "Windows System Drive"
				} else if len(path) > 6 { // 确保有足够的字符
					// 提取盘符
					driveLetter := strings.ToUpper(path[5:6]) // 获取第6个字符作为盘符
					model = fmt.Sprintf("Windows %s Drive", driveLetter)
				}
			}

			// 创建存储指标
			deviceMetric := models.StorageMetrics{
				Path:     path,
				Type:     storageType,
				Model:    model,
				Capacity: size,
				Used:     used,
				Free:     free,
				Usage:    usage,
			}

			// 使用路径作为键
			deviceMap[path] = deviceMetric
		}
	}

	// 将deviceMap中的设备添加到指标信息中
	for _, device := range deviceMap {
		metricsInfo.Storages = append(metricsInfo.Storages, device)
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
