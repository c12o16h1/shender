package main

import (
	"net/http"
	"strings"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/davecgh/go-spew/spew"
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


func verifiedRequest(r *http.Request) bool {
	if isBot(r) && isHTML(r) && !isFile(r) {
		return true
	}
	return false
}

func isCached(cacher cache.Cacher, r *http.Request) ([]byte, error,) {
	spew.Dump(r.URL.String())
	body, err := cacher.Get([]byte(r.URL.String()))
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
