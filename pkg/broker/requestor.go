package broker

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"

	"github.com/c12o16h1/shender/pkg/models"
)

const (
	RECEIVE_SLEEP_TIMEOUT time.Duration = 100 * time.Millisecond
)

func Request(jobsCh chan models.Job, sleeperChan <-chan int64, conn *websocket.Conn) error {
	jobsEmptyTrigger := cap(jobsCh) / 2
	// Request new urls to crawl
	for {
		select {
		case sleep := <-sleeperChan:
			// Sleep
			time.Sleep(time.Duration(sleep) * time.Second)
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
