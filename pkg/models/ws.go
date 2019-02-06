package models

import (
	"encoding/json"
)

type WSMessage struct {
	Type    WSType `json:"type"`
	Token   string `json:"token"`
	Message string `json:"message"`
}

type WSType uint

const (
	TypeRequestEnqueueURL  WSType = 11 // Message to send URL to server to enqueue for crawling by 3-rd party crawler
	TypeRequestGetUrls     WSType = 12 // Message to server to get URLs for crawl
	TypeResponseGetUrls    WSType = 13 // Message from server with url to crawl
	TypeRequestCrawl       WSType = 14 // Message to server with result of crawling some URL
	TypeResponseCachedPage WSType = 15 // Message from server with content of cached page
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
