package main

import (
	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/models"
)

func main() {
	cfg := config.New()

	// Setup renderer queues
	// Incoming queue is a queue for incoming Jobs,
	// It have limited capacity and contain Jobs to process
	// In case of channel is full client app will send "busy" signal to server
	incomingQueue := make(chan models.Job, cfg.Main.IncomingQueueLimit)

	//Debug
	go func() {
		sampleIcomingQueue(incomingQueue)
	}()

	// Outgoing queue is queue of result of Jobs (rendered pages sources)
	// It has some capacity, but it should not hit that limit ever,
	// Limit exists only for emergency cases and to do not overflow memory limits
	outgoingQueue := make(chan models.JobResult, cfg.Main.OutgoingQueueLimit)

	run(incomingQueue, outgoingQueue)
}
