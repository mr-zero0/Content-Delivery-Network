package configSaver

import (
	"os"
	"testing"
	"github.com/hcl/cdn/configServer/inMemoryConfig"
	"github.com/hcl/cdn/common/config"
	"context"
	"log"
	"sync"
	"path/filepath"
	"time"
)
var testDir = "C:\\Users\\prathmesh.meena\\Documents\\CDN_Latest"

func setup() {
	// Ensure the test directory exists
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		os.Mkdir(testDir, os.ModePerm)
	}
	
	inMemConfig = &inMemoryConfig.InMemoryConfig{}
	inMemConfig.AddDs(&config.DeliveryService{
		Name:      "service1",
		ClientURL: "http://client1.com",
		OriginURL: "http://origin1.com",
	})
	inMemConfig.AddCn(&config.CacheNode{
		IP:         "192.168.1.1",
		Port:       8080,
		Type:       "Edge",
		ParentIP:   "192.168.1.2",
		ParentPort: 8081,
	})
	log.Printf("Setup inMemConfig: %v", inMemConfig)
}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func TestSaveDSToFile(t *testing.T) {
	Init(context.Background(), &sync.WaitGroup{}, inMemConfig, testDir)
	log.Println("Saving delivery services to file")
	SaveDSToFile()

	// Wait for the asynchronous save to complete
	time.Sleep(300 * time.Millisecond)

	// Check if the file exists
	filePath := filepath.Join(testDir, DSFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", DSFileName)
	} else {
		log.Printf("File %s created successfully", DSFileName)
	}

	// Log the contents of the file
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("error reading file %s: %v", DSFileName, err)
	} else {
		log.Printf("Contents of %s: %s", DSFileName, string(fileContents))
	}
}

func TestLoadDSFromFile(t *testing.T) {
	Init(context.Background(), &sync.WaitGroup{}, inMemConfig, testDir)
	log.Println("Saving delivery services to file")
	SaveDSToFile()

	// Wait for the asynchronous save to complete
	time.Sleep(100 * time.Millisecond)

	// Clear the in-memory config and load from file
	inMemConfig = &inMemoryConfig.InMemoryConfig{}
	log.Println("Loading delivery services from file")
	err := LoadDSFromFile()
	if err != nil {
		t.Fatalf("failed to load delivery services from file: %v", err)
	}

	// Verify the loaded data
	version, serviceList := inMemConfig.GetDsNames()
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
	if len(serviceList) != 1 || serviceList[0] != "service1" {
		t.Errorf("expected service1, got %v", serviceList)
	} else {
		log.Printf("Loaded delivery services: %v", serviceList)
	}
}

func TestSaveCNToFile(t *testing.T) {
	Init(context.Background(), &sync.WaitGroup{}, inMemConfig, testDir)
	log.Println("Saving cache nodes to file")
	SaveCNToFile()

	// Wait for the asynchronous save to complete
	time.Sleep(500 * time.Millisecond)

	// Check if the file exists
	filePath := filepath.Join(testDir, CNFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Errorf("expected file %s to be created", CNFileName)
	} else {
		log.Printf("File %s created successfully", CNFileName)
	}

	// Log the contents of the file
	fileContents, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("error reading file %s: %v", CNFileName, err)
	} else {
		log.Printf("Contents of %s: %s", CNFileName, string(fileContents))
	}
}

func TestLoadCNFromFile(t *testing.T) {
	Init(context.Background(), &sync.WaitGroup{}, inMemConfig, testDir)
	log.Println("Saving cache nodes to file")
	SaveCNToFile()

	// Wait for the asynchronous save to complete
	time.Sleep(1000 * time.Millisecond)

	// Clear the in-memory config and load from file

	inMemConfig = &inMemoryConfig.InMemoryConfig{}
	log.Println("Loading cache nodes from file")
	err := LoadCNFromFile()
	if err != nil {
		t.Fatalf("failed to load cache nodes from file: %v", err)
	}

	// Verify the loaded data
	version, nodeList := inMemConfig.GetCnIps()
	if version != 1 {
		t.Errorf("expected version 1, got %d", version)
	}
	if len(nodeList) != 1 || nodeList[0] != "192.168.1.1" {
		t.Errorf("expected 192.168.1.1, got %v", nodeList)
	} else {
		log.Printf("Loaded cache nodes: %v", nodeList)
	}
}

// Edge case: Test loading from a non-existent file
// func TestLoadDSFromNonExistentFile(t *testing.T) {
// 	Init(context.Background(), &sync.WaitGroup{}, inMemConfig, testDir)

// 	// Remove the file if it exists
// 	os.Remove(filepath.Join(testDir, DSFileName))

// 	log.Println("Loading delivery services from non-existent file")
// 	err := LoadDSFromFile()
// 	if err == nil {
// 		t.Errorf("expected error when loading from non-existent file, got nil")
// 	} else {
// 		log.Printf("Expected error when loading from non-existent file: %v", err)
// 	}
// }