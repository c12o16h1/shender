package webserver

import (
	"log"
	"net/http"
	"strings"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/pkg/errors"
)

const (
	HEADER_ACCEPT_ALL  = "*/*"
	HEADER_ACCEPT_HTML = "text/html"
)

var (
	ERR_NOT_CACHED = "not cached"

	maxExtensionLengt = 4      // Max length of file extension
	html              = "html" // file extension for html files
	dotByte           = "."[0] // byte for dot
)

func PickHandler(cacher cache.Cacher, fs http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If request fits requirements - process them with cache handler
		if verifiedRequest(r) {
			// Only if we have something in cache - show it and return
			body, err := isCached(cacher, r)
			if err != nil {
				// Spawn goroutine to enqueue crawling
				go func(cacher cache.Cacher, url string) {
					if err := enqueue(cacher, url); err != nil {
						log.Print("can't enqueue url: ", url)
					}
				}(cacher, urlFromRequest(r))
				// Process with file handler
				fs.ServeHTTP(w, r)
				return
			}
			// Show cached content
			w.Write(body)
			return
		}
		// Default process with file handler
		log.Print("FC")
		fs.ServeHTTP(w, r)
	})
}

func verifiedRequest(r *http.Request) bool {
	if isBot(r) && isHTML(r) && !isFile(r) {
		return true
	}
	return false
}

func isCached(cacher cache.Cacher, r *http.Request) ([]byte, error,) {
	body, err := cacher.Get([]byte(urlFromRequest(r)))
	if err != nil || len(body) == 0 {
		return nil, errors.Wrap(err, ERR_NOT_CACHED)
	}
	return body, nil
}

// TODO all below move to separate place

//If it's not bot - lets' ignore request
func isBot(r *http.Request) bool {
	return true // Debug purposes only
	if r.UserAgent() == "googlebot" {
		return true
	}
	return false
}

// If client asking not for HTML - lets ignore this request
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

// Here we check that this is web page, not a file (image or css f.e.)
// but we do exception for HTML files
func isFile(r *http.Request) bool {
	max := len(r.RequestURI) - 1
	min := max - maxExtensionLengt
	// Look on URL from the end to beginning
	// And try to find is it file or
	// And yes, we have loop, that's just faster and non consume any additional memory
	// Regexp is not best option for such small operation
	var extension string
	for i := max; i > 0 && i >= min; i-- {
		// If we found dot, so this is a file
		// if file extension is not allowed - this is file
		if r.RequestURI[i] == dotByte && extension != html {
			return true
		}
		//Prepend symbol, because we going from end to beginning
		extension = string(r.RequestURI[i]) + extension
	}
	return false
}

func urlFromRequest(r *http.Request) string {
	return r.Host + r.RequestURI
}
