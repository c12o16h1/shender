package broker

import (
	"encoding/json"
	"time"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

const (
	RECEIVE_SLEEP_TIMEOUT time.Duration = 100 * time.Millisecond
)

/*
Requests new URLS to crawl
 */
func Request(conn *websocket.Conn, jobsCh chan models.Job, sleeperChan <-chan int64, sleepTime *time.Duration) error {
	jobsEmptyTrigger := cap(jobsCh) / 2
	// Request new urls to crawl
	for {
		select {
		case <-sleeperChan:
			// Sleep
			time.Sleep(*sleepTime)
		default:
			// If we have not enough URL to crawl
			if len(jobsCh) < jobsEmptyTrigger {
				msg := models.WSMessage{
					Type: models.TypeRequestGetUrls,
				}
				b, err := json.Marshal(msg)
				if err != nil {
					return errors.Wrap(err, "enqueueUrl: json.Marshal:")
				}
				err = conn.WriteMessage(websocket.BinaryMessage, b)
				if err != nil {
					return errors.Wrap(err, "enqueueUrl: write:")
				}
			}
		}
		time.Sleep(RECEIVE_SLEEP_TIMEOUT)
	}
}
