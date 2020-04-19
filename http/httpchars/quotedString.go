package httpchars

import (
	"errors"
)

var ErrNoStartingQuotesForQuotedString = errors.New("no starting quotes for quoted string")
var ErrNoEndingQuotesForQuotedString = errors.New("no ending quotes for quoted string")
var ErrMalformedQuotedString = errors.New("malformed quoted string")
var ErrInvalidCharacter = errors.New("invalid character")

func ParseQuotedString(qstr string) ([]byte, int, error) {
	if len(qstr) == 0 || qstr[0] != '"' {
		return nil, 0, ErrNoStartingQuotesForQuotedString
	}
	return ParseQuotedStringNoStartingQuote(qstr[1:])
}

//ParseQuotedStringNoStartingQuote parses a quoted string from the character after the opening quote.
//It returns the parsed string, the index after the end of the quoted string (so after the final ")
//or the index of the character that caused an error
//and a possible error - which will either be nil or ErrMalformedQuotedString.
func ParseQuotedStringNoStartingQuote(qstr string) ([]byte, int, error) {
	str := make([]byte, 0, len(qstr)+1)
	for i := 0; i < len(qstr); i++ {
		c := qstr[i]
		if !IsQDText(c) {
			if c == '\\' {
				i++
				c = qstr[i]
				if !IsQuotedPairChar(c) {
					return str, i, ErrMalformedQuotedString
				}
			} else if c == '"' {
				return str, i + 1, nil
			} else {
				return str, i, ErrMalformedQuotedString
			}
		}
		str = append(str, c)
	}
	return str, len(qstr), ErrNoEndingQuotesForQuotedString
}

func ParseFullQuotedString(qstr string) ([]byte, error) {
	if len(qstr) == 0 || qstr[0] != '"' {
		return nil, ErrNoStartingQuotesForQuotedString
	}
	return ParseFullQuotedStringNoStartingQuote(qstr[1:])
}

func ParseFullQuotedStringNoStartingQuote(qstr string) ([]byte, error) {
	str, end, err := ParseQuotedStringNoStartingQuote(qstr)
	if end != len(qstr) {
		err = ErrMalformedQuotedString
	}
	return str, err
}

func FormatTokenOrQuotedString(str string) ([]byte, error) {
	isQuoted := false
	out := make([]byte, 1, len(str)+4)
	//The first character will be quotes if it's quoted.

	var err error
	for i := range str {
		c := str[i]
		if IsTokenChar(c) {
			out = append(out, c)
		} else if IsQDText(c) {
			out = append(out, c)
			if !isQuoted {
				isQuoted = true
			}
		} else if IsQuotedPairChar(c) {
			out = append(out, '\\', c)
			if !isQuoted {
				isQuoted = true
			}
		} else {
			err = ErrInvalidCharacter
			break
		}
	}

	if isQuoted {
		out[0] = '"'
		out = append(out, '"')
		return out, err
	}
	//Since we're not quoted, we don't want to include the first starting character.
	return out[1:], err
}
