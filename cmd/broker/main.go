package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/c12o16h1/shender/pkg/broker"
	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/config"
	"github.com/c12o16h1/shender/pkg/models"
	"github.com/c12o16h1/shender/pkg/webserver"
	"github.com/gorilla/websocket"
)

var (
	shortSleeper = 1 * time.Second
	)

func main() {
	// Initialization
	cfg := config.New()
	appID := "qwerty" // TODO move to config
	// Create new Cacher connection
	cacher, err := cache.New(cfg.Cache)
	if err != nil {
		log.Fatal(err)
	}
	defer cacher.Close()

	// Create new fileserver
	// Handler to serve files (common case)
	fsHandler := http.FileServer(http.Dir(cfg.Main.Dir))

	// Setup renderer queues
	// Incoming queue is a queue for incoming Jobs,
	// It have limited capacity and contain Jobs to process
	// In case of channel is full client app will send "busy" signal to server
	incomingQueue := make(chan models.Job, cfg.Main.IncomingQueueLimit)

	// Outgoing queue is queue of result of Jobs (rendered pages sources)
	// It has some capacity, but it should not hit that limit ever,
	// Limit exists only for emergency cases and to do not overflow memory limits
	outgoingQueue := make(chan models.JobResult, cfg.Main.OutgoingQueueLimit)

	// Queue for passing cache from websocket listener to cache DB
	storagerQueue := make(chan models.DataResponseCachedPage, cfg.Main.IncomingQueueLimit)

	// Sleeper channels to pause in execution in some routines in case of error
	// Chan to pause request to get new urls for crawling
	sleeperRequestGetUrls := make(chan time.Duration, 1)
	// Chan to pause request to push crawled content
	sleeperResponseCachedPage := make(chan time.Duration, 1)
	// Chan to pause request to enqueue urls to be crawled
	sleeperTypeRequestSendURL := make(chan time.Duration, 1)
	// Chan to pause request to get cached pages from server
	sleeperRequestCachedPage := make(chan time.Duration, 1)

	// Service channels
	// Channel to trigger renew of websockets
	renewWebsockets := make(chan int, 1)

	/*
	Establishing WS connection to main server
	 */

	// Service

	// Create/renew websockets connection
	var wsc *models.WSConn
	defer wsc.Close()
	u := url.URL{Scheme: "ws", Host: cfg.Main.WSHost, Path: ""}
	go func() {
		for {
			<-renewWebsockets
			oldConn := wsc
			if oldConn != nil {
				oldConn.Close()
			}
			wsc = nil
			ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
			if err != nil {
				log.Print("dial:", err)
				time.Sleep(1 * time.Second)
				// try again
				renewWebsockets <- 0
				continue
			}
			wsc = models.NewWSConn(ws)
		}
	}()
	// Intially create websockets conn
	renewWebsockets <- 0

	// Processing

	/*
	Spawn crawl-oriented goroutines.
	This Goroutines should not cause application exit in any obstacles,
	because them isn't critically important.
	Most important thing is webserver.
	 */

	/*
   Spawn goroutine to listen all messages from server
   And properly handle them
	*/
	go func() {
		for {
			// Wait and go next loop if ws is unhealthy
			if wsc == nil {
				time.Sleep(1 * time.Second)
				log.Print("wsc is nil")
				continue
			}
			if err := broker.Listen(
				wsc,
				incomingQueue,
				storagerQueue,
				sleeperRequestGetUrls,
				sleeperResponseCachedPage,
				sleeperTypeRequestSendURL,
				sleeperRequestCachedPage,
			); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	// Other Apps pages crawling
	/*
	Spawn goroutine to get URLs for crawling from server
	so they'll be crawled by this app.
	*/
	go func() {
		for {
			// Wait and go next loop if ws is unhealthy
			if wsc == nil {
				time.Sleep(1 * time.Second)
				log.Print("wsc is nil")
				continue
			}
			if err := broker.Request(wsc, incomingQueue, sleeperRequestGetUrls); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	/*
	Spawn goroutine to process crawling of pages for other members of system.
	This goroutine ensure that server has enough resources to do render,
	spawn new renderer instance, connect to them via RPC, and do render for URL from incoming queue.
	Then save push result to outgoing queue
	 */
	go func() {
		for {
			if err := broker.Crawl(incomingQueue, outgoingQueue); err != nil {
				log.Print("Crawl: ", err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	/*
	Spawn goroutine to push content of crawled pages to server
	*/
	go func() {
		for {
			// Wait and go next loop if ws is unhealthy
			if wsc == nil {
				time.Sleep(1 * time.Second)
				log.Print("wsc is nil")
				continue
			}
			if err := broker.Push(wsc, outgoingQueue, sleeperResponseCachedPage); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	// This App cache
	/*
	Spawn goroutine to send/enqueue URL to central server
	so they'll be crawled by other members.
	*/
	go func() {
		for {
			// Wait and go next loop if ws is unhealthy
			if wsc == nil {
				time.Sleep(1 * time.Second)
				log.Print("wsc is nil")
				continue
			}
			if err := broker.Enqueue(&cacher, wsc, appID, sleeperTypeRequestSendURL); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	/*
	Spawn goroutine to get cached pages from central server
	so bots may see cached pages content
	*/
	go func() {
		for {
			// Wait and go next loop if ws is unhealthy
			if wsc == nil {
				time.Sleep(1 * time.Second)
				log.Print("wsc is nil")
				continue
			}
			if err := broker.RequestCache(wsc, appID, sleeperRequestCachedPage, renewWebsockets); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	/*
	Spawn goroutine to save cache in local cache DB
	so bots may see cached pages content
	*/
	go func() {
		for {
			if err := broker.Storage(&cacher, storagerQueue, sleeperRequestCachedPage); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}()

	// Testing part

	//Debug
	//go func() {
	//	sampleIcomingQueue(incomingQueue)
	//}()

	/*
	That's most critical part of system,
	no matter what - serving pages must proceed fine.
	If this process cause any error - we have panic and recover procedure
	 */
	// TODO: handle Panic by recover
	if err := serve(cfg.Main, cacher, fsHandler); err != nil {
		log.Panic(err)
	}
}

func serve(config *config.MainConfig, cacher cache.Cacher, fsHandler http.Handler) error {
	http.Handle("/", webserver.PickHandler(cacher, fsHandler))
	return http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil)
}
