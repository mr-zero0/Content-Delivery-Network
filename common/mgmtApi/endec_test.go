package mgmtApi

import (
	"reflect"
	"testing"

	"github.com/hcl/cdn/common/config" // Import the internal config package
)

// Test ConfigToProtoDS and ProtoToConfigDS
func TestConfigToProtoDSAndBack(t *testing.T) {
	// Create a sample DeliveryServices object
	original := &config.DeliveryServices{
		ServiceList: []config.DeliveryService{
			{
				Name:      "Service1",
				ClientURL: "http://client1.example.com",
				OriginURL: "http://origin1.example.com",
				RewriteRules: []config.RewriteRule{
					{HeaderName: "Host", Operation: 1, Value: "example.com"},
				},
			},
			{
				Name:      "Service2",
				ClientURL: "http://client2.example.com",
				OriginURL: "http://origin2.example.com",
				RewriteRules: []config.RewriteRule{
					{HeaderName: "Referer", Operation: 2, Value: "http://example.com"},
				},
			},
		},
	}

	// Convert to protobuf
	protoResult := ConfigToProtoDS(original)

	// Convert back to internal format
	finalResult := ProtoToConfigDS(protoResult)

	// Compare the original and final result
	if !reflect.DeepEqual(original, finalResult) {
		t.Errorf("Mismatch after conversion. Got %+v, expected %+v", finalResult, original)
	}
}

// Test ConfigToProtoCN and ProtoToConfigCN
func TestConfigToProtoCNAndBack(t *testing.T) {
	// Create a sample CacheNode
	original := &config.CacheNode{
		IP:         "192.168.1.1",
		Port:       8080,
		Type:       "Edge",
		ParentIP:   "192.168.1.2",
		ParentPort: 80,
	}

	// Convert to protobuf
	protoResult := ConfigToProtoCN(original)

	// Convert back to internal format
	finalResult := ProtoToConfigCN(protoResult)

	// Compare the original and final result
	if !reflect.DeepEqual(original, finalResult) {
		t.Errorf("Mismatch after conversion. Got %+v, expected %+v", finalResult, original)
	}
}

// Test edge cases for nil inputs
func TestNilInputs(t *testing.T) {
	if ConfigToProtoDS(nil) != nil {
		t.Error("ConfigToProtoDS(nil) should return nil")
	}

	if ProtoToConfigDS(nil) != nil {
		t.Errorf("ProtoToConfigDS(nil) should return nil, got %+v", ProtoToConfigDS(nil))
	}

	if ConfigToProtoCN(nil) != nil {
		t.Error("ConfigToProtoCN(nil) should return nil")
	}

	if ProtoToConfigCN(nil) != nil {
		t.Errorf("ProtoToConfigCN(nil) should return nil, got %v", ProtoToConfigCN(nil))
	}
}
