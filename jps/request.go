package jps

import (
	"jotpot/net/http"
	"jotpot/net/http/header"
	"net/url"
)

type Request struct {
	Method      string
	URL         *url.URL
	HTTPVersion http.HTTPVersion
	Header      header.Header
	Close       bool
}
