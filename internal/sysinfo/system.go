package sysinfo

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/open-beagle/beagle-wind-game/internal/models"
	"github.com/open-beagle/beagle-wind-game/internal/utils"
)

// SystemCollector 系统信息采集器
type SystemCollector struct {
	// 可能的配置选项
	options map[string]string
	// 日志框架
	logger utils.Logger
}

// NewSystemCollector 创建新的系统信息采集器
func NewSystemCollector(options map[string]string) *SystemCollector {
	if options == nil {
		options = make(map[string]string)
	}
	// 创建日志器
	logger := utils.New("SystemCollector")
	return &SystemCollector{
		options: options,
		logger:  logger,
	}
}

// GetSystemInfo 采集系统信息
func (c *SystemCollector) GetSystemInfo() (models.SystemInfo, error) {
	var systemInfo models.SystemInfo

	// 采集操作系统信息
	if err := c.collectOSInfo(&systemInfo); err != nil {
		c.logger.Warn("采集操作系统信息失败: %v", err)
	}

	// 采集GPU驱动信息
	if err := c.collectGPUDriverInfo(&systemInfo); err != nil {
		c.logger.Warn("采集GPU驱动信息失败: %v", err)
	}

	// 采集CUDA版本信息
	if err := c.collectCUDAInfo(&systemInfo); err != nil {
		c.logger.Warn("采集CUDA版本信息失败: %v", err)
	}

	// 采集容器运行时信息
	if err := c.collectContainerRuntimeInfo(&systemInfo); err != nil {
		c.logger.Warn("采集容器运行时信息失败: %v", err)
	}

	// 确保关键字段有默认值，不会为空
	c.ensureDefaultValues(&systemInfo)

	return systemInfo, nil
}

// ensureDefaultValues 确保关键字段有值
func (c *SystemCollector) ensureDefaultValues(systemInfo *models.SystemInfo) {
	// 如果OS信息为空，尝试使用uname命令获取基本信息
	if systemInfo.OSDistribution == "" {
		out, err := exec.Command("uname", "-s").Output()
		if err == nil {
			systemInfo.OSDistribution = strings.TrimSpace(string(out))
		} else {
			c.logger.Warn("警告: 无法获取操作系统分发版信息: %v", err)
		}
	}

	if systemInfo.OSVersion == "" {
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			systemInfo.OSVersion = strings.TrimSpace(string(out))
		} else {
			c.logger.Warn("警告: 无法获取操作系统版本信息: %v", err)
		}
	}

	if systemInfo.OSArchitecture == "" {
		out, err := exec.Command("uname", "-m").Output()
		if err == nil {
			systemInfo.OSArchitecture = strings.TrimSpace(string(out))
		} else {
			c.logger.Warn("警告: 无法获取操作系统架构信息: %v", err)
		}
	}

	if systemInfo.KernelVersion == "" {
		out, err := exec.Command("uname", "-r").Output()
		if err == nil {
			systemInfo.KernelVersion = strings.TrimSpace(string(out))
		} else {
			c.logger.Warn("警告: 无法获取内核版本信息: %v", err)
		}
	}

	// 如果GPU驱动信息为空，但在硬件中有GPU，记录日志
	if systemInfo.GPUDriverVersion == "" {
		// 使用lspci命令检测GPU存在
		out, err := exec.Command("lspci").Output()
		if err == nil && strings.Contains(string(out), "VGA") {
			if strings.Contains(string(out), "NVIDIA") {
				c.logger.Warn("警告: 检测到NVIDIA GPU，但无法获取驱动版本")
			} else if strings.Contains(string(out), "AMD") {
				c.logger.Warn("警告: 检测到AMD GPU，但无法获取驱动版本")
			} else if strings.Contains(string(out), "Intel") {
				c.logger.Warn("警告: 检测到Intel GPU，但无法获取驱动版本")
			} else {
				c.logger.Warn("警告: 检测到未知类型的GPU，但无法获取驱动版本")
			}
		}
	}

	// 记录CUDA版本缺失的日志
	if systemInfo.GPUComputeAPIVersion == "" && strings.Contains(systemInfo.GPUDriverVersion, "NVIDIA") {
		c.logger.Warn("警告: 检测到NVIDIA驱动，但无法获取CUDA版本")
	}
}

