package broker

import (
	"encoding/json"
	"log"
	"time"

	"github.com/c12o16h1/shender/pkg/cache"
	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	ENQUEUE_SLEEP_TIMEOUT time.Duration = 30 * time.Second
)

var (
	prefixLen = len(models.PREFIX_ENQUEUE) // (len(PREFIX_ENQUEUE) - 1) + 1 (for semicolon)
)
/*
Enqueuer sends app URL to server to enqueue to be crawled
 */
func Enqueue(cacher *cache.Cacher, conn *models.WSConn, appID string, sleeperCh <-chan time.Duration) error {
	// Enqueue our URL to push into server
	for {
		select {
		case sleepTime := <-sleeperCh:
			// Sleep
			time.Sleep(sleepTime)
		default:
			urls, err := getURLs(*cacher, 5)
			if err != nil {
				log.Print(err)
			}
			for _, url := range urls {
				// remove PREFIX_ENQUEUE
				url = url[prefixLen:]
				if err := enqueueUrl(url, appID, conn); err != nil {
					return err
				}
			}
			time.Sleep(ENQUEUE_SLEEP_TIMEOUT)
		}

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

func enqueueUrl(url string, appID string, conn *models.WSConn) error {
	urlRich := models.URLRich{
		Url:   url,
		AppID: appID,
	}
	bUrl, err := json.Marshal(urlRich)
	if err != nil {
		return errors.Wrap(err, "enqueueUrl: json.Marshal:")
	}

	msg := models.WSMessage{
		Type: models.TypeRequestSendURL,
		Data: string(bUrl),
	}
	b, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "enqueueUrl: json.Marshal:")
	}
	err = conn.WriteMessage(websocket.BinaryMessage, b)
	if err != nil {
		return errors.Wrap(err, "enqueueUrl: write:")
	}
	return nil
}
