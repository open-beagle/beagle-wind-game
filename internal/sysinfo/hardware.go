package sysinfo

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/open-beagle/beagle-wind-game/internal/models"
)

// HardwareCollector 硬件信息采集器
type HardwareCollector struct {
	// 可能的配置选项
	options map[string]string
}

// NewHardwareCollector 创建新的硬件信息采集器
func NewHardwareCollector(options map[string]string) *HardwareCollector {
	if options == nil {
		options = make(map[string]string)
	}
	return &HardwareCollector{
		options: options,
	}
}

// GetHardwareInfo 采集硬件信息
func (c *HardwareCollector) GetHardwareInfo() (models.HardwareInfo, error) {
	// 初始化扁平结构
	hardwareInfo := models.HardwareInfo{
		CPUs:     []models.CPUDevice{},
		Memories: []models.MemoryDevice{},
		GPUs:     []models.GPUDevice{},
		Storages: []models.StorageDevice{},
		Networks: []models.NetworkDevice{},
	}

	// 采集CPU信息
	if err := c.collectCPUInfo(&hardwareInfo); err != nil {
		fmt.Printf("采集CPU信息失败: %v\n", err)
	}

	// 采集内存信息
	if err := c.collectMemoryInfo(&hardwareInfo); err != nil {
		fmt.Printf("采集内存信息失败: %v\n", err)
	}

	// 采集GPU信息
	if err := c.collectGPUInfo(&hardwareInfo); err != nil {
		fmt.Printf("采集GPU信息失败: %v\n", err)
	}

	// 采集存储设备信息
	if err := c.collectStorageInfo(&hardwareInfo); err != nil {
		fmt.Printf("采集存储设备信息失败: %v\n", err)
	}

	// 采集网络设备信息
	if err := c.collectNetworkInfo(&hardwareInfo); err != nil {
		fmt.Printf("采集网络设备信息失败: %v\n", err)
	}

	return hardwareInfo, nil
}

