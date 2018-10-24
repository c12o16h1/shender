package main

import (
	"github.com/c12o16h1/shender/pkg/webserver"
	"log"
	"os"
	"strconv"
)

var (
	DIR        = "./www"
	PORT uint16 = 80
)

func init() {
	// Set PORT from env params
	port := os.Getenv("PORT");
	if port != "" {
		p, err := strconv.Atoi(port);
		if err == nil && p > 0 {
			PORT = uint16(p)
		}
	}
	// Set DIR from env params
	dir := os.Getenv("DIR");
	if dir != "" {
		DIR = dir
	}
}

func main() {
	// Run eb server
	err := webserver.Serve(DIR, PORT);
	if err != nil {
		log.Fatal()
	}
}
