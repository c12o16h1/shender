package main

import (
	"fmt"
	"log"
	"net/http"

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

	// Handler to serve files (common case)
	fsHandler := http.FileServer(http.Dir(cfg.Main.Dir))

	// Web server
	// Is a fast Go web-server with caching
	// Which handle http requests and respond with static content (aka nginx),

	// but for SE bots return cached (rendered) page content
	if err := serve(cfg.Main, cacher, fsHandler); err != nil {
		log.Fatal(err)
	}
}

func serve(config *config.MainConfig, cacher cache.Cacher, fsHandler http.Handler) error {
	http.Handle("/", pickHandler(cacher, fsHandler))
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}

func pickHandler(cacher cache.Cacher, fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If request fits requirements - process them with cache handler
		if verifiedRequest(r) {
			// Only if we have something in cache - show it and return
			if body, err := isCached(cacher, r); err != nil {
				w.Write(body)
				// Spawn goroutine to enqueue crawling
				go func(cacher cache.Cacher, url string) {
					if err := enqueue(cacher, url); err != nil {
						log.Print("can't enqueue url: ", url)
					}
				}(cacher, r.URL.String())
				return
			}
		}
		// Otherwise process with file handler
		fs.ServeHTTP(w, r)
	})
}
