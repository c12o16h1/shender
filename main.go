package main

import (
	"log"

	"github.com/c12o16h1/shender/pkg/webserver"
	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/render"
	"github.com/c12o16h1/shender/pkg/models"
)

func main() {
	cfg := config.New()

	// Connect to cache
	// Cache is interface for key-value storage
	// By default it use BadgerDB cache
	// Lately support of Redis will be added
	cache, err := cache.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}
	defer cache.Close()

	// Setup renderer queues
	// Incoming queue is a queue for incoming Jobs,
	// It have limited capacity and contain Jobs to process
	// In case of channel is full client app will send "busy" signal to server
	incomingQueue := make(chan models.Job, cfg.Main.IncomingQueueLimit)
	// Outgoing queue is queue of result of Jobs (rendered pages sources)
	// It has some capacity, but it should not hit that limit ever,
	// Limit exists only for emergency cases and to do not overflow memory limits
	outgoingQueue := make(chan models.JobResult, cfg.Main.OutgoingQueueLimit)

	render.Run(incomingQueue, outgoingQueue, cfg.Render)

	// Web server
	// Is a fast Go web-server with caching
	// Which handle http requests and respond with static content (aka nginx),
	// but for SE bots return cached (rendered) page content
	if err := webserver.Serve(cfg.Main, cache); err != nil {
		log.Fatal(err)
	}
}
