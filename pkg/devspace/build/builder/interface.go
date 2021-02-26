package builder

import (
	"github.com/loft-sh/devspace/pkg/devspace/config/generated"
	"github.com/loft-sh/devspace/pkg/util/log"
)

// Interface defines methods for builders docker, kaniko and custom
type Interface interface {
	ShouldRebuild(cache *generated.CacheConfig, forceRebuild, ignoreContextPathChanges bool) (bool, error)
	Build(log log.Logger) error
}
