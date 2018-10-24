package main

import (
	"github.com/c12o16h1/shender/pkg/webserver"
	"log"
	"os"
	"strconv"
)

var (
	DIR         = "./www"
	PORT uint16 = 80
)

func init() {
	// Set PORT from env params
	port := os.Getenv("PORT");
	if port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			PORT = uint16(p)
		}
	}
	// Set DIR from env params
	if dir := os.Getenv("DIR"); dir != "" {
		DIR = dir
	}
}

func main() {
	// Run web server
	if err := webserver.Serve(DIR, PORT); err != nil {
		log.Fatal()
	}
}
