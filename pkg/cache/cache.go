package cache

type Error string

func (e Error) Error() string { return string(e) }

type DriverType string

const (
	TypeBadgerDB DriverType = "badgerdb"
	//TypeRedis    DriverType = "redis"

	ErrorUnknownDriver = Error("Unknown cache driver")
)

type Cacher interface {
	Set(k []byte, v []byte) error
	Get(k []byte) ([]byte, error)
	Delete(k []byte) error
	Close() error
}

func New(t DriverType) (Cacher, error) {
	switch t {
	case TypeBadgerDB:
		return newBadgerDBCache()
	}
	return nil, ErrorUnknownDriver
}