// GetSimplifiedHardwareInfo 获取简化版硬件信息（键值对格式）
func (c *HardwareCollector) GetSimplifiedHardwareInfo() (map[string]string, error) {
	hardwareInfo, err := c.GetHardwareInfo()
	if err != nil {
		return nil, err
	}

	// 转换为简化版格式
	config := make(map[string]string)

	// CPU信息
	var cpuInfos []string
	for _, device := range hardwareInfo.CPUs {
		// 获取型号中的厂商信息
		var manufacturer string
		if strings.Contains(strings.ToLower(device.Model), "intel") {
			manufacturer = "Intel"
		} else if strings.Contains(strings.ToLower(device.Model), "amd") {
			manufacturer = "AMD"
		} else if strings.Contains(strings.ToLower(device.Model), "arm") {
			manufacturer = "ARM"
		} else {
			manufacturer = "Unknown"
		}

		// 从Model中提取型号部分（去除厂商信息）
		model := device.Model
		if strings.HasPrefix(strings.ToLower(model), strings.ToLower(manufacturer)) {
			model = strings.TrimSpace(model[len(manufacturer):])
		}

		// 获取CPU功耗
		tdp := c.estimateCPUTDP(device.Model)

		cpuInfos = append(cpuInfos, fmt.Sprintf("%s %s %dcore %.1fGHz %dW",
			manufacturer, model, device.Cores, device.Frequency, tdp))
	}
	if len(cpuInfos) > 0 {
		config["CPU"] = strings.Join(cpuInfos, "; ")
	}

	// 内存信息
	var totalMemoryGB int64 = 0
	for _, device := range hardwareInfo.Memories {
		totalMemoryGB += device.Size / (1024 * 1024 * 1024)
	}

	// 标准化内存容量到标准值（4,8,16,32,64等）
	standardizedMemoryGB := c.standardizeMemorySize(totalMemoryGB)

	// 检测虚拟环境
	isVirtual := c.isVirtualEnvironment()
	isWSL := c.isWSLEnvironment()

	if standardizedMemoryGB > 0 {
		memoryFormat := "%d GB"
		if isWSL {
			memoryFormat = "%d GB (WSL)"
		} else if isVirtual {
			memoryFormat = "%d GB (Virtual)"
		}
		config["RAM"] = fmt.Sprintf(memoryFormat, standardizedMemoryGB)
	}

	// GPU信息
	var gpuInfos []string
	for _, device := range hardwareInfo.GPUs {
		// 显存大小转换为GB，确保精度合适
		memoryGB := float64(device.MemoryTotal) / (1024.0 * 1024.0 * 1024.0)
		// 四舍五入到最接近的整数GB
		memoryGBRounded := math.Round(memoryGB)

		// 使用数字序号替代PciSlot，简化表示
		slotNumber := "0"
		if strings.Contains(device.Model, "GeForce") || strings.Contains(device.Model, "RTX") || strings.Contains(device.Model, "GTX") {
			// 获取数字编号
			slotNumber = "0" // 默认编号
		}

		// 标准化GPU型号名称
		model := c.standardizeGPUModel(device.Model)

		gpuInfos = append(gpuInfos, fmt.Sprintf("%s,%s %dGB %dW",
			slotNumber, model, int(memoryGBRounded), device.TDP))
	}
	if len(gpuInfos) > 0 {
		config["GPU"] = strings.Join(gpuInfos, "; ")
	}

	// 存储设备信息
	var storageInfos []string
	for _, device := range hardwareInfo.Storages {
		// 跳过容量为0的设备
		if device.Capacity == 0 {
			continue
		}

		capacityGB := float64(device.Capacity) / (1024 * 1024 * 1024)

		// 存储容量表示
		var capacityStr string
		if capacityGB >= 900 { // 接近1TB的值都四舍五入为1TB
			// 计算TB数，四舍五入到最接近的整数TB
			capacityTB := math.Round(capacityGB / 1024)
			// 确保至少是1TB
			if capacityTB < 1 {
				capacityTB = 1
			}
			capacityStr = fmt.Sprintf("%.0fTB", capacityTB)
		} else {
			// 不到900GB用GB表示，四舍五入到整数GB
			capacityStr = fmt.Sprintf("%.0fGB", math.Round(capacityGB))
		}

		// 根据规范格式化输出：<挂载路径>,<类型> <型号> <容量>
		deviceInfo := ""

		// 设备类型和型号
		if device.Model != "" {
			// 包含路径信息
			if device.Path != "" {
				deviceInfo = fmt.Sprintf("%s,%s %s %s",
					device.Path, device.Type, device.Model, capacityStr)
			} else {
				// 无路径信息
				deviceInfo = fmt.Sprintf("%s %s %s",
					device.Type, device.Model, capacityStr)
			}
		} else {
			// 无型号信息
			if device.Path != "" {
				deviceInfo = fmt.Sprintf("%s,%s %s",
					device.Path, device.Type, capacityStr)
			} else {
				deviceInfo = fmt.Sprintf("%s %s",
					device.Type, capacityStr)
			}
		}

		storageInfos = append(storageInfos, deviceInfo)
	}
	if len(storageInfos) > 0 {
		config["Storage"] = strings.Join(storageInfos, "; ")
	}

	return config, nil
}

// standardizeMemorySize 标准化内存大小到常见值
func (c *HardwareCollector) standardizeMemorySize(sizeGB int64) int64 {
	standardSizes := []int64{4, 8, 16, 32, 64, 128, 256, 512, 1024}

	if sizeGB <= 0 {
		return 0
	}

	// 找到最接近的标准内存大小
	closest := standardSizes[0]
	for _, size := range standardSizes {
		if size >= sizeGB && (size-sizeGB) < (sizeGB-closest) {
			closest = size
		} else if size <= sizeGB && (sizeGB-size) < (closest-sizeGB) {
			closest = size
		}
	}

	return closest
}

// isVirtualEnvironment 检测是否在虚拟环境中
func (c *HardwareCollector) isVirtualEnvironment() bool {
	// 方法1: 检查dmidecode输出
	dmidecodeCmd, err := exec.Command("dmidecode", "-s", "system-manufacturer").Output()
	if err == nil {
		vendor := strings.ToLower(strings.TrimSpace(string(dmidecodeCmd)))
		if strings.Contains(vendor, "vmware") ||
			strings.Contains(vendor, "qemu") ||
			strings.Contains(vendor, "virtualbox") ||
			strings.Contains(vendor, "xen") ||
			strings.Contains(vendor, "microsoft") {
			return true
		}
	}

	// 方法2: 检查/proc/cpuinfo中的标志
	cpuinfo, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		if strings.Contains(strings.ToLower(string(cpuinfo)), "hypervisor") {
			return true
		}
	}

	// 方法3: 检查虚拟化特定文件
	paths := []string{
		"/proc/xen",
		"/proc/self/status", // 检查其中的"VxID"
		"/sys/hypervisor/type",
		"/sys/devices/virtual",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return true
		}
	}

	return false
}