// GetSimplifiedSystemInfo 获取简化版系统信息（键值对格式）
func (c *SystemCollector) GetSimplifiedSystemInfo() (map[string]string, error) {
	systemInfo, err := c.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	// 转换为简化版格式
	config := make(map[string]string)

	// 操作系统信息
	if systemInfo.OSDistribution != "" {
		config["os_type"] = systemInfo.OSDistribution
	}
	if systemInfo.OSVersion != "" {
		config["os_version"] = systemInfo.OSVersion
	}
	if systemInfo.OSArchitecture != "" {
		config["os_architecture"] = systemInfo.OSArchitecture
	}

	// GPU驱动信息
	if systemInfo.GPUDriverVersion != "" {
		// 确保GPU驱动格式符合标准
		if strings.Contains(systemInfo.GPUDriverVersion, ":") {
			// 处理可能的旧格式 "NVIDIA: 572.83"
			parts := strings.Split(systemInfo.GPUDriverVersion, ":")
			if len(parts) == 2 {
				vendor := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				config["gpu_driver"] = fmt.Sprintf("%s Driver %s", vendor, version)
			} else {
				config["gpu_driver"] = systemInfo.GPUDriverVersion
			}
		} else if !strings.Contains(systemInfo.GPUDriverVersion, "Driver") {
			// 处理其他格式，确保包含"Driver"字样
			if strings.Contains(systemInfo.GPUDriverVersion, "NVIDIA") {
				config["gpu_driver"] = strings.ReplaceAll(systemInfo.GPUDriverVersion, "NVIDIA ", "NVIDIA Driver ")
			} else if strings.Contains(systemInfo.GPUDriverVersion, "AMD") {
				config["gpu_driver"] = strings.ReplaceAll(systemInfo.GPUDriverVersion, "AMD ", "AMD Driver ")
			} else if strings.Contains(systemInfo.GPUDriverVersion, "Intel") {
				config["gpu_driver"] = strings.ReplaceAll(systemInfo.GPUDriverVersion, "Intel ", "Intel Driver ")
			} else {
				// 如果无法确定厂商，使用通用格式
				config["gpu_driver"] = fmt.Sprintf("Driver %s", systemInfo.GPUDriverVersion)
			}
		} else {
			// 已经符合标准格式
			config["gpu_driver"] = systemInfo.GPUDriverVersion
		}
	}

	// CUDA版本信息
	if systemInfo.GPUComputeAPIVersion != "" {
		config["cuda_version"] = systemInfo.GPUComputeAPIVersion
	}

	return config, nil
}

// collectOSInfo 采集操作系统信息
func (c *SystemCollector) collectOSInfo(systemInfo *models.SystemInfo) error {
	// 尝试多种方法获取OS信息，确保能在各种环境下工作

	// 1. 尝试使用lsb_release命令获取OS信息
	lsbReleaseSuccess := false
	out, err := exec.Command("lsb_release", "-a").Output()
	if err == nil {
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, "Distributor ID:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					systemInfo.OSDistribution = strings.TrimSpace(parts[1])
					lsbReleaseSuccess = true
				}
			} else if strings.Contains(line, "Release:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					systemInfo.OSVersion = strings.TrimSpace(parts[1])
				}
			}
		}
	}

	// 2. 如果lsb_release命令失败，尝试读取/etc/os-release文件
	if !lsbReleaseSuccess || systemInfo.OSDistribution == "" || systemInfo.OSVersion == "" {
		if osReleaseData, err := os.ReadFile("/etc/os-release"); err == nil {
			lines := strings.Split(string(osReleaseData), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "NAME=") {
					value := strings.Trim(strings.TrimPrefix(line, "NAME="), "\"")
					systemInfo.OSDistribution = value
				} else if strings.HasPrefix(line, "VERSION_ID=") {
					value := strings.Trim(strings.TrimPrefix(line, "VERSION_ID="), "\"")
					systemInfo.OSVersion = value
				}
			}
		}
	}

	// 3. 在Windows/WSL环境下，尝试特定方法
	if c.isWSLEnvironment() {
		// 检测WSL版本和发行版
		systemInfo.OSDistribution += " (WSL)"
		// 如果是WSL环境，获取Windows版本
		if winVerData, err := os.ReadFile("/proc/version"); err == nil {
			winVerStr := string(winVerData)
			if match := regexp.MustCompile(`Microsoft Windows (\d+\.\d+\.\d+)`).FindStringSubmatch(winVerStr); len(match) > 1 {
				systemInfo.OSVersion += " on Windows " + match[1]
			}
		}
	}

	// 获取系统架构
	out, err = exec.Command("uname", "-m").Output()
	if err == nil {
		systemInfo.OSArchitecture = strings.TrimSpace(string(out))
	}

	// 获取内核版本
	out, err = exec.Command("uname", "-r").Output()
	if err == nil {
		systemInfo.KernelVersion = strings.TrimSpace(string(out))
	}

	return nil
}

