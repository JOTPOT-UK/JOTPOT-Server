package http1

import (
	"bufio"
	"errors"
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1/httpchars"
)

var ErrNoHeaderKey = errors.New("no or invalid header key")
var ErrMalformedHeader = errors.New("malformed header")
var ErrIllegalCharacterInHeaderName = errors.New("illegal character in header name")
var ErrIllegalCharacterInHeaderValue = errors.New("illegal character in header value")

//ParseHeaders parses the headers from the given reader, and calls Add on the given header.Adder for every parsed header.
func ParseHeaders(r *bufio.Reader, h header.Adder) (error, bool) {
	line, err := r.ReadBytes('\n')
	var nextLine []byte
	var nextErr error
	for {
		if len(line) == 2 {
			//Hopefully, we just have a CRLF
			if err != nil && err != io.EOF {
				return err, false
			}
			if line[0] == '\r' {
				return nil, false
			}
			return http.ErrExpectingCarraigeReturnBeforeNewline, true
		} else if len(line) < 2 {
			//This should not have happened...
			if err != nil {
				return err, true
			}
			return http.ErrExpectingCarraigeReturnBeforeNewline, false
		}
		//Now, the line should be in the format:
		//  field-name ":" OWS field-value OWS CR
		//     where field-name    = token
		//           field-value   = *( field-content / obs-fold )
		//           field-content = field-vchar [ 1*( SP / HTAB ) field-vchar ]
		//           field-vchar   = VCHAR / obs-text
		//           obs-fold      = CRLF 1*( SP / HTAB )

		//Parse which bytes are the key... This is at least one token character!
		//We must start with a token character to be at least 1
		if !httpchars.IsTokenChar(line[0]) {
			return ErrNoHeaderKey, true
		}
		keyEnd := 1
		for httpchars.IsTokenChar(line[keyEnd]) {
			keyEnd++
			if keyEnd == len(line) {
				return ErrMalformedHeader, true
			}
		}
		//At this point, keyEnd is 1 more than the last character index of the key,
		// so we verify that we now have a colon
		if line[keyEnd] != ':' {
			return ErrIllegalCharacterInHeaderName, true
		}
		valueStart := keyEnd + 1
		//At this point, we have optional whitespace at the start, so consume it...
		for httpchars.IsWhitespace(line[valueStart]) {
			valueStart++
		}
		valueEnd := valueStart
		//Keep track of the amount of trailing whitespace so that it can be removed at the end
		whitespace := 0
		//Now valueEnd is the first byte of:
		// field-value = *( field-content / obs-fold )
		for {
			//First thing's first, let's write the parser for field-content...
			// Is must start with a VChar
			if httpchars.IsVChar(line[valueEnd]) {
				valueEnd++
				//Then, if there's no whitespace, parse the next thingy!
				if !httpchars.IsWhitespace(line[valueEnd]) {
					whitespace = 0
					continue
				}
				//Add all the whitespace to the value, but keep track of how much
				// in case we need to remove it at the end.
				valueEnd++
				whitespace = 1
				for httpchars.IsWhitespace(line[valueEnd]) {
					valueEnd++
					whitespace++
				}
				//If the character after the whitespace isn't a VChar, then it should be CRLF to end the header.
				if !httpchars.IsVChar(line[valueEnd]) {
					if line[valueEnd] == '\r' && line[valueEnd+1] == '\n' {
						//End the parsing! (Trailing whitespace is removed after this loop)
						break
					}
					return ErrIllegalCharacterInHeaderValue, true
				}
				whitespace = 0
				valueEnd++
				continue
			} else if line[valueEnd] == '\r' && line[valueEnd+1] == '\n' {
				//Now for the obs-fold... Or the end of the line...
				nextLine, nextErr = r.ReadBytes('\n')
				//If the next line starts with whitespace, then it is an obs-fold
				if httpchars.IsWhitespace(nextLine[0]) {
					//So append the next line to the current line...
					line = append(line, nextLine...)
					//And add the CRLF and whitespace to the value
					valueEnd += 3
					whitespace = 0 //TODO: How much?
					//Consume all of the whitespace
					for httpchars.IsWhitespace(line[valueEnd]) {
						valueEnd++
						whitespace++
					}
					continue
				}
				//End of the header
				break
			} else {
				return ErrIllegalCharacterInHeaderValue, true
			}
		}
		//Remove the whitespace
		valueEnd -= whitespace
		h.Add(string(line[0:keyEnd]), string(line[valueStart:valueEnd]))
		if err != nil {
			return err, false
		}
		if nextLine != nil {
			line = nextLine
			err = nextErr
			nextLine = nil
		} else {
			line, err = r.ReadBytes('\n')
		}
	}
}
