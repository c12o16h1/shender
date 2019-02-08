package broker

import (
	"encoding/json"
	"log"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
)

const (
	ERR_INVALID_URL_MESSAGE = models.Error("Invalid message or token for add URL to crawl")
)

func Listen(conn *websocket.Conn, jobs chan<- models.Job, sleeperRequestGetUrls chan<- int64, sleeperResponseCachedPage chan<- int64, sleeperTypeRequestSendURL chan<- int64) error {
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
			if len(m.Message) > 0 && len(m.Token) > 0 {
				j := models.Job{
					Token: m.Token,
					Url:   m.Message,
				}
				// Add to channel
				jobs <- j
			} else {
				log.Print(ERR_INVALID_URL_MESSAGE)
			}
		case models.TypeError:
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
			}
			// Sleep if channel is free

		}
		log.Printf("recv: %s", message)
	}
}
