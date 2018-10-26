package config

import (
	"os"
	"strconv"
)

const (
	DEFAULT_PORT       uint16 = 80
	DEFAULT_DIR        string = "./www"
	DEFAULT_CACHE_TYPE string = "badgerdb"
)

// TODO use JSON from server, not ENV
type Configurator interface {
	Configure()
}

// As this would be global config for "microservices" in one app,
// we should operate with pointers, not values.
// F.e. if we'll need to decrease number of workers, we will change that param in config
// and "microservice" should take that by pointer
// we don't want to kill "microservice" just to renew config
// So, this is "good" global var
type Config struct {
	Configurator
	Main  *MainConfig  `json:"main"`
	Cache *CacheConfig `json:"cache"`
}

func (c *Config) Configure() {
	c.Main.Configure()
	c.Cache.Configure()
}

type MainConfig struct {
	Configurator
	Port uint16 `json:"port"`
	Dir  string `json:"dir"`
}

func (c *MainConfig) Configure() {
	c.Port = DEFAULT_PORT
	c.Dir = DEFAULT_DIR

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			c.Port = uint16(p)
		}
	}
	// Set DIR from env params
	if dir := os.Getenv("DIR"); dir != "" {
		c.Dir = dir
	}
}

type CacheConfig struct {
	Configurator
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
