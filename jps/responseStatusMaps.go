package jps

import "strconv"

type ResponseStatusToUint16Map struct {
	OK          []uint16
	Redirect    []uint16
	ClientError []uint16
	ServerError []uint16
	Default     uint16
}

func (maps *ResponseStatusToUint16Map) Get(s ResponseStatus) uint16 {
	var slice []uint16
	switch s.Type() {
	case ResponseTypeOK:
		slice = maps.OK
	case ResponseTypeRedirect:
		slice = maps.Redirect
	case ResponseTypeClientError:
		slice = maps.ClientError
	case ResponseTypeServerError:
		slice = maps.ServerError
	default:
		return maps.Default
	}
	if len(slice) == 0 {
		return maps.Default
	}
	code := int(s.GetCode())
	if code >= len(slice) {
		code = 0
	}
	return slice[code]
}

func (maps *ResponseStatusToUint16Map) Set(s ResponseStatus, u uint16) {
	var slice *[]uint16
	switch s.Type() {
	case ResponseTypeOK:
		slice = &maps.OK
	case ResponseTypeRedirect:
		slice = &maps.Redirect
	case ResponseTypeClientError:
		slice = &maps.ClientError
	case ResponseTypeServerError:
		slice = &maps.ServerError
	default:
		panic("invalid ResponseStatus")
	}
	code := int(s.GetCode())
	l := len(*slice)
	if l == code {
		(*slice) = append(*slice, u)
	} else if l < code {
		//min(l)=0 => min(code)=1
		if code == 1 {
			//code=1 => l=0
			*slice = append(*slice, maps.Default, u)
		} else {
			//min(code)=2
			var dflt uint16
			if l == 0 {
				dflt = maps.Default
				*slice = append(*slice, dflt, dflt)
				l = 2
			} else {
				//min(l)=1
				dflt = (*slice)[0]
				*slice = append(*slice, dflt)
				l++
			}
			//min(l)=2
			for l < code {
				*slice = append(*slice, dflt)
				l++
			}
			*slice = append(*slice, u)
		}
	} else {
		(*slice)[code] = u
	}
}

func (maps *ResponseStatusToUint16Map) SetMap(m map[ResponseStatus]uint16) {
	for s, u := range m {
		maps.Set(s, u)
	}
}

type ResponseStatusUint16Pair struct {
	S ResponseStatus
	U uint16
}

func (maps *ResponseStatusToUint16Map) SetAll(m []ResponseStatusUint16Pair) {
	for _, pair := range m {
		maps.Set(pair.S, pair.U)
	}
}

type CodeWithString struct {
	Code uint16
	Str  string
}

func (cws *CodeWithString) String() string {
	return strconv.FormatUint(uint64(cws.Code), 10) + " " + cws.Str
}

type ResponseStatusToCodeWithStringMap struct {
	OK          []CodeWithString
	Redirect    []CodeWithString
	ClientError []CodeWithString
	ServerError []CodeWithString
	Default     CodeWithString
}

func NewResponseStatusToCodeWithStringMap(Default CodeWithString) ResponseStatusToCodeWithStringMap {
	return ResponseStatusToCodeWithStringMap{Default: Default}
}

func NewResponseStatusToCodeWithStringMapFromValuePairs(Default CodeWithString, values []ResponseStatusCodeWithStringPair) ResponseStatusToCodeWithStringMap {
	maps := NewResponseStatusToCodeWithStringMap(Default)
	maps.SetAll(values)
	return maps
}

func (maps *ResponseStatusToCodeWithStringMap) Get(s ResponseStatus) *CodeWithString {
	var slice []CodeWithString
	switch s.Type() {
	case ResponseTypeOK:
		slice = maps.OK
	case ResponseTypeRedirect:
		slice = maps.Redirect
	case ResponseTypeClientError:
		slice = maps.ClientError
	case ResponseTypeServerError:
		slice = maps.ServerError
	default:
		return &maps.Default
	}
	if len(slice) == 0 {
		return &maps.Default
	}
	code := int(s.GetCode())
	if code >= len(slice) {
		code = 0
	}
	return &slice[code]
}

func (maps *ResponseStatusToCodeWithStringMap) Set(s ResponseStatus, cws CodeWithString) {
	var slice *[]CodeWithString
	switch s.Type() {
	case ResponseTypeOK:
		slice = &maps.OK
	case ResponseTypeRedirect:
		slice = &maps.Redirect
	case ResponseTypeClientError:
		slice = &maps.ClientError
	case ResponseTypeServerError:
		slice = &maps.ServerError
	default:
		panic("invalid ResponseStatus")
	}
	switch s.Type() {
	case ResponseTypeOK:
		slice = &maps.OK
	case ResponseTypeRedirect:
		slice = &maps.Redirect
	case ResponseTypeClientError:
		slice = &maps.ClientError
	case ResponseTypeServerError:
		slice = &maps.ServerError
	default:
		panic("invalid ResponseStatus")
	}
	code := int(s.GetCode())
	l := len(*slice)
	if l == code {
		(*slice) = append(*slice, cws)
	} else if l < code {
		//min(l)=0 => min(code)=1
		if code == 1 {
			//code=1 => l=0
			*slice = append(*slice, maps.Default, cws)
		} else {
			//min(code)=2
			var dflt CodeWithString
			if l == 0 {
				dflt = maps.Default
				*slice = append(*slice, dflt, dflt)
				l = 2
			} else {
				//min(l)=1
				dflt = (*slice)[0]
				*slice = append(*slice, dflt)
				l++
			}
			//min(l)=2
			for l < code {
				*slice = append(*slice, dflt)
				l++
			}
			*slice = append(*slice, cws)
		}
	} else {
		(*slice)[code] = cws
	}
}

func (maps *ResponseStatusToCodeWithStringMap) SetMap(m map[ResponseStatus]CodeWithString) {
	for s, u := range m {
		maps.Set(s, u)
	}
}

type ResponseStatusCodeWithStringPair struct {
	S ResponseStatus
	U CodeWithString
}

func (maps *ResponseStatusToCodeWithStringMap) SetAll(m []ResponseStatusCodeWithStringPair) {
	for _, pair := range m {
		maps.Set(pair.S, pair.U)
	}
}