// isWSLEnvironment 检测是否在WSL环境中
func (c *HardwareCollector) isWSLEnvironment() bool {
	// 检查/proc/version是否包含Microsoft或WSL
	procVersion, err := os.ReadFile("/proc/version")
	if err == nil {
		content := strings.ToLower(string(procVersion))
		return strings.Contains(content, "microsoft") || strings.Contains(content, "wsl")
	}

	// 检查WSL特定环境变量
	wslEnv := os.Getenv("WSL_DISTRO_NAME")
	if wslEnv != "" {
		return true
	}

	return false
}

// collectCPUInfo 采集CPU信息
func (c *HardwareCollector) collectCPUInfo(hardwareInfo *models.HardwareInfo) error {
	// 使用lscpu命令获取CPU详细信息
	lscpuPath, err := findCommand("lscpu")
	if err != nil {
		return fmt.Errorf("未找到lscpu命令: %v", err)
	}

	out, err := exec.Command(lscpuPath).Output()
	if err != nil {
		return fmt.Errorf("执行lscpu命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")

	var model string
	var cores int32
	var threads int32
	var frequency float64
	var socket string
	var sockets int32 = 1 // 默认为1个插槽

	for _, line := range lines {
		if strings.Contains(line, "Model name:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				model = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "CPU(s):") && !strings.Contains(line, "NUMA") && !strings.Contains(line, "socket") {
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
				socket = strings.TrimSpace(parts[1])
				if val, err := strconv.ParseInt(socket, 10, 32); err == nil {
					sockets = int32(val)
				}
			}
		} else if strings.Contains(line, "CPU max MHz:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					frequency = val / 1000.0 // 转换为GHz
				}
			}
		} else if frequency == 0 && strings.Contains(line, "CPU MHz:") {
			// 如果没有找到max MHz，使用当前MHz作为备选
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				if val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
					frequency = val / 1000.0 // 转换为GHz
				}
			}
		}
	}

	// 尝试通过/proc/cpuinfo获取更准确的频率信息
	cpuinfoData, err := os.ReadFile("/proc/cpuinfo")
	if err == nil {
		cpuinfoLines := strings.Split(string(cpuinfoData), "\n")
		for _, line := range cpuinfoLines {
			if strings.Contains(line, "cpu MHz") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					if val, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64); err == nil {
						// 只有在lscpu没有获取到频率时才更新
						if frequency == 0 {
							frequency = val / 1000.0 // 转换为GHz
						}
					}
				}
			}
		}
	}

	// 如果还是无法获取频率，可以尝试通过cpufreq-info获取
	if frequency == 0 {
		cpufreqPath, err := findCommand("cpufreq-info")
		if err == nil {
			out, err := exec.Command(cpufreqPath).Output()
			if err == nil {
				cpufreqLines := strings.Split(string(out), "\n")
				for _, line := range cpufreqLines {
					if strings.Contains(line, "current CPU frequency") {
						re := regexp.MustCompile(`(\d+\.\d+) [GM]Hz`)
						matches := re.FindStringSubmatch(line)
						if len(matches) > 1 {
							if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
								frequency = val
							}
						}
					}
				}
			}
		}
	}

	// 确保频率有一个合理的默认值
	if frequency == 0 {
		// 检查CPU型号中是否包含频率信息
		re := regexp.MustCompile(`(\d+\.\d+)GHz`)
		matches := re.FindStringSubmatch(model)
		if len(matches) > 1 {
			if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
				frequency = val
			}
		} else {
			// 使用一个合理的默认值，可以根据CPU型号调整
			frequency = 2.0 // 默认2.0GHz
		}
	}

	// 为多插槽CPU系统添加多个CPU设备
	for i := int32(0); i < sockets; i++ {
		// 标准化CPU型号名称
		standardizedModel := c.standardizeCPUModel(model)

		// 添加到硬件信息
		hardwareInfo.CPUs = append(hardwareInfo.CPUs, models.CPUDevice{
			Model:        standardizedModel,
			Cores:        cores,
			Threads:      threads,
			Frequency:    frequency,
			Cache:        0,        // 默认值
			Architecture: "x86_64", // 默认架构
		})
	}

	return nil
}

