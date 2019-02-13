package broker

import "time"

const (
	WS_BUMP_TIMEOUT  = 1000 * time.Millisecond // Default timeout to bump server with requests
	WS_ERROR_TIMEOUT = 60 * time.Second        // Default timeout to pause requests on error from server
)
