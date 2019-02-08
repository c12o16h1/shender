package broker

import (
	"encoding/json"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)
/*
Pushes crawled page cache to server
 */
func Push(chRes <-chan models.JobResult, conn *websocket.Conn) error {
	for {
		res := <-chRes
		msg := models.WSMessage{
			Type:    models.TypeResponseCachedPage,
			Message: res.HTML,
			Token:   res.Token,
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
