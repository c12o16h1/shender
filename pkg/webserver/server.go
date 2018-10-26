package webserver

import (
	"net/http"
	"fmt"

	"github.com/yosssi/go-fileserver"

	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/cache"
)

func Serve(config *config.MainConfig, cache cache.Cacher) error {
	fs := fileserver.New(fileserver.Options{})
	http.Handle("/", fs.Serve(http.Dir(config.Dir)))
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
