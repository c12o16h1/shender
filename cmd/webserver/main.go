package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/yosssi/go-fileserver"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
)

func main() {
	cfg := config.New()

	// Connect to cacher
	// Cache is interface for key-value storage
	// By default it use BadgerDB cacher
	// Lately support of Redis will be added
	cacher, err := cache.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}
	defer cacher.Close()

	// Web server
	// Is a fast Go web-server with caching
	// Which handle http requests and respond with static content (aka nginx),
	// but for SE bots return cached (rendered) page content
	if err := serve(cfg.Main, cacher); err != nil {
		log.Fatal(err)
	}

}

func serve(config *config.MainConfig, cache cache.Cacher) error {
	fs := fileserver.New(fileserver.Options{})
	http.Handle("/", fs.Serve(http.Dir(config.Dir)))
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
