package http

import "github.com/JOTPOT-UK/JOTPOT-Server/http/header"

import "github.com/JOTPOT-UK/JOTPOT-Server/http/httpchars"

type CacheControlDirective struct {
	Name     string
	ArgGiven bool
	Arg      string
}

func ParseCacheControl(s string) ([]CacheControlDirective, error) {
	out := make([]CacheControlDirective, 0, 1)
	for {
		d := CacheControlDirective{}
		d.Name, s = httpchars.ReadToken(s)
		if d.Name == "" {

		} else {
			if len(s) > 0 && s[0] == '=' {
				if len(s) == 1 {
					return out, ErrMalformedCacheControlHeaderExpectingArgumentAfterEquals
				}
				d.ArgGiven = true
				if s[1] == '"' {
					d.Arg, err = httpchars.ParseQuotesStringNoStartQuote(s[2:])
				} else {
					d.Arg, s = httpchars.ReadToken(s[1:])
				}
			}
			out = append(out, d)
			s = httpchars.ConsumeWhitespace(s)
			if len(s) > 0 && s[0] == ',' {
				s = httpchars.ConsumeWhitespace(s[1:])
				continue
			} else {
				break
			}
		}
	}
	return out, nil
}

type RequestCacheControl struct {
	H *header.Interface
}
