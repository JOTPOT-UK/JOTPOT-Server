package http

import (
	"strconv"
	"strings"
)

type HTTPVersion struct {
	Major, Minor uint16
}

func ParseHTTPVersion(s string) (HTTPVersion, bool) {
	if !strings.HasPrefix(s, "HTTP/") {
		return HTTPVersion{0, 0}, false
	}
	s = s[5:]
	switch s {
	case "1.1":
		return HTTPVersion{1, 1}, true
	case "1.0":
		return HTTPVersion{1, 0}, true
	}
	dot := strings.Index(s, ".")
	if dot < 0 {
		return HTTPVersion{0, 0}, false
	}
	major, err := strconv.ParseUint(s[:dot], 10, 16)
	if err != nil {
		return HTTPVersion{0, 0}, false
	}
	minor, err := strconv.ParseUint(s[dot+1:], 10, 16)
	if err != nil {
		return HTTPVersion{uint16(major), 0}, false
	}
	return HTTPVersion{uint16(major), uint16(minor)}, true
}
