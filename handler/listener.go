package handler

import (
	"jotpot/net/jps/server"
	"net"
	"time"
)

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
