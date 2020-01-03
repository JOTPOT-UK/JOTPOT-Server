package http1

import (
	"errors"
	"strings"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1/httpchars"

	"github.com/JOTPOT-UK/JOTPOT-Server/mediatype"
)

func FormatMediaType(mt *mediatype.Type) ([]byte, error) {
	out := make([]byte, len(mt.Type)+len(mt.Subtype)+1)
	i := copy(out, mt.Type)
	out[i] = '/'
	i++
	i += copy(out[i:], mt.Subtype)
	for _, p := range mt.Params {
		value, err := FormatTokenOrQuotedString([]byte(p.Value)) //TODO: Do we have to convert to []byte?
		if err != nil {
			return nil, err
		}
		out = append(
			append(
				append(
					append(out, ';'),
					p.Name...),
				'='),
			value...)
	}
	return out, nil
}

var ErrMalformedContentType = errors.New("malformed content-type")

func ParseMediaType(str string) (*mediatype.Type, error) {
	mt := new(mediatype.Type)
	i := strings.IndexByte(str, '/')
	if i == -1 {
		mt.Type = str
		return mt, ErrMalformedContentType
	}
	mt.Type = strings.ToLower(str[:i])
	str = str[i+1:]
	i = strings.IndexByte(str, ';')
	if i == -1 {
		mt.Subtype = str
		return mt, nil
	}
	mt.Subtype = strings.ToLower(httpchars.RemoveWhitespacePostfix(str[:i]))
	str = httpchars.RemoveWhitespacePrefix(str[i+1:])
	var err error
	for {
		i = strings.IndexByte(str, '=')
		if i == -1 {
			return mt, ErrMalformedContentType
		}
		p := mediatype.Param{
			Name: strings.ToLower(str[:i]),
		}
		str = str[i+1:]
		if str[0] == '"' {
			p.Value, i, err = ParseQuotedStringNoStartingQuote([]byte(str[1:]))
			if err != nil {
				return mt, err
			}
			mt.Params = append(mt.Params, p)
			str = httpchars.RemoveWhitespacePrefix(str[i+1:])
			if len(str) == 0 {
				break
			} else if str[0] == ';' {
				str = httpchars.RemoveWhitespacePrefix(str[1:])
				continue
			} else {
				return mt, ErrMalformedContentType
			}
		} else {
			i = strings.IndexByte(str, ';')
			if i == -1 {
				p.Value = str
				mt.Params = append(mt.Params, p)
				break
			}
			p.Value = httpchars.RemoveWhitespacePostfix(str[:i])
			mt.Params = append(mt.Params, p)
			str = httpchars.RemoveWhitespacePrefix(str[i+1:])
		}
	}
	return mt, nil
}
