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

/*
RequestCache requests new URLs to crawl
 */
func RequestCache(conn *models.WSConn, appID string, sleeperChan <-chan time.Duration, renewWS chan<- int) error {
	// Request new urls to crawl
	for {
		select {
		case sleepTime := <-sleeperChan:
			// Sleep
			time.Sleep(sleepTime)
		default:
			msg := models.WSMessage{
				Type:  models.TypeRequestCachedPage,
				AppID: appID,
			}
			b, err := json.Marshal(msg)
			if err != nil {
				return errors.Wrap(err, "RequestCache: json.Marshal:")
			}
			err = conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				// ASk to renew WS connection
				log.Print("RENEW WS")
				renewWS <- 0
				return errors.Wrap(err, "RequestCache: write:")
			}
		}
		time.Sleep(WS_BUMP_TIMEOUT)
	}
}

/*
Storing cache in local cache DB
 */
func Storage(c *cache.Cacher, storagerCh <-chan models.DataResponseCachedPage, sleeperChan chan<- time.Duration) error {
	for {
		ch := <-storagerCh
		if err := (*c).Set([]byte(ch.URL), []byte(ch.HTML)); err != nil {
			sleeperChan <- 0 // Pause receiving of new cache
			return errors.Wrap(err, "Storage: (*c).Set:")
		}
	}
}
