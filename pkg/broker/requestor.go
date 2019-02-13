package broker

import (
	"encoding/json"
	"time"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

/*
Requests new URLS to crawl
 */
func Request(conn *models.WSConn, jobsCh chan models.Job, sleeperChan <-chan time.Duration) error {
	jobsEmptyTrigger := cap(jobsCh) / 2
	// Request new urls to crawl
	for {
		select {
		case sleepTime := <-sleeperChan:
			// Sleep
			time.Sleep(sleepTime)
		default:
			// If we have not enough URL to crawl
			if len(jobsCh) < jobsEmptyTrigger {
				msg := models.WSMessage{
					Type: models.TypeRequestGetUrls,
				}
				b, err := json.Marshal(msg)
				if err != nil {
					return errors.Wrap(err, "Request: json.Marshal:")
				}
				err = conn.WriteMessage(websocket.BinaryMessage, b)
				if err != nil {
					return errors.Wrap(err, "Request: write:")
				}
			}
		}
		time.Sleep(WS_BUMP_TIMEOUT)
	}
}