// isWSLEnvironment 检测是否在WSL环境中
func (c *SystemCollector) isWSLEnvironment() bool {
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

// collectGPUDriverInfo 采集GPU驱动信息
func (c *SystemCollector) collectGPUDriverInfo(systemInfo *models.SystemInfo) error {
	// 首先尝试检测GPU类型

	// 尝试NVIDIA GPU驱动
	if err := c.collectNvidiaDriverInfo(systemInfo); err != nil {
		c.logger.Warn("采集NVIDIA驱动信息失败: %v", err)
	}

	// 如果没有找到NVIDIA驱动，尝试AMD GPU驱动
	if systemInfo.GPUDriverVersion == "" {
		if err := c.collectAMDDriverInfo(systemInfo); err != nil {
			c.logger.Warn("采集AMD驱动信息失败: %v", err)
		}
	}

	// 如果仍未找到驱动，尝试Intel GPU驱动
	if systemInfo.GPUDriverVersion == "" {
		if err := c.collectIntelDriverInfo(systemInfo); err != nil {
			c.logger.Warn("采集Intel驱动信息失败: %v", err)
		}
	}

	// 在WSL环境下，可以尝试从Windows获取GPU驱动信息
	if systemInfo.GPUDriverVersion == "" && c.isWSLEnvironment() {
		if err := c.collectWSLGPUDriverInfo(systemInfo); err != nil {
			c.logger.Warn("采集WSL GPU驱动信息失败: %v", err)
		}
	}

	return nil
}

// collectWSLGPUDriverInfo 采集WSL环境下的GPU驱动信息
func (c *SystemCollector) collectWSLGPUDriverInfo(systemInfo *models.SystemInfo) error {
	// 尝试使用powershell命令获取Windows上的驱动信息
	// 注意：这需要WSL可以执行Windows命令
	cmd := `powershell.exe -Command "Get-CimInstance Win32_VideoController | Select-Object Name, DriverVersion | ConvertTo-Json"`
	out, err := exec.Command("bash", "-c", cmd).Output()
	if err == nil && len(out) > 0 {
		// 解析JSON输出
		if strings.Contains(string(out), "NVIDIA") {
			re := regexp.MustCompile(`"DriverVersion":\s*"([^"]+)"`)
			matches := re.FindStringSubmatch(string(out))
			if len(matches) > 1 {
				systemInfo.GPUDriverVersion = "NVIDIA Driver " + matches[1]
				return nil
			}
		} else if strings.Contains(string(out), "AMD") {
			re := regexp.MustCompile(`"DriverVersion":\s*"([^"]+)"`)
			matches := re.FindStringSubmatch(string(out))
			if len(matches) > 1 {
				systemInfo.GPUDriverVersion = "AMD Driver " + matches[1]
				return nil
			}
		} else if strings.Contains(string(out), "Intel") {
			re := regexp.MustCompile(`"DriverVersion":\s*"([^"]+)"`)
			matches := re.FindStringSubmatch(string(out))
			if len(matches) > 1 {
				systemInfo.GPUDriverVersion = "Intel Driver " + matches[1]
				return nil
			}
		}
	}
	return fmt.Errorf("无法从WSL获取GPU驱动信息")
}

// collectNvidiaDriverInfo 采集NVIDIA驱动信息
func (c *SystemCollector) collectNvidiaDriverInfo(systemInfo *models.SystemInfo) error {
	// 使用我们的工具函数查找nvidia-smi
	nvidiaPath, err := findCommand("nvidia-smi")
	if err != nil {
		return fmt.Errorf("未找到nvidia-smi命令: %v", err)
	}

	// 使用完整路径执行命令
	out, err := exec.Command(nvidiaPath, "--query-gpu=driver_version", "--format=csv,noheader").Output()
	if err != nil {
		return fmt.Errorf("执行nvidia-smi命令失败: %v", err)
	}

	// 提取驱动版本
	driverVersion := strings.TrimSpace(string(out))
	if strings.Contains(driverVersion, "\n") {
		// 如果有多个设备，只取第一个
		driverVersion = strings.Split(driverVersion, "\n")[0]
	}

	// 更新系统信息
	if driverVersion != "" {
		systemInfo.GPUDriverVersion = "NVIDIA Driver " + driverVersion
	}

	return nil
}

// collectAMDDriverInfo 采集AMD驱动信息
func (c *SystemCollector) collectAMDDriverInfo(systemInfo *models.SystemInfo) error {
	// 使用我们的工具函数查找rocm-smi
	rocmPath, err := findCommand("rocm-smi")
	if err != nil {
		return fmt.Errorf("未找到rocm-smi命令: %v", err)
	}

	// 使用完整路径执行命令
	out, err := exec.Command(rocmPath, "--showdriverversion").Output()
	if err != nil {
		return fmt.Errorf("执行rocm-smi命令失败: %v", err)
	}

	// 提取驱动版本
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Driver Version") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				driverVersion := strings.TrimSpace(parts[1])
				systemInfo.GPUDriverVersion = "AMD Driver " + driverVersion
				break
			}
		}
	}

	return nil
}

