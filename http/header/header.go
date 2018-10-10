package header

import (
	"net/textproto"
	"strings"
)

func ShouldRemove(b byte) bool {
	return b == ' ' || b == '\t'
}

func RemoveWhitespace(v string) string {
	var i int
	for ShouldRemove(v[i]) {
		i++
	}
	v = v[i:]
	i = len(v) - 1
	for ShouldRemove(v[i]) {
		i--
	}
	return v[:i+1]
}

func ValueIsToken(v, t string) bool {
	return RemoveWhitespace(v) == t
}

func ValueContainsToken(v, t string) bool {
	if i := strings.IndexByte(v, ','); i != -1 {
		return ValueIsToken(v[:i], t) || ValueContainsToken(v[i+1:], t)
	}
	return ValueIsToken(v, t)
}

func ValuesContainToken(vs []string, t string) bool {
	for i := range vs {
		if ValueContainsToken(vs[i], t) {
			return true
		}
	}
	return false
}

func JPSPrefix(s string) string {
	if strings.HasPrefix(s, "Jps-") {
		return "Jps-Had-" + s[4:]
	}
	return s
}

func ToRawKey(key string) string {
	return JPSPrefix(textproto.CanonicalMIMEHeaderKey(key))
}

type Header map[string][]string

//GetValuesRawKey gets the values of the given raw (as it appears in the map) header key.
// Each token is split, and whitespace before and afterwards id removed.
// So, if the header fields in the request were:
//  example: a  , b
//  example: c, d
//  example: e
// Then GetValuesRawKey("Example") would return {"a", "b", "c", "d", "e"}
func (h Header) GetValuesRawKey(key string) []string {
	vs := h[key]
	//Don't do anthing if there aren't any items!
	if len(vs) == 0 {
		return vs
	}
	//RFC 7230 Section 3.2.2 notes that the Set-Cookie header does not use the list syntax
	if key == "Set-Cookie" {
		//Remove all the whitespace and return
		for v := range vs {
			vs[v] = RemoveWhitespace(vs[v])
		}
		return vs
	}
	var i int
	for v := 0; v < len(vs); v++ {
		//Find where the comma is
		i = strings.IndexByte(vs[v], ',')
		if i != -1 {
			//Lengthen the slice by one and precopy the last element
			vs = append(vs, vs[len(vs)-1])
			//Shift everything to the right
			copy(vs[v+2:len(vs)-1], vs[v+1:len(vs)-2])
			//Split - the appended item will be processed next iteration
			vs[v+1] = vs[v][i+1:]
			vs[v] = RemoveWhitespace(vs[v][:i])
		} else {
			vs[v] = RemoveWhitespace(vs[v])
		}
	}
	h[key] = vs
	return vs
}

func (h Header) GetValues(key string) []string {
	return h.GetValuesRawKey(ToRawKey(key))
}

func (h Header) Get(key string) string {
	vs := h.GetValues(key)
	if len(vs) == 0 {
		return ""
	}
	return vs[0]
}

func (h Header) SetValuesRawKey(key string, values []string) {
	h[key] = values
}

func (h Header) Set(key, value string) {
	h.SetValuesRawKey(ToRawKey(key), []string{value})
}

func (h Header) SetValues(key string, values []string) {
	h.SetValuesRawKey(ToRawKey(key), values)
}

func (h Header) AddRawKey(key, value string) {
	h.SetValuesRawKey(key, append(h.GetValuesRawKey(key), value))
}

func (h Header) Add(key, value string) {
	h.AddRawKey(ToRawKey(key), value)
}

func (h Header) DelRawKey(key string) {
	delete(h, key)
}

func (h Header) Del(key string) {
	h.DelRawKey(ToRawKey(key))
}

func (h Header) ContainsTokenRaw(key, token string) bool {
	vs := h.GetValuesRawKey(key)
	for i := range vs {
		if strings.ToLower(vs[i]) == token {
			return true
		}
	}
	return false
}

func (h Header) ContainsToken(key, token string) bool {
	return h.ContainsTokenRaw(ToRawKey(key), strings.ToLower(token))
}
