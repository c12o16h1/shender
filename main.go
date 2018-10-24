package main

import (
	"github.com/c12o16h1/shender/pkg/webserver"
	"log"
)

const (
	DIR = "/Users/preston/go/src/github.com/c12o16h1/shender/www"
	PORT = 8080
)

func main() {
	// Run eb server
	err := webserver.Serve(DIR, PORT); if err != nil {
		log.Fatal()
	}
}