package broker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/models"
)

const (
	ENQUEUE_SLEEP_TIMEOUT time.Duration = 30 * time.Second
)

var (
	prefixLen = len(models.PREFIX_ENQUEUE) // (len(PREFIX_ENQUEUE) - 1) + 1 (for semicolon)
)

func Enqueue(cacher *cache.Cacher, conn *websocket.Conn) error {
	for {
		urls, err := getURLs(*cacher, 5)
		if err != nil {
			log.Print(err)
		}
		for _, url := range urls {
			// remove PREFIX_ENQUEUE
			url = url[prefixLen:]
			if err := enqueueUrl(url, conn); err != nil {
				return err
			}
		}
		time.Sleep(ENQUEUE_SLEEP_TIMEOUT)
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

func enqueueUrl(url string, conn *websocket.Conn) error {
	msg := models.WSMessage{
		Type:    models.TypeRequestEnqueueURL,
		Message: url,
	}
	b, err := json.Marshal(msg)
	if err != nil {
		log.Println("enqueueUrl: json.Marshal:", err)
		return err
	}
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		log.Println("enqueueUrl: write:", err)
		return err
	}
	return nil
}