// collectIntelDriverInfo 采集Intel驱动信息
func (c *SystemCollector) collectIntelDriverInfo(systemInfo *models.SystemInfo) error {
	// 使用我们的工具函数查找glxinfo
	glxinfoPath, err := findCommand("glxinfo")
	if err != nil {
		return fmt.Errorf("未找到glxinfo命令: %v", err)
	}

	// 使用完整路径执行命令
	out, err := exec.Command(glxinfoPath, "-B").Output()
	if err != nil {
		return fmt.Errorf("执行glxinfo命令失败: %v", err)
	}

	// 提取驱动版本
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "OpenGL version string") && strings.Contains(line, "Intel") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				driverVersion := strings.TrimSpace(parts[1])
				systemInfo.GPUDriverVersion = "Intel Driver " + driverVersion
				break
			}
		}
	}

	return nil
}

// collectCUDAInfo 采集CUDA版本信息
func (c *SystemCollector) collectCUDAInfo(systemInfo *models.SystemInfo) error {
	// 尝试从nvcc命令获取CUDA版本
	nvccPath, err := findCommand("nvcc")
	if err == nil {
		out, err := exec.Command(nvccPath, "--version").Output()
		if err == nil {
			// 从nvcc输出提取CUDA版本
			re := regexp.MustCompile(`release (\d+\.\d+)`)
			matches := re.FindStringSubmatch(string(out))
			if len(matches) > 1 {
				systemInfo.GPUComputeAPIVersion = matches[1]
				return nil
			}
		}
	}

	// 如果nvcc命令失败，从nvidia-smi获取CUDA版本
	nvidiaPath, err := findCommand("nvidia-smi")
	if err == nil {
		out, err := exec.Command(nvidiaPath, "--query-gpu=driver_version,cuda_version", "--format=csv,noheader").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 0 {
				fields := strings.Split(lines[0], ", ")
				if len(fields) >= 2 {
					cudaVersion := strings.TrimSpace(fields[1])
					// CUDA版本通常是类似"11.4"这样的格式
					systemInfo.GPUComputeAPIVersion = cudaVersion
					return nil
				}
			}
		}
	}

	// 尝试其他方式获取CUDA版本
	cudaPath := "/usr/local/cuda"
	if _, err := os.Stat(cudaPath); err == nil {
		files, err := os.ReadDir(cudaPath)
		if err == nil {
			var versionDirs []string
			for _, file := range files {
				if file.IsDir() && strings.HasPrefix(file.Name(), "version") {
					versionDirs = append(versionDirs, file.Name())
				}
			}
			if len(versionDirs) > 0 {
				// 取第一个版本目录
				versionDir := versionDirs[0]
				versionMatch := regexp.MustCompile(`version-(\d+\.\d+)`).FindStringSubmatch(versionDir)
				if len(versionMatch) > 1 {
					systemInfo.GPUComputeAPIVersion = versionMatch[1]
					return nil
				}
			}
		}

		// 尝试从cuda中的版本文件获取
		versionFile := fmt.Sprintf("%s/version.txt", cudaPath)
		if versionData, err := os.ReadFile(versionFile); err == nil {
			version := strings.TrimSpace(string(versionData))
			// 提取版本号
			re := regexp.MustCompile(`(\d+\.\d+)`)
			matches := re.FindStringSubmatch(version)
			if len(matches) > 1 {
				systemInfo.GPUComputeAPIVersion = matches[1]
				return nil
			}
		}
	}

	return nil
}

