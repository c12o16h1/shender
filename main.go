package main

import (
	"log"

	"github.com/c12o16h1/shender/pkg/webserver"
	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/render"
	"github.com/c12o16h1/shender/pkg/models"
)

func main() {
	cfg := config.New()

	// Connect to cache
	cache, err := cache.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	var renderBuffer chan []models.Job
	render.Run(renderBuffer, cfg.Render)

	// Run web server
	if err := webserver.Serve(cfg.Main, cache); err != nil {
		log.Fatal(err)
	}
}
