package configSaver

import (
	"log/slog"
	"os"
	"path/filepath"
)

const DSFileName = "deliveryServices.json"
const CNFileName = "cacheNodes.json"

func SaveDSToFile() {
	bgWg.Add(1)
	go func() {
		defer bgWg.Done()
		file, err := os.Create(filepath.Join(saveDir, DSFileName))
		if err != nil {
			slog.Error("Error creating file", "file", DSFileName, "error", err)
			return
		}
		defer file.Close()
		err = inMemConfig.SaveDsTo(file)
		if err != nil {
			slog.Error("Error output to file", "file", DSFileName, "error", err)
			return
		}
		slog.Info("Configuration saved", "file", DSFileName)
	}()
}

func LoadDSFromFile() error {
	file, err := os.Open(filepath.Join(saveDir, DSFileName))
	if err != nil {
		return err
	}
	defer file.Close()
	err = inMemConfig.ReadDsFrom(file)
	if err != nil {
		return err
	}
	return nil
}

func SaveCNToFile() {
	bgWg.Add(1)
	go func() {
		defer bgWg.Done()
		file, err := os.Create(filepath.Join(saveDir, CNFileName))
		if err != nil {
			slog.Error("Error creating file", "file", CNFileName, "error", err)
			return
		}
		defer file.Close()
		err = inMemConfig.SaveCnTo(file)
		if err != nil {
			slog.Error("Error output to file", "file", CNFileName, "error", err)
			return
		}
		slog.Info("Configuration saved", "file", CNFileName)
	}()
}

func LoadCNFromFile() error {
	file, err := os.Open(filepath.Join(saveDir, CNFileName))
	if err != nil {
		return err
	}
	defer file.Close()
	err = inMemConfig.ReadCnFrom(file)
	if err != nil {
		return err
	}
	return nil
}
