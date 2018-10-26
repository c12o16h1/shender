package cache

import "github.com/c12o16h1/shender/pkg/config"

type Error string

func (e Error) Error() string { return string(e) }

const (
	TypeBadgerDB = "badgerdb"
	//TypeRedis  = "redis"

	ErrorUnknownDriver = Error("Unknown cache driver")
)

type Cacher interface {
	Set(k []byte, v []byte) error
	Get(k []byte) ([]byte, error)
	Delete(k []byte) error
	Close() error
}

func New(config *config.CacheConfig) (Cacher, error) {
	switch config.Type {
	case TypeBadgerDB:
		return newBadgerDBCache()
	}
	return nil, ErrorUnknownDriver
}
