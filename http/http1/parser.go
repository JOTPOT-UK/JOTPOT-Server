package http1

import (
	"bufio"
	"io"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/jpserror"
)

func SendError(err jpserror.HTTPError, writer io.Writer) {
	writer.Write([]byte("HTTP/1.1 " + strconv.FormatUint(uint64(err.Code), 10) + " " + http.StatusText(err.Code) + "\r\nConnection: close\r\nContent-Length: " + strconv.FormatUint(uint64(len(err.Message)), 10) + "\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" + err.Message))
}

func Parse(reader *bufio.Reader) (*jps.Request, bool, error) {
	tp := textproto.NewReader(reader)
	//First line, ie "METHOD URI HTTP/1.x"
	reqLine, err := tp.ReadLine()
	if err != nil {
		return nil, false, err
	}
	//TODO: Defer?
	//Start and end indexs of the URI section
	URIStart := strings.Index(reqLine, " ") + 1
	URIEnd := strings.Index(reqLine[URIStart:], " ")
	//TODO: Defer?
	//If URIEnd is -1, the URI doesn't end. And if URIStart is 0, then the URI never starts, but if that is the case, then URIEnd will be -1, so we only need to check that.
	if URIEnd == -1 {
		return nil, false, jpserror.ErrMalformedHTTPRequest
	}
	URIEnd += URIStart
	//Create the request object, and set the method and HTTP version.
	req := new(jps.Request)
	req.Method = reqLine[:URIStart-1]
	var ok bool
	req.HTTPVersion, ok = http.ParseHTTPVersion(reqLine[URIEnd+1:])
	if !ok {
		return nil, false, jpserror.ErrMalformedHTTPVersion
	}

	//Parse URI
	if req.Method == "CONNECT" && reqLine[URIStart] != '/' {
		//If it is a CONNECT request, and the URL doesn't start with a /,
		// then it shouldn't contain the scheme. Therefore we should add "http://"
		//  to the URI to parse, and then remove it afterwards.
		req.URL, err = url.ParseRequestURI("http://" + reqLine[URIStart:URIEnd])
		req.URL.Scheme = ""
	} else {
		req.URL, err = url.ParseRequestURI(reqLine[URIStart:URIEnd])
	}
	if err != nil {
		return nil, false, jpserror.ErrMalformedURI
	}

	//Next up, the headers! This is easy!
	tpHeader, err := tp.ReadMIMEHeader()

	//Make sure header names have a length
	for k := range tpHeader {
		if len(k) == 0 {
			ok = false
			break
		}
	}
	if err != nil || !ok {
		return nil, false, jpserror.ErrMalformedHeaders
	}
	req.Header = header.Header(tpHeader)

	hosts := req.Header.GetValues("Host")
	if len(hosts) != 1 {
		return nil, false, jpserror.ErrMustBe1HostHeader
	}
	//Make sure the host is set in the URL,
	// if the host wasn't in the URI line, then use the Host header field.
	if req.URL.Host == "" {
		req.URL.Host = hosts[0]
	} else if req.URL.Host != hosts[0] {
		return nil, false, &jpserror.HTTPError{Code: 400, Message: "Host header must be identicle to host in request line"}
	}

	//Determine if this is an upgrade request, if so, then close, and there is no content length. Otherwuse, we should close the connection if it is HTTP/0.*
	if req.Method == "PRI" && len(tpHeader) == 0 && req.URL.Path == "*" && req.HTTPVersion.Major == 2 && req.HTTPVersion.Minor == 0 {
		req.Close = true
		return req, true, nil
	} else if req.HTTPVersion.Major < 1 {
		req.Close = true
	} else {
		//Else, we should close the connection if there is a header to tell us to
		req.Close = req.Header.ContainsTokenRaw("Connection", "close")
		if req.HTTPVersion.Major == 1 && req.HTTPVersion.Minor == 0 {
			//But in HTTP/1.0, the default is to close, so if there is no keep-alive, then close.
			req.Close = req.Close || !req.Header.ContainsTokenRaw("Connection", "keep-alive")
		}
	}
	return req, false, nil
}

func ParsePipes(req *jps.Request) *jpserror.HTTPError {
	codes := req.Header.GetValuesRawKey("Transfer-Encoding")
	var chunked bool
	for i := range codes {
		if codes[i] == "chunked" {
			if chunked {
				return jpserror.ErrCannotChunkAnAlreadyChunkedBody
			}
			chunked = true
		}
	}
	return nil
}
