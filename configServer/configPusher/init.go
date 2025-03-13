package configPusher

import (
	"context"
	"sync"

	"github.com/hcl/cdn/configServer/inMemoryConfig"
)

var inMemConfig *inMemoryConfig.InMemoryConfig
var bgContext context.Context
var bgWg *sync.WaitGroup

func Init(ctx context.Context, wg *sync.WaitGroup, c *inMemoryConfig.InMemoryConfig) {
	inMemConfig = c
	bgContext = ctx
	bgWg = wg
}
