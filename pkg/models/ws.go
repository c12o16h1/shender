package models

type WSMessage struct {
	Type    WSType `json:"type"`
	Token   string `json:"token"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type WSType uint

const (
	TypeRequestSendURL     WSType = 11  // Message to send URL to server to enqueue for crawling by 3-rd party crawler
	TypeRequestGetUrls     WSType = 20  // Message to server to get URLs for crawl
	TypeResponseGetUrls    WSType = 21  // Message from server with url to crawl
	TypeSleeperGetUrls     WSType = 22  // Message from server with timeout to sleep before next request. Seconds.
	TypeRequestCrawl       WSType = 30  // Message to server with result of crawling some URL
	TypeResponseCachedPage WSType = 40  // Message from server with content of cached page
	TypeError              WSType = 100 // Error
	TypeOk                 WSType = 101 // Ok
)
