package http1

import (
	"bufio"
	"net/url"
	"strings"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
)

func KeepAlive(req *http.Request) bool {
	return req.Version.Major >= 1 &&
		!req.Header.Has("Connection", []string{"close"}) &&
		(!(req.Version.Minor == 0 && req.Version.Major == 1) ||
			req.Header.Has("Connection", []string{"keep-alive"}))
}

func ParseRequest(reader *bufio.Reader, headerProcessor *header.Processor) (*http.Request, bool, error, bool) {
	//First line, ie "METHOD URI HTTP/1.x"
	reqLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, true, err, false
	}
	if len(reqLine) < 2 || reqLine[len(reqLine)-2] != '\r' {
		return nil, true, http.ErrMalformedHTTPRequestExpectingCarraigeReturnBeforeNewline, true
	}
	//TODO: Defer?
	//Start and end indexs of the URI section
	URIStart := strings.Index(reqLine, " ") + 1
	URIEnd := strings.Index(reqLine[URIStart:], " ")
	//TODO: Defer?
	//If URIEnd is -1, the URI doesn't end. And if URIStart is 0, then the URI never starts, but if that is the case, then URIEnd will be -1, so we only need to check that.
	if URIEnd == -1 {
		return nil, true, http.ErrMalformedHTTPRequestExpecting2SpacesOnFirstRequestLine, true
	}
	URIEnd += URIStart
	//Create the request object, and set the method and HTTP version.
	req := new(http.Request)
	req.MethodStr = reqLine[:URIStart-1]
	req.Version, err = http.ParseHTTPVersion(reqLine[URIEnd+1 : len(reqLine)-2])
	if err != nil {
		return nil, true, err, true
	}

	//Parse URI
	if req.MethodStr == "CONNECT" && reqLine[URIStart] != '/' {
		//If it is a CONNECT request, and the URL doesn't start with a /,
		// then it shouldn't contain the scheme. Therefore we should add "http://"
		//  to the URI to parse, and then remove it afterwards.
		req.URLp, err = url.ParseRequestURI("http://" + reqLine[URIStart:URIEnd])
		req.URLp.Scheme = ""
	} else {
		req.URLp, err = url.ParseRequestURI(reqLine[URIStart:URIEnd])
	}
	if err != nil {
		return nil, true, http.MakeErrMalformedURI(err), true
	}

	req.Header = header.New(12, headerProcessor)
	headerUnsafeAdder := req.Header.UnsafeAdder()
	err, canRespond := ParseHeaders(reader, &headerUnsafeAdder)
	headerUnsafeAdder.Release()
	if err != nil {
		return nil, true, http.MakeErrMalformedHeaders(err), canRespond
	}

	hosts := req.Header.GetValues("Host")
	if len(hosts) == 1 {
		//Make sure the host is set in the URL,
		// if the host wasn't in the URI line, then use the Host header field.
		if req.URLp.Host == "" {
			req.URLp.Host = hosts[0]
		} else if req.URLp.Host != hosts[0] {
			//TODO: Should we return the error below?
			//return nil, true, false, &http.Error{Code: 400, Message: "Host header must be identicle to host in request line"}
		}
	} else if req.Version.Major == 1 && req.Version.Minor > 0 || req.Version.Major > 1 {
		return nil, true, http.ErrMustBe1HostHeader, true
	}

	/*if req.MethodStr == "PRI" && len(req.Header.Values) == 0 && req.URLp.Path == "*" && req.Version.Major == 2 && req.Version.Minor == 0 {
		return req, true, nil
	}*/
	return req, !KeepAlive(req), nil, false
}

func ParsePipes(req *http.Request) *http.Error {
	codes := req.Header.GetValuesRawKey("Transfer-Encoding")
	var chunked bool
	for i := range codes {
		if codes[i] == "chunked" {
			if chunked {
				return http.ErrCannotChunkAnAlreadyChunkedBody
			}
			chunked = true
		}
	}
	return nil
}
