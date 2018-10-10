package http1

import (
	"bufio"
	"fmt"
	"jotpot/net/jps"
	"jotpot/net/jps/server"
	"net"
)

func numLeadingCRorLF(v []byte) (n int) {
	for i := range v {
		if v[i] != '\r' && v[i] != '\n' {
			return
		}
		n++
	}
	return
}

func Handle(s *server.Server, con net.Conn) {
	reader := bufio.NewReader(con)
	var httpErr *jps.HTTPError
	var err error
	var peek []byte
	//var h2Upgrade bool
	for {
		for {
			peek, err = reader.Peek(2)
			if peek[0] == '\r' && peek[1] == '\n' {
				reader.Discard(2)
			} else {
				break
			}
		}
		req := server.NewIncomingRequest(s, con, reader)
		if err != nil {
			req.Request, _, httpErr, err = Parse(reader)
		}
		if httpErr != nil || err != nil {
			fmt.Println("Err:")
			fmt.Println(httpErr)
			fmt.Println(err)
			return
		} else {
			s.Handlers.Call(req)
		}
		if req.Request.Close {
			body, _, _, _ := req.GetBody()
			body.Close()
			return
		}
		//TODO: Pragma
	}
}
