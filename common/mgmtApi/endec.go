package mgmtApi

import (
	"github.com/hcl/cdn/common/config" // Internal config package
)

// ConfigToProtoDS converts a config.DeliveryServices object to a protobuf-compatible list of DeliveryService messages.
func ConfigToProtoDS(serviceList *config.DeliveryServices) []*DeliveryService {
	// Return nil if input is nil
	if serviceList == nil {
		return nil
	}

	// Initialize the return slice with the same length as the service list
	ret := make([]*DeliveryService, len(serviceList.ServiceList))

	// Iterate over each service in the input list
	for i, service := range serviceList.ServiceList {

		// Create a new protobuf DeliveryService
		protoService := &DeliveryService{
			Name:         service.Name,
			ClientURL:    service.ClientURL,
			OriginURL:    service.OriginURL,
			RewriteRules: make([]*RewriteRule, len(service.RewriteRules)),
		}

		// Iterate over rewrite rules for the service
		for j, rule := range service.RewriteRules {

			// Create a new protobuf RewriteRule
			protoRule := &RewriteRule{
				HeaderName: rule.HeaderName,
				Operation:  int32(rule.Operation),
				Value:      rule.Value,
			}
			protoService.RewriteRules[j] = protoRule
		}

		ret[i] = protoService
	}

	return ret
}

// ProtoToConfigDS converts the protobuf DeliveryService slice to the internal DeliveryServices struct
func ProtoToConfigDS(serviceList []*DeliveryService) *config.DeliveryServices {
	// Return nil if the input is nil
	if serviceList == nil {
		return nil
	}

	// Initialize the internal DeliveryServices structure
	internalServiceList := &config.DeliveryServices{
		ServiceList: make([]config.DeliveryService, len(serviceList)),
	}

	// Iterate over each protobuf service
	for i, protoService := range serviceList {
		if protoService == nil { // Skip nil entries to avoid panics
			continue
		}

		// Create an internal DeliveryService object
		internalService := config.DeliveryService{
			Name:         protoService.Name,
			ClientURL:    protoService.ClientURL,
			OriginURL:    protoService.OriginURL,
			RewriteRules: make([]config.RewriteRule, len(protoService.RewriteRules)),
		}

		// Iterate over rewrite rules for the protobuf service
		for j, protoRule := range protoService.RewriteRules {
			if protoRule == nil { // Skip nil rules
				continue
			}

			// Create an internal RewriteRule object
			internalRule := config.RewriteRule{
				HeaderName: protoRule.HeaderName,
				Operation:  int(protoRule.Operation), // Convert int32 to int
				Value:      protoRule.Value,
			}
			internalService.RewriteRules[j] = internalRule
		}

		// Assign the converted service to the internal list
		internalServiceList.ServiceList[i] = internalService
	}

	return internalServiceList
}

// ConfigToProto converts the internal Config struct to the protobuf Config message
func ConfigToProtoCN(cacheNode *config.CacheNode) *CacheNode {
	// Return nil if the input is nil
	if cacheNode == nil {
		return nil
	}
	ret := &CacheNode{
		Name:       cacheNode.Name,
		Ip:         cacheNode.IP,
		Port:       int32(cacheNode.Port),
		Type:       cacheNode.Type,
		ParentIP:   cacheNode.ParentIP,
		ParentPort: int32(cacheNode.ParentPort),
	}
	return ret
}

// ProtoToConfigCN converts a protobuf CacheNode message to the internal Config.CacheNode struct.
func ProtoToConfigCN(cacheNode *CacheNode) *config.CacheNode {
	// Return nil if the input is nil
	if cacheNode == nil {
		return nil
	}

	// Map the fields from the protobuf CacheNode to config.CacheNode
	ret := &config.CacheNode{
		Name:       cacheNode.Name,
		IP:         cacheNode.Ip,
		Port:       int(cacheNode.Port),
		Type:       cacheNode.Type,
		ParentIP:   cacheNode.ParentIP,
		ParentPort: int(cacheNode.ParentPort),
	}
	return ret
}
