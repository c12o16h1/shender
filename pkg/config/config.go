package config

import (
	"os"
	"strconv"

	"github.com/c12o16h1/shender/pkg/models"
)

const (
	DEFAULT_PORT                 uint16 = 80
	DEFAULT_DIR                  string = "./www"
	DEFAULT_INCOMING_QUEUE_LIMIT uint   = 10
	DEFAULT_OUTGOING_QUEUE_LIMIT uint   = 100

	DEFAULT_CACHE_TYPE string = "badgerdb"

	DEFAULT_RENDER_WORKERS_COUNT uint = 4
)

// As this would be global config for "microservices" in one app,
// we should operate with pointers, not values.
// F.e. if we'll need to decrease number of workers, we will change that param in config
// and "microservice" should take that by pointer
// we don't want to kill "microservice" just to renew config
// So, this is "good" global var
type Config struct {
	models.Configurator
	Main   *MainConfig   `json:"main"`
	Cache  *CacheConfig  `json:"cache"`
	Render *RenderConfig `json:"render"`
}

func (c *Config) Configure() {
	c.Main.Configure()
	c.Cache.Configure()
	c.Render.Configure()
}

type MainConfig struct {
	models.Configurator
	Port               uint16 `json:"port"`
	Dir                string `json:"dir"`
	IncomingQueueLimit uint   `json:"incoming_queue_limit"`
	OutgoingQueueLimit uint   `json:"outgoing_queue_limit"`
}

func (c *MainConfig) Configure() {
	c.Port = DEFAULT_PORT
	c.Dir = DEFAULT_DIR
	c.IncomingQueueLimit = DEFAULT_INCOMING_QUEUE_LIMIT
	c.OutgoingQueueLimit = DEFAULT_OUTGOING_QUEUE_LIMIT

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

type RenderConfig struct {
	models.Configurator
	WorkersCount uint `json:"workers_count"` // desired count of workers
}

func (c *RenderConfig) Configure() {
	c.WorkersCount = DEFAULT_RENDER_WORKERS_COUNT
	if count := os.Getenv("RENDER_WORKERS_COUNT"); count != "" {
		if cnt, err := strconv.Atoi(count); err == nil && cnt > 0 {
			c.WorkersCount = uint(cnt)
		}
	}
}

func New() *Config {
	cfg := Config{
		Main:   &MainConfig{},
		Cache:  &CacheConfig{},
		Render: &RenderConfig{},
	}
	cfg.Configure()
	return &cfg
}
