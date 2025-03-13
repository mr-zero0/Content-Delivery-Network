package config

import (
	"context"
	"net/http"
	"sync"
	"log/slog"
	"github.com/hcl/cdn/cacheNode/common"
	"github.com/hcl/cdn/common/config"
)

type ConfigReconciler struct {
	ctx          context.Context
	wg           *sync.WaitGroup
	activeConfig *RunConfig
	storage      common.RequestHandler
	mu           sync.Mutex // To ensure thread-safe updates
}

// NewConfigReconciler initializes a new ConfigReconciler
func NewConfigReconciler(ctx context.Context, wg *sync.WaitGroup, c *RunConfig) *ConfigReconciler {
	return &ConfigReconciler{
		ctx:          ctx,
		wg:           wg,
		activeConfig: c,
	}
}

func (cr *ConfigReconciler) SetStorage(st common.RequestHandler) {
	cr.storage = st
}

func (cr *ConfigReconciler) DoBg(req *http.Request) {
	cr.wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup) {
		defer wg.Done()
		req = req.WithContext(cr.ctx)
		if cr.storage != nil {
            cr.storage.Do(req)
        } else {
            slog.Error("Storage is not initialized")
        }
    }(cr.ctx, cr.wg)
}

// UpdateDeliveryServices updates the list of delivery services
func (cr *ConfigReconciler) UpdateDeliveryServices(serviceList *config.DeliveryServices) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	oldServiceList := cr.activeConfig.ServiceList

	// Identify services to delete
	toDelete := map[string]config.DeliveryService{}

	if oldServiceList != nil {
		for _, oldService := range oldServiceList.ServiceList {
			toDelete[oldService.Name] = oldService
		}
	}

	// Identify services to add or update
	for _, newService := range serviceList.ServiceList {
		if oldService, exists := toDelete[newService.Name]; exists {
			if oldService.ClientURL != newService.ClientURL || oldService.OriginURL != newService.OriginURL {
				// Issue DELETE for the old ClientURL
				req, _ := http.NewRequest(http.MethodDelete, oldService.ClientURL, nil)
				cr.DoBg(req)
			}
			delete(toDelete, newService.Name)
		} else {
			
		}
	}

	// Delete remaining old services
	for _, oldService := range toDelete {
		req, _ := http.NewRequest(http.MethodDelete, oldService.ClientURL, nil)
		cr.DoBg(req)
	}

	// Update the active configuration
	cr.activeConfig.UpdateConfig(nil, serviceList)
}

// UpdateCacheNode updates the configuration of a cache node
func (cr *ConfigReconciler) UpdateCacheNode(cacheNode *config.CacheNode) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	oldCacheNode := cr.activeConfig.Node
	if oldCacheNode != nil && oldCacheNode.IP == cacheNode.IP {
		if oldCacheNode.Port != cacheNode.Port || oldCacheNode.ParentIP != cacheNode.ParentIP || oldCacheNode.ParentPort != cacheNode.ParentPort {
			// Handle cache node update logic, restarting services
			// Issue DELETE or UPDATE requests to storage if required
			req, _ := http.NewRequest(http.MethodDelete, "http://"+oldCacheNode.IP, nil)
			cr.DoBg(req)
		}
	} else {
		// New cache node addition
		//req, _ := http.NewRequest(http.MethodPost, "http://"+cacheNode.IP, nil)
		//cr.DoBg(req)
	}

	// Update the active configuration
	cr.activeConfig.UpdateConfig(cacheNode, nil)
}
