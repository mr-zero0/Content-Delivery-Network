package storage

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"path"
	"path/filepath"
)

// By default, set high disk usage simulator as below disk threshold
var highDiskUsageSimulator int = 0

func simulateHighDiskUsage() {
	highDiskUsageSimulator = 1
}

func GetDefaultCDNDirectory() string {
	//Get current working directory to store all CDN contents
	return path.Join(".", "cdn")
}

func getDiskUsage(baseDir string) int {
	var diskUsage int

	//Create SimulateDiskFull under base cdn directory to simulate disk full scenario
	diskFullFile := filepath.Join(baseDir, "SimulateDiskFull")
	_, err := os.Stat(diskFullFile)
	if err == nil {
		os.RemoveAll(diskFullFile)
		highDiskUsageSimulator = 1
	}
	
	if highDiskUsageSimulator == 0 {
		cmdStr := fmt.Sprintf("df %s | tail -1 | tr -s ' ' | cut -d' ' -f5 | sed 's/.$//' | tr -d '\\n'", baseDir)
		diskUsageCmd := exec.Command("bash", "-c", cmdStr)
		output, err := diskUsageCmd.CombinedOutput()
		if err != nil {
			slog.Info("Storage:CacheEvictor:Failed to fetch linux disk usage", "error", err)
		}
		diskUsage, err = strconv.Atoi(string(output))
		if err != nil {
			slog.Info("Storage:CacheEvictor:Failed to convert linux disk usage to int", "error", err)
		}
	} else {
		diskUsage = 80
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
