package webserver

import (
	"github.com/yosssi/go-fileserver"
	"net/http"
	"fmt"
	"github.com/c12o16h1/shender/pkg/cache"
)

func Serve(dir string, port uint16, cache cache.Cache) error {
	fs := fileserver.New(fileserver.Options{})
	http.Handle("/", fs.Serve(http.Dir(dir)))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
