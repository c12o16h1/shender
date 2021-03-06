package config

import (
	"os"
	"strconv"

	"github.com/c12o16h1/shender/pkg/models"
)

const (
	DEFAULT_PORT                 uint16 = 80
	DEFAULT_DIR                  string = "./www"
	DEFAULT_INCOMING_QUEUE_LIMIT uint   = 20
	DEFAULT_OUTGOING_QUEUE_LIMIT uint   = 100
	DEFAULT_WS_HOST                     = "localhost:8080"

	DEFAULT_CACHE_TYPE string = "badgerdb"
)

// As this would be global config for "microservices" in one app,
// we should operate with pointers, not values.
// F.e. if we'll need to decrease number of workers, we will change that param in config
// and "microservice" should take that by pointer
// we don't want to kill "microservice" just to renew config
// So, this is "good" global var
type Config struct {
	models.Configurator
	Main  *MainConfig  `json:"main"`
	Cache *CacheConfig `json:"cache"`
}

func (c *Config) Configure() {
	c.Main.Configure()
	c.Cache.Configure()
}

type MainConfig struct {
	models.Configurator
	Port               uint16 `json:"port"`
	Dir                string `json:"dir"`
	IncomingQueueLimit uint   `json:"incoming_queue_limit"`
	OutgoingQueueLimit uint   `json:"outgoing_queue_limit"`
	WSHost             string `json:"ws_host"`
}

func (c *MainConfig) Configure() {
	c.Port = DEFAULT_PORT
	c.Dir = DEFAULT_DIR
	c.IncomingQueueLimit = DEFAULT_INCOMING_QUEUE_LIMIT
	c.OutgoingQueueLimit = DEFAULT_OUTGOING_QUEUE_LIMIT
	c.WSHost = DEFAULT_WS_HOST

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			c.Port = uint16(p)
		}
	}
	if dir := os.Getenv("DIR"); dir != "" {
		c.Dir = dir
	}

	if iql := os.Getenv("DEFAULT_INCOMING_QUEUE_LIMIT"); iql != "" {
		if l, err := strconv.Atoi(iql); err == nil && l > 0 {
			c.IncomingQueueLimit = uint(l)
		}
	}

	if oql := os.Getenv("DEFAULT_OUTGOING_QUEUE_LIMIT"); oql != "" {
		if l, err := strconv.Atoi(oql); err == nil && l > 0 {
			c.IncomingQueueLimit = uint(l)
		}
	}

	if h := os.Getenv("WS_HOST"); h != "" {
		c.Dir = h
	}
}

type CacheConfig struct {
	models.Configurator
	Type string `json:"type"`
	Host string `json:"host"`
	Port string `json:"port"`
	User string `json:"user"`
	Pass string `json:"password"`
}

func (c *CacheConfig) Configure() {
	c.Type = DEFAULT_CACHE_TYPE
	if ctype := os.Getenv("CACHE_TYPE"); ctype != "" {
		c.Type = ctype
	}
}

func New() *Config {
	cfg := Config{
		Main:  &MainConfig{},
		Cache: &CacheConfig{},
	}
	cfg.Configure()
	return &cfg
}
