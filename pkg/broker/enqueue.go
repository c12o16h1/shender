package broker

import (
	"log"
	"time"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/models"
	"github.com/davecgh/go-spew/spew"
)

func Enqueue (cacher *cache.Cacher) error {
	for {
		urls, err  := getURLs(*cacher, 5)
		if err != nil {
			log.Print(err)
		}
		spew.Dump(urls)
		time.Sleep(30 * time.Second)
	}
}

// Get non-cached URLS to enqueue them later
func getURLs(cacher cache.Cacher, amount uint) ([]string, error) {
	urls, err := cacher.Spop([]byte(models.PREFIX_ENQUEUE), amount)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, url := range urls {
		result = append(result, string(url))
	}
	return result, nil
}
