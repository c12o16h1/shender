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

	// Handler to serve files (common case)
	fsHandler := fileserver.New(fileserver.Options{}).Serve(http.Dir(cfg.Main.Dir))

	//Handler to serve cache (for SE bots)
	cacheHandler := NewCacheHandler(&cacher)

	// Web server
	// Is a fast Go web-server with caching
	// Which handle http requests and respond with static content (aka nginx),

	// but for SE bots return cached (rendered) page content
	if err := serve(cfg.Main, cacheHandler, fsHandler); err != nil {
		log.Fatal(err)
	}
}

func serve(config *config.MainConfig, handler CacheHandler, fsHandler http.Handler) error {
	http.Handle("/", pickHandler(handler, fsHandler))
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}

func pickHandler(cache CacheHandler, fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If request fits requirements - process them with cache handler
		if cache.Verify(r) {
			cache.ServeHTTP(w, r)
			return
		}
		// Otherwise process with file handler
		fs.ServeHTTP(w, r)
	})
}
