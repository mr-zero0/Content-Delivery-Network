package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/hcl/cdn/common/config"
	"github.com/hcl/cdn/common/helper"
)

// configuration
type RunConfig struct {
	Valid       bool                     `json:"-"`           // is config Valid?
	Filename    string                   `json:"-"`           // name of the file to persist config
	Node        *config.CacheNode        `json:"node"`        // cache node details
	ServiceList *config.DeliveryServices `json:"serviceList"` // list of deliver services
	mu          sync.RWMutex
}

func NewConfigFromFile(filename string) (*RunConfig, error) {
	var err error
	// read all config value from file
	// Check if the file already exists
	if _, err = os.Stat(filename); os.IsNotExist(err) {
		// Create the file if it does not exist
		file, err := os.Create(filename)
		if err != nil {
			return nil, fmt.Errorf("failed to create file: %w", err)
		}
		defer file.Close()
	}

	ret := &RunConfig{
		Filename: filename,
		Valid:    false,
	}

	slog.Info("Going to populate node and dsConfig instances in config Struct")
	slog.Info("Locking the runConfig RWMutex~")
	ret.mu.Lock()
	defer ret.mu.Unlock()

	err = ret.readConfigFile()
	if err != nil {
		return ret, nil
	}
	slog.Info("Config Read Successfully.")

	ret.Valid = true
	// read from config file
	ret.buildDSLookupInfo()
	a, _ := json.Marshal(ret)
	slog.Info(string(a))
	return ret, nil
}

func (c *RunConfig) IsValid() bool {
	return c.Valid
}

func (c *RunConfig) WaitForValidConfig(ctx context.Context) error {
	for !c.IsValid() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		slog.Info("Waiting for config to become ready...")
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (c *RunConfig) readConfigFile() error {
	// Attempt to read the file
	file, err := os.ReadFile(c.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			// Handle case where file does not exist
			slog.Error("Config:Reader:Config file not found", "file", c.Filename)
		} else {
			// Handle other errors while reading the file
			slog.Error("Config:Reader:Failed to open config file", "file", c.Filename, "error", err)
		}
		return fmt.Errorf("failed to read config file %s: %w", c.Filename, err)
	}

	// Log the file content for debugging
	slog.Info("Config:Reader:File content read", "file", c.Filename, "content", string(file))

	// Check for empty or invalid content
	if len(file) == 0 {
		slog.Error("Config:Reader:Config file is empty", "file", c.Filename)
		return fmt.Errorf("config file %s is empty", c.Filename)
	}

	// Attempt to unmarshal the JSON content
	err = json.Unmarshal(file, c)
	fmt.Print(c)
	if err != nil {
		// Log the exact error and file content for analysis
		slog.Error("Config:Reader:Failed to parse JSON", "file", c.Filename, "content", string(file), "error", err)
		return fmt.Errorf("failed to parse JSON in config file %s: %w", c.Filename, err)
	}

	// Successfully read and parsed the file
	slog.Info("Config:Reader:Config file successfully loaded", "file", c.Filename)
	return nil
}

func (c *RunConfig) saveConfigFile() error {
	// Open the file for writing
	file, err := os.OpenFile(c.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a JSON encoder and write the configuration to the file
	encoder := json.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		return err
	}
	slog.Info("Configuration successfully saved to file.")
	return nil
}

func (c *RunConfig) UpdateConfig(node *config.CacheNode, serviceList *config.DeliveryServices) {
	// may be required for config update
	if node != nil {
		c.Node = node
	}
	if serviceList != nil {
		c.ServiceList = serviceList
	}
	if c.Node != nil && c.ServiceList != nil {
		c.Valid = true
	}
	err := c.saveConfigFile()
	if err != nil {
		// handle
		slog.Error("Error saving configuration to file", "error", err)
	} else {
		slog.Info("Configuration successfully saved to file.")
	}
}

func (c *RunConfig) buildDSLookupInfo() {
	// build additional data structures as required
	// build Trie structure
}

func (c *RunConfig) BindingIp() string {
	if c.Node != nil {
		return c.Node.IP
	}
	return ""
}

func (c *RunConfig) BindingPort() int {
	if c.Node != nil {
		return c.Node.Port
	}
	return 0
}

func (c *RunConfig) ParentIp() string {
	if c.Node != nil {
		return c.Node.ParentIP
	}
	return ""
}

func (c *RunConfig) ParentPort() int {
	if c.Node != nil {
		return c.Node.ParentPort
	}
	return 0
}

func (c *RunConfig) DSLookup(req *http.Request) (*config.DeliveryService, error) {
	urlStr := helper.GetString(req)
	slog.Info(fmt.Sprintf("Searching for client URL: %s", urlStr))

	c.mu.RLock()
	defer c.mu.RUnlock()

	// Iterate over the serviceList
	for i, ds := range c.ServiceList.ServiceList {
		slog.Info(fmt.Sprintf("Comparing with ClientURL ds[%d]: %s && %s", i, ds.ClientURL, urlStr))
		// res := strings.EqualFold(urlStr, ds.ClientURL)
		res := strings.HasPrefix(urlStr, ds.ClientURL)
		if res {
			slog.Info(fmt.Sprintf("Found matching URL: %s", ds.ClientURL))
			return &ds, nil
		} else {
			slog.Info(fmt.Sprintf("Different URL Strings: %s != %s", urlStr, ds.ClientURL))
		}
	}
	return nil, errors.New("not Found")
}


// func (c *RunConfig) DSLookup(url url.URL) (*config.DeliveryService, error) {
// 	urlStr := fmt.Sprintf("%s://%s", url.Scheme, url.Host)
// 	slog.Info(fmt.Sprintf("Searching for client URL: %s", urlStr))

// 	slog.Info("DSLookup : Acquiring runConfig's read access mutex")
// 	c.mu.RLock()
// 	defer c.mu.RUnlock()

// 	var dsList *config.DeliveryServices
// 	dsList = c.ServiceList
// 	slog.Info(fmt.Sprintf("Type of dsList: %T", c.ServiceList))

// 	for i, ds := range dsList.ServiceList {
// 		slog.Info(fmt.Sprintf("List of DeliveryServices [%d]: %s ", i, ds.Name))
// 	}

// 	slog.Info(fmt.Sprintf("Type of ServiceList: %T", c.ServiceList))
// 	slog.Info(fmt.Sprintf("Type of ServiceList.ServiceList: %T", *c.ServiceList))

// 	// Log length and type of serviceList.ServiceList
// 	//slog.Info(fmt.Sprintf("Length of serviceList.ServiceList: %d", len(*c.serviceList.ServiceList)))
// 	slog.Info(fmt.Sprintf("Type of ServiceList.ServiceList: %T", c.ServiceList.ServiceList))

// 	// Iterate over the serviceList
// 	for i, ds := range c.ServiceList.ServiceList {
// 		slog.Info(fmt.Sprintf("Comparing with ClientURL ds[%d]: %s", i, ds.ClientURL))
// 		res := strings.EqualFold(urlStr, ds.ClientURL)
// 		if res {
// 			slog.Info(fmt.Sprintf("Found matching URL: %s", ds.ClientURL))
// 			return &ds, nil
// 		} else {
// 			slog.Info(fmt.Sprintf("Different URL Strings: %s != %s", urlStr, ds.ClientURL))
// 		}
// 	}
// 	return nil, errors.New("not Found")
// }
