package models

import (
	"sync"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	conn *websocket.Conn
	mtx  *sync.Mutex
}

// Creates new instance of websockets wrapper
func NewWSConn(c *websocket.Conn) *WSConn {
	return &WSConn{
		conn: c,
		mtx:  &sync.Mutex{},
	}
}

// Allows to send messages in goroutines and avoid race condition and errors
func (w *WSConn) WriteMessage(messageType int, data []byte) error {
	w.mtx.Lock()
	err := w.conn.WriteMessage(messageType, data)
	w.mtx.Unlock()
	return err
}

// Just proxy of function
func (w *WSConn) ReadMessage() (messageType int, p []byte, err error) {
	return w.conn.ReadMessage()
}

// Just proxy of function
func (w *WSConn) Ping() () {
	return
}

// Close connection
func (w *WSConn) Close() {
	w.conn.Close()
}
