package http1

import (
	"bufio"
	"io"
	"jotpot/net/http"
	"jotpot/net/http/header"
	"jotpot/net/jps"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
)

func SendError(err jps.HTTPError, writer io.Writer) {
	writer.Write([]byte("HTTP/1.1 " + strconv.FormatUint(uint64(err.Code), 10) + " " + http.StatusText(err.Code) + "\r\nConnection: close\r\nContent-Length: " + strconv.FormatUint(uint64(len(err.Message)), 10) + "\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n" + err.Message))
}

func Parse(reader *bufio.Reader) (*jps.Request, bool, *jps.HTTPError, error) {
	tp := textproto.NewReader(reader)
	//First line, ie "METHOD URI HTTP/1.x"
	reqLine, err := tp.ReadLine()
	if err != nil {
		return nil, false, nil, err
	}
	//TODO: Defer?
	//Start and end indexs of the URI section
	URIStart := strings.Index(reqLine, " ") + 1
	URIEnd := strings.Index(reqLine[URIStart:], " ")
	//If URIEnd is -1, the URI doesn't end. And if URIStart is 0, then the URI never starts, but if that is the case, then URIEnd will be -1, so we only need to check that.
	if URIEnd == -1 {
		return nil, false, &jps.HTTPError{400, "Malformed HTTP request"}, nil
	}
	URIEnd += URIStart
	//Create the request object, and set the method and HTTP version.
	req := new(jps.Request)
	req.Method = reqLine[:URIStart-1]
	var ok bool
	req.HTTPVersion, ok = http.ParseHTTPVersion(reqLine[URIEnd+1:])
	if !ok {
		return nil, false, &jps.HTTPError{400, "Malformed HTTP version"}, nil
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
		//TODO: Scheme?
	}
	if err != nil {
		return nil, false, &jps.HTTPError{400, "Malformed URI"}, nil
	}

	//Next up, the headers! This is easy!
	tpHeader, err := tp.ReadMIMEHeader()

	//Make sure header names have a length
	for k, _ := range tpHeader {
		if len(k) == 0 {
			ok = false
			break
		}
	}
	if err != nil || !ok {
		return nil, false, &jps.HTTPError{400, "Malformed headers"}, err
	}
	req.Header = header.Header(tpHeader)

	//Make sure the host is set in the URL,
	// if the host wasn't in the URI line, then use the Host header field.
	if req.URL.Host == "" {
		req.URL.Host = req.Header.Get("Host")
	}

	//Determine if this is an upgrade request, if so, then close, and there is no content length. Otherwuse, we should close the connection if it is HTTP/0.*
	if req.Method == "PRI" && len(tpHeader) == 0 && req.URL.Path == "*" && req.HTTPVersion.Major == 2 && req.HTTPVersion.Minor == 0 {
		req.Close = true
		return req, true, nil, nil
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
	return req, false, nil, nil
}

func ParsePipes(req *jps.Request) {
	codes := req.Header.GetValuesRawKey("Transfer-Encoding")
	var chunked bool
	for i := range codes {
		if codes[i] == "chunked" {
			if chunked {

			}
			chunked = true
		}
	}
}
