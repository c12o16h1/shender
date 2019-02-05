package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/models"
	"github.com/c12o16h1/shender/pkg/webserver"
	"github.com/gorilla/websocket"
)

func main() {
	cfg := config.New()
	// Create new Cacher connection
	cacher, err := cache.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}
	defer cacher.Close()

	//Create new fileserver
	// Handler to serve files (common case)
	fsHandler := http.FileServer(http.Dir(cfg.Main.Dir))


	// Enqueue URL to be posted to server
	// So they'll be crawled
	go enqueue(&cacher)

	// Setup renderer queues
	// Incoming queue is a queue for incoming Jobs,
	// It have limited capacity and contain Jobs to process
	// In case of channel is full client app will send "busy" signal to server
	incomingQueue := make(chan models.Job, cfg.Main.IncomingQueueLimit)
	//ws()

	//Debug
	//go func() {
	//	sampleIcomingQueue(incomingQueue)
	//}()

	// Outgoing queue is queue of result of Jobs (rendered pages sources)
	// It has some capacity, but it should not hit that limit ever,
	// Limit exists only for emergency cases and to do not overflow memory limits
	outgoingQueue := make(chan models.JobResult, cfg.Main.OutgoingQueueLimit)
	// Run broker in goroutine
	go run(incomingQueue, outgoingQueue)

	// Web server
	// Is a fast Go web-server with caching
	// Which handle http requests and respond with static content (aka nginx),

	// but for SE bots return cached (rendered) page content
	if err := serve(cfg.Main, cacher, fsHandler); err != nil {
		log.Fatal(err)
	}
}

func serve(config *config.MainConfig, cacher cache.Cacher, fsHandler http.Handler) error {
	http.Handle("/", webserver.PickHandler(cacher, fsHandler))
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}


//Garbage

var addr = flag.String("addr", "localhost:8080", "http service address")

func ws() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			log.Print(t)
			msg := models.WSMessage{
				Type:    models.TypeEnqueueURL,
				Message: "http://google.com",
			}
			b, err := json.Marshal(msg)
			if err != nil {
				log.Println("json.Marshal:", err)
				return
			}
			err = c.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
