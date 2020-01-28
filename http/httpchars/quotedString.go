package httpchars

import (
	"errors"
)

var ErrNoStartingQuotesForQuotedString = errors.New("no starting quotes for quoted string")
var ErrNoEndingQuotesForQuotedString = errors.New("no ending quotes for quoted string")
var ErrMalformedQuotedString = errors.New("malformed quoted string")
var ErrInvalidCharacter = errors.New("invalid character")

func ParseQuotedString(qstr string) (string, int, error) {
	if len(qstr) == 0 || qstr[0] != '"' {
		return "", 0, ErrNoStartingQuotesForQuotedString
	}
	return ParseQuotedStringNoStartingQuote([]byte(qstr[1:])) //TODO: []byte or string?
}

func ParseQuotedStringNoStartingQuote(qstr []byte) (string, int, error) {
	str := make([]byte, 0, len(qstr))
	for i := 0; i < len(qstr); i++ {
		c := qstr[i]
		if !IsQDText(c) {
			if c == '\\' {
				i++
				c = qstr[i]
				if !IsQuotedPairChar(c) {
					return string(str), i, ErrMalformedQuotedString
				}
			} else if c == '"' {
				return string(str), i, nil
			} else {
				return string(str), i, ErrMalformedQuotedString
			}
		}
		str = append(str, c)
	}
	return string(str), len(qstr), ErrNoEndingQuotesForQuotedString
}

func FormatTokenOrQuotedString(str []byte) ([]byte, error) {
	isQuoted := false
	out := make([]byte, 0, len(str))
	for _, c := range str {
		if IsTokenChar(c) {
			out = append(out, c)
		} else if IsQDText(c) {
			if isQuoted {
				out = append(out, c)
			} else {
				ns := make([]byte, len(out)+2, len(str)+2)
				ns[0] = '"'
				n := copy(ns[1:], out) + 1
				ns[n] = c
				out = ns
				isQuoted = true
			}
		} else if IsQuotedPairChar(c) {
			if isQuoted {
				out = append(out, '\\', c)
			} else {
				ns := make([]byte, len(out)+3, len(str)+4)
				ns[0] = '"'
				n := copy(ns[1:], out) + 1
				ns[n] = '\\'
				ns[n+1] = c
				out = ns
				isQuoted = true
			}
		} else {
			return out, ErrInvalidCharacter
		}
	}
	if isQuoted {
		out = append(out, '"')
	}
	return out, nil
}