// collectContainerRuntimeInfo 采集容器运行时信息
func (c *SystemCollector) collectContainerRuntimeInfo(systemInfo *models.SystemInfo) error {
	// 尝试获取Docker版本
	dockerPath, err := findCommand("docker")
	if err == nil {
		// 获取Docker版本
		out, err := exec.Command(dockerPath, "--version").Output()
		if err == nil {
			// 从输出中提取Docker版本
			re := regexp.MustCompile(`Docker version (\d+\.\d+\.\d+)`)
			matches := re.FindStringSubmatch(string(out))
			if len(matches) > 1 {
				systemInfo.DockerVersion = matches[1]
			}
		}

		// 使用docker version命令获取containerd和runc版本
		versionOut, err := exec.Command(dockerPath, "version").Output()
		if err == nil {
			versionStr := string(versionOut)

			// 获取containerd版本
			containerdRe := regexp.MustCompile(`containerd:\s+Version:\s+(\S+)`)
			containerdMatches := containerdRe.FindStringSubmatch(versionStr)
			if len(containerdMatches) > 1 {
				systemInfo.ContainerdVersion = containerdMatches[1]
			}

			// 获取containerd提交ID（如果版本为空）
			if systemInfo.ContainerdVersion == "" {
				containerdCommitRe := regexp.MustCompile(`containerd:\s+GitCommit:\s+(\S+)`)
				containerdCommitMatches := containerdCommitRe.FindStringSubmatch(versionStr)
				if len(containerdCommitMatches) > 1 {
					systemInfo.ContainerdVersion = containerdCommitMatches[1]
				}
			}

			// 获取runc版本
			runcRe := regexp.MustCompile(`runc:\s+Version:\s+(\S+)`)
			runcMatches := runcRe.FindStringSubmatch(versionStr)
			if len(runcMatches) > 1 {
				systemInfo.RuncVersion = runcMatches[1]
			}

			// 获取runc提交ID（如果版本为空）
			if systemInfo.RuncVersion == "" {
				runcCommitRe := regexp.MustCompile(`runc:\s+GitCommit:\s+(\S+)`)
				runcCommitMatches := runcCommitRe.FindStringSubmatch(versionStr)
				if len(runcCommitMatches) > 1 {
					systemInfo.RuncVersion = runcCommitMatches[1]
				}
			}
		}

		// 如果docker version没有获取到containerd和runc版本，回退到docker info命令
		if systemInfo.ContainerdVersion == "" || systemInfo.RuncVersion == "" {
			infoOut, err := exec.Command(dockerPath, "info").Output()
			if err == nil {
				infoStr := string(infoOut)

				// 从docker info中提取containerd版本
				if systemInfo.ContainerdVersion == "" {
					containerdInfoRe := regexp.MustCompile(`containerd version: (\S+)`)
					containerdInfoMatches := containerdInfoRe.FindStringSubmatch(infoStr)
					if len(containerdInfoMatches) > 1 {
						systemInfo.ContainerdVersion = containerdInfoMatches[1]
					}
				}

				// 从docker info中提取runc版本
				if systemInfo.RuncVersion == "" {
					runcInfoRe := regexp.MustCompile(`runc version: (\S+)`)
					runcInfoMatches := runcInfoRe.FindStringSubmatch(infoStr)
					if len(runcInfoMatches) > 1 {
						systemInfo.RuncVersion = runcInfoMatches[1]
					}
				}
			}
		}
	}

	// 如果docker命令获取失败，仍然尝试单独获取containerd和runc版本
	if systemInfo.ContainerdVersion == "" {
		containerdPath, err := findCommand("containerd")
		if err == nil {
			out, err := exec.Command(containerdPath, "--version").Output()
			if err == nil {
				// 从输出中提取containerd版本
				re := regexp.MustCompile(`containerd version (\S+)`)
				matches := re.FindStringSubmatch(string(out))
				if len(matches) > 1 {
					systemInfo.ContainerdVersion = matches[1]
				}
			}
		}
	}

	if systemInfo.RuncVersion == "" {
		runcPath, err := findCommand("runc")
		if err == nil {
			out, err := exec.Command(runcPath, "--version").Output()
			if err == nil {
				// 从输出中提取runc版本
				re := regexp.MustCompile(`runc version (\S+)`)
				matches := re.FindStringSubmatch(string(out))
				if len(matches) > 1 {
					systemInfo.RuncVersion = matches[1]
				}
			}
		}
	}

	return nil
}
