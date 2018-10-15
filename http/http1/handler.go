package http1

import (
	"bufio"
	"fmt"
	"net"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/server"
)

//Handle takes a server and a connection to that server, and handles it as a HTTP/1.x request.
func Handle(s *server.Server, con net.Conn) {
	reader := bufio.NewReader(con)
	var err error
	var peek []byte
	for {
		//Section 3.5 of RFC7230 says that:
		// a server that is expecting to receive
		// and parse a request-line SHOULD ignore at least one empty line (CRLF)
		// received prior to the request-line
		//We will ignore up to 5, as it would take a REALLY buggy client to go over that!
		for i := byte(0); i < 5; i++ {
			peek, err = reader.Peek(2)
			if peek[0] == '\r' && peek[1] == '\n' {
				reader.Discard(2)
			} else {
				break
			}
		}
		//Creare a new incoming request
		req := server.NewIncomingRequest(s, con, reader)
		req.Request, _, err = Parse(reader)
		if err != nil {
			fmt.Println("Err:")
			fmt.Println(err)
			return
		} else {
			s.Handlers.Call(req)
		}
		if req.Request.Close {
			body, _, _ := req.GetBody()
			body.Close()
			return
		}
		//TODO: Pragma
	}
}
