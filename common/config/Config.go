package config

const (
	HdrReWriteOpAdd = iota
	HdrReWriteOpOverwrite
	HdrReWriteOpDelete
)

const (
	CacheNodeMid  = "Mid"
	CacheNodeEdge = "Edge"
)

// Rule for rewriting HTTP headers
type RewriteRule struct {
	HeaderName string `json:"headerName"` //name of the http header
	Operation  int    `json:"operation"`  //HdrRewriteOp
	Value      string `json:"value"`      //Value for Add/Overwrite
}

// One Deliver Service
type DeliveryService struct {
	Name         string        `json:"name"`         //name of the DS ... cannot be updated
	ClientURL    string        `json:"clientURL"`    //URL from Client
	OriginURL    string        `json:"originURL"`    //URL to Origin
	RewriteRules []RewriteRule `json:"rewriteRules"` //ReWrite Rules
}

// One Cache Node
type CacheNode struct {
	Name       string `json:"name"`               // name of the cacheNode
    IP         string `json:"ip"`                 // IP address of the CacheNode ... cannot be updated
    Port       int    `json:"port"`               // Port for the CacheNode to listen
    Type       string `json:"type"`               // Type of the node
    ParentIP   string `json:"parentIP"`           // IP of upstream cache, "" if no upstream is present
    ParentPort int    `json:"parentPort"`         // Port for the CacheNode where upstream cache is listening
    MgmtPort   int    `json:"mgmtPort,omitempty"` //management port
	PromPort   int    `json:"promPort,omitempty"` // prometheus port
}
// List of Cache Nodes
type CacheNodes struct {
	Version  int         `json:"version"` //internally maintained version number
	NodeList []CacheNode `json:"nodeList"`
}

// List of Delivery Services
type DeliveryServices struct {
	Version     int               `json:"version"` //Internally maintained version number
	ServiceList []DeliveryService `json:"serviceList"`
}

// // configuration
// type RunConfig struct {
// 	valid       bool          `json:"-"`              // is config valid?
// 	filename    string          `json:"-"`            // name of the file to persist config
// 	node        *config.CacheNode           `json:"node"`        // cache node details
// 	serviceList *config.DeliveryServices    `json:"serviceList"` // list of deliver services
// }

// Config represents the configuration for the cache node
type Config struct {
	ServiceList []*DeliveryService `json:"serviceList"` // List of delivery services
	Node        *CacheNode         `json:"node"`        // Cache node details
}
