package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"

	"github.com/c12o16h1/shender/pkg/models"
)

const (
	WORKER_LIFE_TIME = 30 * time.Second

	ERR_INVALID_PORT = models.Error("Invalid port")
)

func main() {
	port := flag.Int("port", 0, "this binary port")
	flag.Parse()
	if *port == 0 {
		log.Fatal(ERR_INVALID_PORT)
	}

	w, _ := NewWorker(*port - 10000, *port - 9999)
	// Close worker anyway if after some time
	go func() {
		time.Sleep(WORKER_LIFE_TIME)
		var out string
		w.Close(0, &out)
	}()
	rpc.Register(w)
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		rpc.ServeConn(c)
	}
}
