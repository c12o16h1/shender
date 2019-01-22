package main

import (
	"time"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/models"
)

const ENQUEUE_EXPIRY_TIME = 24 * time.Hour

func enqueue(cacher cache.Cacher, url string) error {
	return cacher.Setex([]byte(models.PREFIX_ENQUEUE+url), ENQUEUE_EXPIRY_TIME, nil)
}
