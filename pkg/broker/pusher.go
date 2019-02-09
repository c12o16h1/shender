package broker

import (
	"encoding/json"
	"time"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

/*
Pushes crawled page cache to server
 */
func Push(conn *models.WSConn, chRes <-chan models.JobResult, sleeperCh <-chan int64, sleepTime *time.Duration) error {
	for {
		select {
		case <-sleeperCh:
			// Sleep
			time.Sleep(*sleepTime)
		default:
			res := <-chRes

			data := models.DataResponseCachedPage{
				URL:  res.Url,
				HTML: res.HTML,
			}
			dBytes, err := json.Marshal(data)
			if err != nil {
				return errors.Wrap(err, "Push: json.Marshal:")
			}
			msg := models.WSMessage{
				Type:    models.TypeResponseCachedPage,
				Message: res.Url,        // URL of crawled page
				AppID:   res.AppID,      // Job token
				Data:    string(dBytes), // Bytes of custom payload
			}
			b, err := json.Marshal(msg)
			if err != nil {
				return errors.Wrap(err, "Push: json.Marshal:")
			}
			err = conn.WriteMessage(websocket.BinaryMessage, b)
			if err != nil {
				return errors.Wrap(err, "Push: write:")
			}
		}
	}
}
