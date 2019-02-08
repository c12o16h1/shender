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
	sleeper      = 30 * time.Second
)

func main() {
	// Initialization
	cfg := config.New()
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
	sleeperRequestGetUrls := make(chan int64, 1)
	// Chan to pause request to push crawled content
	sleeperResponseCachedPage := make(chan int64, 1)
	// Chan to pause request to enqueue urls to be crawled
	sleeperTypeRequestSendURL := make(chan int64, 1)
	// Chan to pause request to get cached pages from server
	sleeperRequestCachedPage := make(chan int64, 1)

	/*
	Establishing WS connection to main server
	 */
	u := url.URL{Scheme: "ws", Host: cfg.Main.WSHost, Path: ""}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer conn.Close()

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
	go func(wsc *websocket.Conn, incoming chan<- models.Job, sleeperCh chan<- int64, ) {
		for {
			if err := broker.Listen(
				conn,
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
	}(conn, incomingQueue, sleeperRequestGetUrls)

	// Other Apps pages crawling
	/*
	Spawn goroutine to get URLs for crawling from server
	so they'll be crawled by this app.
	*/
	go func(wsc *websocket.Conn, sleeperCh <-chan int64) {
		for {
			if err := broker.Request(conn, incomingQueue, sleeperCh, &sleeper); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}(conn, sleeperRequestGetUrls)

	/*
	Spawn goroutine to process crawling of pages for other members of system.
	This goroutine ensure that server has enough resources to do render,
	spawn new renderer instance, connect to them via RPC, and do render for URL from incoming queue.
	Then save push result to outgoing queue
	 */
	go func(incoming <-chan models.Job, outgoing chan<- models.JobResult) {
		for {
			if err := broker.Crawl(incomingQueue, outgoingQueue); err != nil {
				log.Print("Crawl: ", err)
				time.Sleep(shortSleeper)
			}
		}
	}(incomingQueue, outgoingQueue)

	/*
	Spawn goroutine to push content of crawled pages to server
	*/
	go func(wsc *websocket.Conn, resChan <-chan models.JobResult, sleeperCh <-chan int64) {
		for {
			if err := broker.Push(conn, resChan, sleeperCh, &sleeper); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}(conn, outgoingQueue, sleeperResponseCachedPage)

	// This App cache
	/*
	Spawn goroutine to send/enqueue URL to central server
	so they'll be crawled by other members.
	*/
	go func(c *cache.Cacher, wsc *websocket.Conn, sleeperCh <-chan int64) {
		for {
			if err := broker.Enqueue(&cacher, conn, sleeperCh, &sleeper); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}(&cacher, conn, sleeperTypeRequestSendURL)

	/*
	Spawn goroutine to get cached pages from central server
	so bots may see cached pages content
	*/
	go func(wsc *websocket.Conn, sleeperCh <-chan int64) {
		for {
			if err := broker.RequestCache(conn, sleeperCh, &sleeper); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}(conn, sleeperRequestCachedPage)

	/*
	Spawn goroutine to save cache in local cache DB
	so bots may see cached pages content
	*/
	go func(c *cache.Cacher, sleeperCh <-chan int64) {
		for {
			if err := broker.Storage(c, storagerQueue, sleeperRequestCachedPage); err != nil {
				log.Print(err)
				time.Sleep(shortSleeper)
			}
		}
	}(&cacher, storagerQueue, sleeperRequestCachedPage)

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
