package handler

import (
	"crypto/tls"
	"net"

	"jotpot/net/http/http1"
	"jotpot/net/jps/server"
)

func Handle(s *server.Server, con net.Conn) {
	defer con.Close()
	var err error
	if tlsCon, ok := con.(*tls.Conn); ok {
		if err = tlsCon.Handshake(); err != nil {
			return
		}
		tlsState := tlsCon.ConnectionState()
		if tlsState.NegotiatedProtocol != "" && tlsState.NegotiatedProtocol != "http/1.0" && tlsState.NegotiatedProtocol != "http/1.1" {
			if tlsState.NegotiatedProtocol == "h2" {
				//HandleHTTP2(con)
			}
			return
		}
	}
	http1.Handle(s, con)
}
