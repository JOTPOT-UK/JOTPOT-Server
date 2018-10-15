package jps

import (
	"net/url"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
)

type Request struct {
	Method      string
	URL         *url.URL
	HTTPVersion http.HTTPVersion
	Header      header.Header
	Close       bool
}
