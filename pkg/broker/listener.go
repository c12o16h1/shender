package broker

import (
	"encoding/json"
	"log"

	"github.com/c12o16h1/shender/pkg/models"
)

const (
	ERR_INVALID_URL_MESSAGE = models.Error("Invalid message or token for add URL to crawl")
	ERR_INVALID_CACHE       = models.Error("Invalid cache content")
)

/*
Listener listen all messages from server
 */
func Listen(
	conn *models.WSConn,
	jobsCh chan<- models.Job,
	storagerCh chan<- models.DataResponseCachedPage,
	sleeperRequestGetUrls chan<- int64,
	sleeperResponseCachedPage chan<- int64,
	sleeperTypeRequestSendURL chan<- int64,
	sleeperRequestCachedPage chan<- int64,
) error {
	for {
		// Listen and read
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return err
		}
		// Unmarshal
		var m models.WSMessage
		if err := json.Unmarshal(message, &m); err != nil {
			return err
		}

		switch m.Type {
		case models.TypeResponseGetUrls:
			// Got URL to crawl
			if len(m.Data) > 0 {
				var urlRich models.URLRich
				if err := json.Unmarshal([]byte(m.Data), &urlRich); err != nil {
					log.Print(ERR_INVALID_URL_MESSAGE)
					continue
				}
				j := models.Job{
					Token: m.Token,
					AppID: urlRich.AppID,
					Url:   urlRich.Url,
				}
				// Add to channel
				jobsCh <- j
			} else {
				log.Print(ERR_INVALID_URL_MESSAGE)
			}

		case models.TypeResponseCachedPage:
			// Got cache to store
			var c models.DataResponseCachedPage
			if err := json.Unmarshal([]byte(m.Data), &c); err != nil {
				log.Print(ERR_INVALID_CACHE)
				continue
			}
			storagerCh <- c

		case models.TypeError:
			// Handle errors
			switch m.Code {
			case models.CodeRequestGetUrls:
				if len(sleeperRequestGetUrls) < cap(sleeperRequestGetUrls) {
					sleeperRequestGetUrls <- 0
				}
			case models.CodeResponseCachedPage:
				if len(sleeperResponseCachedPage) < cap(sleeperResponseCachedPage) {
					sleeperResponseCachedPage <- 0
				}
			case models.CodeRequestSendURL:
				if len(sleeperTypeRequestSendURL) < cap(sleeperTypeRequestSendURL) {
					sleeperTypeRequestSendURL <- 0
				}
			case models.CodeRequestCachedPage:
				if len(sleeperRequestCachedPage) < cap(sleeperRequestCachedPage) {
					sleeperRequestCachedPage <- 0
				}
			}

		}
		log.Printf("recv: %s", message)
	}
}
