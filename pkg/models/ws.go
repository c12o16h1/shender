package models

import (
	"encoding/json"
)

type WSMessage struct {
	Type    WSType `json:"type"`
	Message string `json:"message"`
}

type WSType uint

const (
	TypeEnqueueURL    WSType = 11 //
	TypeGetToCrawl    WSType = 12
	TypeCrawlResponse WSType = 13
	TypeCachedPage    WSType = 14
)

func JSONEncode(smth interface{}) ([]byte, error) {
	b, err := json.Marshal(smth)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func JSONDecode(byt []byte) (interface{}, error) {
	var msg WSMessage
	err := json.Unmarshal(byt, &msg)
	if err != nil {
		return nil, err
	}
	return byt, nil
}
