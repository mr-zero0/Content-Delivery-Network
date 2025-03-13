package storage

import (
	"strings"
	"log/slog"
	"os/exec"
	"strconv"
	"os"
	"path/filepath"
)

// By default, set high disk usage simulator as below disk threshold
var highDiskUsageSimulator int = 0

func simulateHighDiskUsage() {
	highDiskUsageSimulator = 1
}

func GetDefaultCDNDirectory() string {
	//Get current working directory to store all CDN contents
	defaultCdnDir, err := os.Getwd()
	if err != nil {
		defaultCdnDir = "C:"
	}
	return defaultCdnDir + "\\cdn"
}

func getDiskUsage(baseDir string) int {
	var diskUsage int = 0

	//Create SimulateDiskFull under base cdn directory to simulate disk full scenario
	diskFullFile := filepath.Join(baseDir, "SimulateDiskFull")
	_, err := os.Stat(diskFullFile)
	if err == nil {
		os.RemoveAll(diskFullFile)
		highDiskUsageSimulator = 1
	}

	if highDiskUsageSimulator == 0 {
		absolutePath, _ := filepath.Abs(baseDir)
		drive := strings.Split(absolutePath, ":")[0]
		cmd := exec.Command("powershell", "Get-PSDrive ", drive)
		output, _ := cmd.Output()
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, drive+":") {
				diskstats := strings.Fields(line)
				used, _ := strconv.ParseFloat(diskstats[1], 64)
				free, _ := strconv.ParseFloat(diskstats[2], 64)
				diskUsage = int((used/(used+free))*100)
			}
		}
	} else {		
		switch highDiskUsageSimulator {
		case 1:
			diskUsage = 85
			highDiskUsageSimulator++
		default:
			diskUsage = 80
			highDiskUsageSimulator = 0
		}
	}

	slog.Debug("Storage:CacheEvictor", "DiskUsage", diskUsage)
	return diskUsage
}
