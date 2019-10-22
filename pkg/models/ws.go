package models

type WSMessage struct {
	Code    int    `json:"code"`    // Code, for errors
	Type    WSType `json:"type"`    // Type of message
	AppID   string `json:"app_id"`  // AppID of "owner" app
	Token   string `json:"token"`   // Unique token
	Error   string `json:"error"`   // Error message
	Message string `json:"message"` // Success message
	Data    string `json:"data"`    // Any specific payload
}

type WSType uint

const (
	TypeRequestSendURL     WSType = 11  // Message to send URL to server to enqueue for crawling by 3-rd party crawler
	TypeRequestGetUrls     WSType = 20  // Message to server to get URLs for crawl
	TypeResponseGetUrls    WSType = 21  // Message from server with url to crawl
	TypeRequestCachedPage  WSType = 30  // Message to server with result of crawling some URL
	TypeResponseCachedPage WSType = 40  // Message from server with content of cached page
	TypeError              WSType = 100 // Error
	TypeOk                 WSType = 101 // Ok

	// Error codes for requests
	CodeRequestSendURL     = 411
	CodeRequestGetUrls     = 420
	CodeResponseGetUrls    = 421
	CodeSleeperGetUrls     = 422
	CodeRequestCachedPage  = 430
	CodeResponseCachedPage = 440
)

// Custom data types

/*
Data for crawled page response
This bytes will be stored on server as []byte
in Redis List type as record for cache:app key
And will be returned as is to move to local cache
  */
type DataResponseCachedPage struct {
	URL  string `json:"url"`
	HTML string `json:"html"`
}

/*
Rich struct data to store in Redis
Contain app id and URL to crawl
 */
type URLRich struct {
	Url   string `json:"url"`    // Page url tp crawl
	AppID string `json:"app_id"` // App id of owner
}
