package main

import (
	"net/http"
	"strings"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
)

type CacheHandler struct {
	http.Handler
	cache  *cache.Cacher
	config *config.MainConfig
}

func NewCacheHandler(cache *cache.Cacher) CacheHandler {
	return CacheHandler{
		cache: cache,
	}
}

func (h CacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (h CacheHandler) Verify(r *http.Request) bool {
	if isBot(r) && isHTML(r) {
		return true
	}
	return false
}

// TODO all below move to separate place

func isBot(r *http.Request) bool {
	return true // Debug purposes only
	if r.UserAgent() == "googlebot" {
		return true
	}
	return false
}

const (
	HEADER_ACCEPT_ALL  = "*/*"
	HEADER_ACCEPT_HTML = "text/html"
)

func isHTML(r *http.Request) bool {
	accept := r.Header["Accept"]
	for _, acc := range accept {
		if acc == HEADER_ACCEPT_ALL {
			return true
		}
		if strings.Contains(acc, HEADER_ACCEPT_HTML) {
			return true
		}
	}
	return false
}
