package configSaver

import (
	"context"
	"sync"

	"github.com/hcl/cdn/configServer/inMemoryConfig"
)

var saveDir string

var inMemConfig *inMemoryConfig.InMemoryConfig

var bgContext context.Context
var bgWg *sync.WaitGroup

func Init(ctx context.Context, wg *sync.WaitGroup, c *inMemoryConfig.InMemoryConfig, dir string) {
	inMemConfig = c
	bgContext = ctx
	bgWg = wg
	saveDir = dir
	LoadDSFromFile()
	LoadCNFromFile()
}
