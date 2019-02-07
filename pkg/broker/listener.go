package broker

import (
	"encoding/json"
	"log"
	"strconv"

	"github.com/c12o16h1/shender/pkg/models"
	"github.com/gorilla/websocket"
)

const (
	ERR_INVALID_URL_MESSAGE = models.Error("Invalid message or token for add URL to crawl")
)

func Listen(conn *websocket.Conn, jobs chan models.Job, sleeperCh chan int64) error {
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
		case models.TypeSleeperGetUrls:
			// Sleep if channel is free
			if len(sleeperCh) < cap(sleeperCh){
				i, err := strconv.ParseInt(m.Message, 10, 64)
				if err != nil {
					log.Print(err)
					continue
				}
				sleeperCh <- i
			}
		}
		log.Printf("recv: %s", message)
	}
}
