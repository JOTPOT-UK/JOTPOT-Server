package client

import (
	"io"
	"net"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/jpserror"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
)

type OutgoingRequest struct {
	Server     *Client
	Connection net.Conn
	Request    *jps.Request
	Pipes      []*pipe.WriterPipeGenerator

	body io.Writer
}

func (r *OutgoingRequest) GotBody() bool {
	return r.body != nil
}
func (r *OutgoingRequest) GetBody() (io.Writer, bool, error) {
	if r.body == nil {
		codes, ok := r.Server.Encodings.GetWriterPipeGenerators(r.Request.Header.GetValuesRawKey("Transfer-Encoding"))
		if !ok {
			return nil, false, jpserror.ErrUnsupportedTransferEncoding
		}
		var err error
		r.body = pipe.PipeTo(r.Connection, codes, r.Pipes)
		return r.body, true, err
	}
	return r.body, false, nil
}
