package main

import (
	"log"

	"github.com/c12o16h1/shender/pkg/webserver"
	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
)

func main() {
	cfg := config.New()

	// Connect to cache
	cache, err := cache.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	// Run web server
	if err := webserver.Serve(cfg.Main, cache); err != nil {
		log.Fatal(err)
	}
}
