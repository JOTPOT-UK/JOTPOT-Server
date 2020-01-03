package tls

import (
	"crypto/tls"
	"net"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

func ProtocolNegotiator(con net.Conn, p string, _ jps.BufioReaderGetter) (string, error) {
	if tlsCon, ok := con.(*tls.Conn); ok {
		//Ensure we have made the handshake
		if err := tlsCon.Handshake(); err != nil {
			return p, err
		}
		//If TLS negotiated a protocol, then returns it...
		tlsState := tlsCon.ConnectionState()
		if tlsState.NegotiatedProtocol != "" {
			return tlsState.NegotiatedProtocol, nil
		}
	}
	//By default, just pass the protocol on
	return p, nil
}
