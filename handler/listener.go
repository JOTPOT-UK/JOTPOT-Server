package handler

import (
	"net"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/server"
)

//ListenOn listens on l, with server s, and for every connection, Handler is called in a new goroutine.
func ListenOn(s *server.Server, l net.Listener) error {
	var con net.Conn
	var err error
	var ne net.Error
	var ok, delayed bool
	retryDelay := 2 * time.Millisecond
	for {
		con, err = l.Accept()
		if err != nil {
			if ne, ok = err.(net.Error); ok && ne.Temporary() {
				delayed = true
				time.Sleep(retryDelay)
				if retryDelay < 128 {
					retryDelay *= 2
				}
				continue
			}
			return err
		} else if delayed {
			delayed = false
			retryDelay = 2 * time.Millisecond
		}
		go Handle(s, con)
	}
}
