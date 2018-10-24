package webserver

import (
	"github.com/yosssi/go-fileserver"
	"net/http"
	"fmt"
)

func Serve(dir string, port uint) error {
	fs := fileserver.New(fileserver.Options{})
	http.Handle("/", fs.Serve(http.Dir(dir)))
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
