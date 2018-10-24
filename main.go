package main

import (
	"log"
	"os"
	"strconv"

	"github.com/c12o16h1/shender/pkg/webserver"
	"github.com/c12o16h1/shender/pkg/cache"
)

var (
	DIR                    = "./www"
	PORT  uint16           = 80
	CACHE cache.DriverType = cache.TypeBadgerDB
)

func init() {
	// Set PORT from env params
	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil && p > 0 {
			PORT = uint16(p)
		}
	}
	// Set DIR from env params
	if dir := os.Getenv("DIR"); dir != "" {
		DIR = dir
	}
	// Set CACHE from env params
	if c := os.Getenv("CACHE"); c != "" {
		CACHE = cache.DriverType(c)
	}
}

func main() {
	// Connect to cache
	c, err := cache.New(CACHE)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	// Run web server
	if err := webserver.Serve(DIR, PORT, c); err != nil {
		log.Fatal(err)
	}
}