// standardizeCPUModel 标准化CPU型号名称
func (c *HardwareCollector) standardizeCPUModel(model string) string {
	// 移除商标符号
	model = strings.ReplaceAll(model, "(R)", "")
	model = strings.ReplaceAll(model, "(TM)", "")
	model = strings.ReplaceAll(model, "®", "")
	model = strings.ReplaceAll(model, "™", "")

	// 移除世代前缀，如"13th Gen"
	model = regexp.MustCompile(`\d+th Gen `).ReplaceAllString(model, "")

	// 移除冗余厂商信息
	if strings.Contains(model, "Intel Core") {
		model = strings.ReplaceAll(model, "Intel Core", "Intel")
	}

	// 移除多余空格
	model = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(model), " ")

	return model
}

// estimateCPUTDP 估算CPU功耗
func (c *HardwareCollector) estimateCPUTDP(model string) int32 {
	// 根据CPU型号估算TDP
	model = strings.ToLower(model)

	// Intel Core i9
	if strings.Contains(model, "i9") {
		if strings.Contains(model, "13900k") {
			return 125
		} else if strings.Contains(model, "13900h") {
			return 45
		} else if strings.Contains(model, "12900") {
			return 125
		} else {
			return 95 // 默认i9
		}
	}

	// Intel Core i7
	if strings.Contains(model, "i7") {
		if strings.Contains(model, "h") {
			return 45 // 移动版
		} else if strings.Contains(model, "u") {
			return 15 // 低功耗版
		} else {
			return 65 // 默认i7
		}
	}

	// Intel Core i5
	if strings.Contains(model, "i5") {
		if strings.Contains(model, "h") {
			return 45 // 移动版
		} else if strings.Contains(model, "u") {
			return 15 // 低功耗版
		} else {
			return 65 // 默认i5
		}
	}

	// AMD Ryzen
	if strings.Contains(model, "ryzen") {
		if strings.Contains(model, "9") {
			return 105
		} else if strings.Contains(model, "7") {
			return 65
		} else {
			return 65 // 默认Ryzen
		}
	}

	// 默认值
	return 65
}

// collectMemoryInfo 采集内存信息
func (c *HardwareCollector) collectMemoryInfo(hardwareInfo *models.HardwareInfo) error {
	// 这里实现内存信息采集逻辑
	// 1. 读取/proc/meminfo获取总内存大小
	// 2. 使用dmidecode命令获取内存条详情

	// 示例：通过/proc/meminfo获取内存大小
	out, err := exec.Command("cat", "/proc/meminfo").Output()
	if err != nil {
		return fmt.Errorf("读取内存信息失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")

	var totalMemoryKB int64 = 0

	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				if val, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
					totalMemoryKB = val
				}
			}
			break
		}
	}

	// 转换为字节
	totalMemoryBytes := totalMemoryKB * 1024

	// 添加到硬件信息
	hardwareInfo.Memories = append(hardwareInfo.Memories, models.MemoryDevice{
		Size:      totalMemoryBytes,
		Type:      "Unknown", // 详细信息需要dmidecode命令（需要root权限）
		Frequency: 0,         // 默认值
	})

	return nil
}

// collectGPUInfo 采集GPU信息
func (c *HardwareCollector) collectGPUInfo(hardwareInfo *models.HardwareInfo) error {
	// 检查是否存在NVIDIA GPU
	if err := c.collectNvidiaGPUInfo(hardwareInfo); err != nil {
		fmt.Printf("采集NVIDIA GPU信息失败: %v\n", err)
	}

	// 如果没有找到NVIDIA GPU，尝试AMD GPU
	if len(hardwareInfo.GPUs) == 0 {
		if err := c.collectAMDGPUInfo(hardwareInfo); err != nil {
			fmt.Printf("采集AMD GPU信息失败: %v\n", err)
		}
	}

	// 如果仍未找到GPU，尝试Intel GPU
	if len(hardwareInfo.GPUs) == 0 {
		if err := c.collectIntelGPUInfo(hardwareInfo); err != nil {
			fmt.Printf("采集Intel GPU信息失败: %v\n", err)
		}
	}

	return nil
}

