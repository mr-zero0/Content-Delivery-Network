package inMemoryConfig

import (
	"encoding/json"
	"errors"
	"io"
	"sync"

	"github.com/hcl/cdn/common/config"
	"github.com/hcl/cdn/common/mgmtApi"
)

// Holds the config elements in Memory
type InMemoryConfig struct {
	dsMutex          sync.RWMutex            //mutex for update to DeliveryServices
	deliveryServices config.DeliveryServices //list of delivery services
	cnMutex          sync.RWMutex            //mutex for update to CacheNodes
	cacheNodes       config.CacheNodes       //list of CacheNodes
}

// Serialize DeliveryServices to writer
func (i *InMemoryConfig) SaveDsTo(w io.Writer) error {
	i.dsMutex.RLock()
	defer i.dsMutex.RUnlock()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	if err := encoder.Encode(i.deliveryServices); err != nil {
		return err
	}
	return nil
}

// Serialize ConfigNodes to writer
func (i *InMemoryConfig) SaveCnTo(w io.Writer) error {
	i.cnMutex.RLock()
	defer i.cnMutex.RUnlock()
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Pretty-print JSON
	if err := encoder.Encode(i.cacheNodes); err != nil {
		return err
	}
	return nil
}

// De-serialize DeliveryServices from reader
func (i *InMemoryConfig) ReadDsFrom(r io.Reader) error {
	i.dsMutex.Lock()
	defer i.dsMutex.Unlock()
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&i.deliveryServices); err != nil {
		return err
	}
	return nil
}

// De-serialize ConfigNodes from reader
func (i *InMemoryConfig) ReadCnFrom(r io.Reader) error {
	i.cnMutex.Lock()
	defer i.cnMutex.Unlock()
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&i.cacheNodes); err != nil {
		return err
	}
	return nil
}

// return Version & list of DeliveryService Names
func (i *InMemoryConfig) GetDsNames() (int, []string) {
	i.dsMutex.RLock()
	defer i.dsMutex.RUnlock()
	//Iterate and return only names
	names := make([]string, len(i.deliveryServices.ServiceList))
	for idx, ds := range i.deliveryServices.ServiceList {
		names[idx] = ds.Name
	}
	return i.deliveryServices.Version, names
}

// return Version & list of Config Node IPs
func (i *InMemoryConfig) GetCnNames() (int, []string) {
    i.cnMutex.RLock()
    defer i.cnMutex.RUnlock()
    names := make([]string, len(i.cacheNodes.NodeList))
    for idx, cn := range i.cacheNodes.NodeList {
        names[idx] = cn.Name
    }
    return i.cacheNodes.Version, names
}

// return all details of one DeliveryService
func (i *InMemoryConfig) GetDsDetail(dsName string) (*config.DeliveryService, error) {
	i.dsMutex.RLock()
	defer i.dsMutex.RUnlock()
	//return one config if present or nil
	for _, ds := range i.deliveryServices.ServiceList {
		if ds.Name == dsName {
			return &ds, nil
		}
	}
	return nil, errors.New("delivery service not found")
}

// return all details of one ConfigNode
func (i *InMemoryConfig) GetCnDetailByName(name string) (*config.CacheNode, error) {
    i.cnMutex.RLock()
    defer i.cnMutex.RUnlock()
    for _, cn := range i.cacheNodes.NodeList {
        if cn.Name == name {
            return &cn, nil
        }
    }
    return nil, errors.New("cache node not found")
}
// Add one DeliveryService
func (i *InMemoryConfig) AddDs(newService *config.DeliveryService) error {
	i.dsMutex.Lock()
	defer i.dsMutex.Unlock()
	i.deliveryServices.ServiceList = append(i.deliveryServices.ServiceList, *newService)
	i.deliveryServices.Version++
	return nil
}

// Update an existing DeliveryService
func (i *InMemoryConfig) UpdateDs(updateService *config.DeliveryService) error {
	i.dsMutex.Lock()
	defer i.dsMutex.Unlock()
	for idx, ds := range i.deliveryServices.ServiceList {
		if ds.Name == updateService.Name {
			i.deliveryServices.ServiceList[idx] = *updateService
			i.deliveryServices.Version++
			return nil
		}
	}
	return errors.New("delivery service not found")
}

// Delete an existing DeliveryService
func (i *InMemoryConfig) DeleteDs(name string) error {
	i.dsMutex.Lock()
	defer i.dsMutex.Unlock()
	for idx, ds := range i.deliveryServices.ServiceList {
		if ds.Name == name {
			i.deliveryServices.ServiceList = append(i.deliveryServices.ServiceList[:idx], i.deliveryServices.ServiceList[idx+1:]...)
			i.deliveryServices.Version++
			return nil
		}
	}
	return errors.New("delivery service not found")
}

// Add one ConfigNode
func (i *InMemoryConfig) AddCn(newNode *config.CacheNode) error {
    i.cnMutex.Lock()
    defer i.cnMutex.Unlock()
    i.cacheNodes.NodeList = append(i.cacheNodes.NodeList, *newNode)
    i.cacheNodes.Version++
    return nil
}

// Update an existing ConfigNode
func (i *InMemoryConfig) UpdateCn(updateNode *config.CacheNode) error {
    i.cnMutex.Lock()
    defer i.cnMutex.Unlock()
    for idx, cn := range i.cacheNodes.NodeList {
        if cn.Name == updateNode.Name {
            i.cacheNodes.NodeList[idx] = *updateNode
            i.cacheNodes.Version++
            return nil
        }
    }
    return errors.New("cache node not found")
}

// Delete an existing ConfigNode
func (i *InMemoryConfig) DeleteCn(name string) error {
    i.cnMutex.Lock()
    defer i.cnMutex.Unlock()
    for idx, cn := range i.cacheNodes.NodeList {
        if cn.Name == name {
            i.cacheNodes.NodeList = append(i.cacheNodes.NodeList[:idx], i.cacheNodes.NodeList[idx+1:]...)
            i.cacheNodes.Version++
            return nil
        }
    }
    return errors.New("cache node not found")
}

// Convert Config DS data to Proto data
func (i *InMemoryConfig) ConfigToProtoDS() []*mgmtApi.DeliveryService {
	i.dsMutex.RLock()
	defer i.dsMutex.RUnlock()
	return mgmtApi.ConfigToProtoDS(&i.deliveryServices)
}

// Convert Config CN data to Proto data
func (i *InMemoryConfig) ConfigToProtoCN(name string) (int, *mgmtApi.CacheNode, error) {
    i.cnMutex.RLock()
    defer i.cnMutex.RUnlock()
    // Lookup node by name
    var foundNode *config.CacheNode
    for _, node := range i.cacheNodes.NodeList {
        if node.Name == name {
            foundNode = &node
            break
        }
    }
    if foundNode == nil {
        return 0, nil, errors.New("name not found")
    }
    return foundNode.MgmtPort, mgmtApi.ConfigToProtoCN(foundNode), nil
}