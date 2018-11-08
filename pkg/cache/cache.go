package cache

import (
	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/models"
)

const (
	TypeBadgerDB = "badgerdb"
	//TypeRedis  = "redis"

	ErrorUnknownDriver = models.Error("Unknown cache driver")
)

type Cacher interface {
	Set(k []byte, v []byte) error
	Get(k []byte) ([]byte, error)
	Delete(k []byte) error
	models.Closer
}

func New(config *config.CacheConfig) (Cacher, error) {
	switch config.Type {
	case TypeBadgerDB:
		return newBadgerDBCache()
	}
	return nil, ErrorUnknownDriver
}