// collectNvidiaGPUInfo 采集NVIDIA GPU信息
func (c *HardwareCollector) collectNvidiaGPUInfo(hardwareInfo *models.HardwareInfo) error {
	// 使用我们的工具函数查找nvidia-smi
	nvidiaSmiPath, err := findCommand("nvidia-smi")
	if err != nil {
		return fmt.Errorf("未找到nvidia-smi命令: %v", err)
	}

	// 执行nvidia-smi获取详细信息
	// 增加更多查询字段，包括UUID和序列号
	cmd := fmt.Sprintf("%s --query-gpu=name,memory.total,pci.bus_id,gpu_uuid,serial,power.default_limit --format=csv,noheader", nvidiaSmiPath)
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return fmt.Errorf("执行nvidia-smi命令失败: %v\n命令: %s", err, cmd)
	}

	// 解析输出
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		fields := strings.Split(line, ", ")
		if len(fields) >= 2 {
			// 创建GPU设备，只使用重构后模型支持的字段
			gpu := models.GPUDevice{
				Model: fields[0],
			}

			// 获取GPU显存
			if len(fields) >= 2 {
				memoryStr := strings.TrimSpace(fields[1])
				// 处理单位(通常是MiB)
				memoryStr = strings.ReplaceAll(memoryStr, "MiB", "")
				memoryStr = strings.ReplaceAll(memoryStr, " ", "")
				if memory, err := strconv.ParseFloat(memoryStr, 64); err == nil {
					gpu.MemoryTotal = int64(memory * 1024 * 1024) // 转换为字节
				} else {
					fmt.Printf("解析GPU显存大小失败: %v, 原始值: %s\n", err, memoryStr)
				}
			}

			// 获取PCI ID（不再保存，因为模型中没有这个字段）
			var pciID string
			if len(fields) >= 3 {
				pciID = strings.TrimSpace(fields[2])
				if pciID == "" {
					fmt.Printf("警告: GPU的PCI ID为空\n")
				}
			} else {
				fmt.Printf("警告: 未能获取GPU的PCI ID\n")
			}

			// UUID和序列号（不再保存，因为模型中没有这个字段）
			if len(fields) >= 4 {
				uuid := strings.TrimSpace(fields[3])
				if uuid == "" {
					fmt.Printf("警告: GPU的UUID为空\n")
				} else if uuid == "N/A" {
					fmt.Printf("警告: GPU的UUID不可用\n")
				}
			}

			// 实际序列号（不再保存）
			if len(fields) >= 5 {
				serial := strings.TrimSpace(fields[4])
				if serial == "" {
					fmt.Printf("警告: GPU的序列号为空\n")
				}
			}

			// 添加功耗信息
			if len(fields) >= 6 {
				tdpStr := strings.TrimSpace(fields[5])
				if tdpStr == "" {
					fmt.Printf("警告: GPU的功耗信息为空\n")
				} else if tdpStr != "N/A" {
					// 移除单位(通常是W)
					tdpStr = strings.ReplaceAll(tdpStr, "W", "")
					tdpStr = strings.TrimSpace(tdpStr)
					tdp, err := strconv.ParseFloat(tdpStr, 64)
					if err == nil {
						gpu.TDP = int32(tdp)
					} else {
						fmt.Printf("解析GPU功耗失败: %v, 原始值: %s\n", err, tdpStr)
					}
				}
			}

			// 设置架构
			gpu.Architecture = c.determineNvidiaArchitecture(fields[0])
			if gpu.Architecture == "Unknown" {
				fmt.Printf("警告: 无法确定GPU %s 的架构\n", fields[0])
			}

			// 设置驱动版本
			gpu.DriverVersion = "Unknown"

			// 设置计算能力
			gpu.ComputeCapability = "Unknown"

			// 添加到硬件信息
			hardwareInfo.GPUs = append(hardwareInfo.GPUs, gpu)
		}
	}

	// 尝试从lspci获取更多信息
	if len(hardwareInfo.GPUs) > 0 {
		// 在重构后，暂时禁用增强GPU信息的功能，因为字段发生变化
		// if err := c.enhanceNvidiaGPUInfoWithLspci(hardwareInfo); err != nil {
		// 	fmt.Printf("通过lspci增强GPU信息失败: %v\n", err)
		// }
	} else {
		fmt.Printf("警告: nvidia-smi未返回任何GPU信息\n")

		// 尝试使用lspci检测是否存在NVIDIA GPU
		lspciPath, err := findCommand("lspci")
		if err == nil {
			out, err := exec.Command(lspciPath).Output()
			if err == nil && strings.Contains(string(out), "NVIDIA") {
				fmt.Printf("检测到NVIDIA GPU设备，但nvidia-smi未返回信息，可能是驱动问题\n")

				// 尝试只从lspci获取基本信息
				out, err := exec.Command(lspciPath, "-v").Output()
				if err == nil {
					lines := strings.Split(string(out), "\n")
					for _, line := range lines {
						if strings.Contains(line, "NVIDIA") && (strings.Contains(line, "VGA") || strings.Contains(line, "3D")) {
							parts := strings.SplitN(line, ":", 2)
							if len(parts) >= 2 {
								// 这里不再使用pciSlot
								desc := strings.TrimSpace(parts[1])

								// 提取型号信息
								modelMatch := regexp.MustCompile(`\[([^\]]+)\]`).FindStringSubmatch(desc)
								model := "NVIDIA GPU"
								if len(modelMatch) > 1 {
									model = modelMatch[1]
								}

								gpu := models.GPUDevice{
									Model:             model,
									Architecture:      "Unknown",
									DriverVersion:     "Unknown",
									ComputeCapability: "Unknown",
								}

								hardwareInfo.GPUs = append(hardwareInfo.GPUs, gpu)
								fmt.Printf("从lspci中检测到NVIDIA GPU: %s\n", model)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// determineNvidiaArchitecture 根据型号确定NVIDIA GPU架构
func (c *HardwareCollector) determineNvidiaArchitecture(model string) string {
	model = strings.ToLower(model)

	// Ampere架构
	if strings.Contains(model, "a100") || strings.Contains(model, "a40") ||
		strings.Contains(model, "a30") || strings.Contains(model, "a10") ||
		strings.Contains(model, "rtx 30") || strings.Contains(model, "rtx a") {
		return "Ampere"
	}

	// Ada Lovelace架构
	if strings.Contains(model, "rtx 40") || strings.Contains(model, "4090") ||
		strings.Contains(model, "4080") || strings.Contains(model, "4070") ||
		strings.Contains(model, "4060") {
		return "Ada Lovelace"
	}

	// Turing架构
	if strings.Contains(model, "rtx 20") || strings.Contains(model, "gtx 16") ||
		strings.Contains(model, "quadro rtx") {
		return "Turing"
	}

	// Hopper架构
	if strings.Contains(model, "h100") {
		return "Hopper"
	}

	return "Unknown" // 默认架构
}

// collectAMDGPUInfo 采集AMD GPU信息
func (c *HardwareCollector) collectAMDGPUInfo(hardwareInfo *models.HardwareInfo) error {
	// 实现AMD GPU信息采集
	// 暂时不实现，或后续添加
	return nil
}

// collectIntelGPUInfo 采集Intel GPU信息
func (c *HardwareCollector) collectIntelGPUInfo(hardwareInfo *models.HardwareInfo) error {
	// 使用lspci命令查找Intel显卡
	lspciPath, err := findCommand("lspci")
	if err != nil {
		return fmt.Errorf("未找到lspci命令: %v", err)
	}

	// 使用完整路径执行命令
	out, err := exec.Command(lspciPath, "-v").Output()
	if err != nil {
		return fmt.Errorf("执行lspci命令失败: %v", err)
	}

	// 分析lspci输出，查找Intel GPU
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "VGA") && strings.Contains(line, "Intel") {
			// 提取型号信息
			modelStart := strings.Index(line, "Intel")
			if modelStart >= 0 {
				model := line[modelStart:]
				if idx := strings.Index(model, "["); idx > 0 {
					model = model[:idx]
				}
				model = strings.TrimSpace(model)

				// 添加到硬件信息，使用新的模型结构
				hardwareInfo.GPUs = append(hardwareInfo.GPUs, models.GPUDevice{
					Model:             model,
					MemoryTotal:       2 * 1024 * 1024 * 1024, // 假设2GB显存
					Architecture:      "Intel",
					DriverVersion:     "Unknown",
					ComputeCapability: "Unknown",
					TDP:               15, // 默认TDP较低
				})
				break
			}
		}
	}

	return nil
}

// collectStorageInfo 采集存储设备信息
func (c *HardwareCollector) collectStorageInfo(hardwareInfo *models.HardwareInfo) error {
	// 检查是否在WSL环境中
	isWSL := c.isWSLEnvironment()

	if isWSL {
		// WSL环境下使用df命令获取更准确的信息
		return c.collectStorageInfoWSL(hardwareInfo)
	} else {
		// 非WSL环境使用lsblk命令
		return c.collectStorageInfoNative(hardwareInfo)
	}
}

// collectStorageInfoWSL 在WSL环境下采集存储设备信息
func (c *HardwareCollector) collectStorageInfoWSL(hardwareInfo *models.HardwareInfo) error {
	out, err := exec.Command("df", "-h").Output()
	if err != nil {
		return fmt.Errorf("执行df命令失败: %v", err)
	}

	lines := strings.Split(string(out), "\n")
	for i := 1; i < len(lines); i++ { // 跳过标题行
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 6 {
			continue
		}

		filesystem := fields[0]
		size := fields[1]
		mountpoint := fields[5]

		// 过滤掉特殊文件系统和临时挂载点
		if strings.HasPrefix(filesystem, "none") ||
			strings.HasPrefix(filesystem, "tmpfs") ||
			strings.HasPrefix(mountpoint, "/usr/lib/wsl") ||
			strings.HasPrefix(mountpoint, "/run") ||
			strings.HasPrefix(mountpoint, "/sys") ||
			mountpoint == "/init" ||
			mountpoint == "/dev" {
			continue
		}

		// 主要关注根目录和windows挂载点
		if mountpoint == "/" || strings.HasPrefix(mountpoint, "/mnt") {
			// 解析容量，例如将"1007G"转换为字节
			var capacity int64
			if strings.HasSuffix(size, "G") {
				sizeVal, err := strconv.ParseFloat(size[:len(size)-1], 64)
				if err == nil {
					capacity = int64(sizeVal * 1024 * 1024 * 1024)
				}
			} else if strings.HasSuffix(size, "T") {
				sizeVal, err := strconv.ParseFloat(size[:len(size)-1], 64)
				if err == nil {
					capacity = int64(sizeVal * 1024 * 1024 * 1024 * 1024)
				}
			} else if strings.HasSuffix(size, "M") {
				sizeVal, err := strconv.ParseFloat(size[:len(size)-1], 64)
				if err == nil {
					capacity = int64(sizeVal * 1024 * 1024)
				}
			}

			// 跳过容量过小或解析失败的设备
			if capacity < 1024*1024*1024 { // 小于1GB
				continue
			}

			deviceType := "SSD"
			model := "Virtual Disk"

			// 推断存储类型
			if strings.HasPrefix(mountpoint, "/mnt/") {
				// Windows挂载点，全部视为虚拟硬盘，默认类型由容量决定
				if capacity > 500*1024*1024*1024 { // 500GB
					deviceType = "HDD"
				} else {
					deviceType = "SSD"
				}
			}

			// 添加到硬件信息
			hardwareInfo.Storages = append(hardwareInfo.Storages, models.StorageDevice{
				Type:     deviceType,
				Model:    model,
				Capacity: capacity,
				Path:     mountpoint,
			})
		}
	}

	return nil
}

// collectStorageInfoNative 在本机环境下采集存储设备信息
func (c *HardwareCollector) collectStorageInfoNative(hardwareInfo *models.HardwareInfo) error {
	// 使用lsblk命令获取存储设备信息
	out, err := exec.Command("lsblk", "-b", "-d", "-o", "NAME,SIZE,TYPE,MODEL").Output()
	if err != nil {
		return fmt.Errorf("执行lsblk命令失败: %v", err)
	}

	// 解析输出
	lines := strings.Split(string(out), "\n")
	// 跳过标题行
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) >= 3 {
			name := fields[0]
			size, _ := strconv.ParseInt(fields[1], 10, 64)

			// 跳过容量为0或过小的设备
			if size < 1024*1024*1024 { // 小于1GB
				continue
			}

			// 获取模型信息
			model := ""
			if len(fields) >= 4 {
				model = strings.Join(fields[3:], " ")
			}

			// 如果模型信息为空，尝试从其他来源获取
			if model == "" {
				model = name
			}

			// 确定存储设备类型
			storageType := "HDD"
			if strings.Contains(name, "nvme") {
				storageType = "NVMe"
			} else if strings.Contains(model, "SSD") || strings.Contains(model, "Solid") {
				storageType = "SSD"
			}

			// 添加到硬件信息，使用新的模型结构
			hardwareInfo.Storages = append(hardwareInfo.Storages, models.StorageDevice{
				Type:     storageType,
				Model:    model,
				Capacity: size,
				Path:     "/dev/" + name,
			})
		}
	}

	return nil
}

// collectNetworkInfo 采集网络设备信息
func (c *HardwareCollector) collectNetworkInfo(hardwareInfo *models.HardwareInfo) error {
	// 使用ip命令获取网络接口信息
	_, err := exec.Command("ip", "link", "show").Output()
	if err != nil {
		return fmt.Errorf("执行ip命令失败: %v", err)
	}

	// 这里简化处理，实际上应该解析ip命令输出并获取更详细信息
	// 简单添加一个示例网卡信息，使用新的模型结构
	hardwareInfo.Networks = append(hardwareInfo.Networks, models.NetworkDevice{
		Name:       "eth0",
		MacAddress: "00:00:00:00:00:00",
		IpAddress:  "10.0.0.2",
		Speed:      1000,
	})

	return nil
}

// parseSize 解析大小字符串
func parseSize(sizeStr string) int64 {
	// 解析如 1G, 500M, 2T 等格式
	sizeStr = strings.TrimSpace(sizeStr)
	if len(sizeStr) == 0 {
		return 0
	}

	// 如果最后一个字符是数字，直接尝试解析
	lastChar := sizeStr[len(sizeStr)-1]
	if lastChar >= '0' && lastChar <= '9' {
		if val, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			return val
		}
		return 0
	}

	// 处理带单位的情况
	unit := string(lastChar)
	numPart := sizeStr[:len(sizeStr)-1]

	val, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0
	}

	switch strings.ToUpper(unit) {
	case "K":
		return int64(val * 1024)
	case "M":
		return int64(val * 1024 * 1024)
	case "G":
		return int64(val * 1024 * 1024 * 1024)
	case "T":
		return int64(val * 1024 * 1024 * 1024 * 1024)
	default:
		return int64(val)
	}
}

// standardizeGPUModel 标准化GPU型号名称
func (c *HardwareCollector) standardizeGPUModel(model string) string {
	// 移除冗余前缀
	model = strings.ReplaceAll(model, "GeForce ", "")
	model = strings.ReplaceAll(model, "Quadro ", "")
	model = strings.ReplaceAll(model, "Tesla ", "")

	// 处理笔记本GPU
	if strings.Contains(model, "Laptop") {
		// 将"XXX Laptop GPU"简化为"XXX Laptop"
		model = strings.ReplaceAll(model, "Laptop GPU", "Laptop")
	} else {
		// 移除独立的"GPU"字样
		model = strings.ReplaceAll(model, " GPU", "")
	}

	// 移除AMD前缀
	model = strings.ReplaceAll(model, "Radeon ", "")

	// 标准化Intel GPU命名
	if strings.Contains(model, "Intel") && !strings.Contains(model, "Arc") {
		model = strings.ReplaceAll(model, "Graphics", "")
		model = strings.ReplaceAll(model, "UHD", "UHD Graphics")
	}

	// 移除多余空格
	model = regexp.MustCompile(`\s+`).ReplaceAllString(strings.TrimSpace(model), " ")

	return model
}
