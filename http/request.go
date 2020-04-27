package http

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
)

type Request struct {
	MethodStr       string
	URLp            *url.URL
	Version         Version
	Header          *header.Header
	hasCacheControl bool
	cacheControl    RequestCacheControl
}

//URL returns the requested URL. This may be modified.
func (r *Request) URL() *url.URL {
	return r.URLp
}

func (r *Request) ParseCacheControl() error {
	if cc := r.Header.GetValues("cache-control"); len(cc) > 0 {
		var err error
		r.cacheControl, err = ParseRequestCacheControlHeaders(cc, r.Header.GetValues("pragma"))
		if err != nil {
			return err
		}
	} else {
		r.cacheControl = defaultRequestCacheControl
	}
	r.hasCacheControl = true
	return nil
}

func (r *Request) CacheControl() (*RequestCacheControl, error) {
	var err error
	if !r.hasCacheControl {
		err = r.ParseCacheControl()
	}
	return &r.cacheControl, err
}

func (r *Request) FormatCacheControl() error {
	if r.hasCacheControl {
		cc, prag, err := r.cacheControl.Headers()
		r.Header.SetValues("cache-control", cc)
		switch prag {
		case PragmaNoCacheYes:
			if !r.Header.Has("pragma", []string{"no-cache"}) {
				r.Header.Add("pragma", "no-cache")
			}
		case PragmaNoCacheNo:
			//TODO: Don't just delete it...
			r.Header.Del("pragma")
		}
		return err
	}
	return nil
}

func (r *Request) CacheSettings() (jps.RequestCacheSettings, error) {
	cc, err := r.CacheControl()
	return cc, err
}

func (r *Request) DiscardCacheControl() {
	r.hasCacheControl = false
}

//HTTPMethod returns the HTTP method string of the request, for example "GET" or "POST".
//Note thar HTTP methods are case sensitive, so "GeT" is not equivelent to "GET".
func (r *Request) HTTPMethod() string {
	return r.MethodStr
}

//SetHTTPMethod sets the HTTP method string of the request - for example "GET" or "POST".
//Note thar HTTP methods are case sensitive, so "GeT" is not equivelent to "GET".
func (r *Request) SetHTTPMethod(method string) {
	r.MethodStr = method
}

//Method returns the equivelent jps.Method for the request method.
//It is equivelent to calling GetMethod(r.HTTPMethod())
func (r *Request) Method() jps.Method {
	return GetMethod(r.MethodStr)
}

//SetMethod sets the HTTP request method to the method equivelent the given jps.Method.
//It is equivelent to calling r.SetHTTPMethod(GetMethodStr(method)) - but it returns jps.ErrMethodNotSupported if GetMethodStr returns "" (aka there isn't an equivilent method).
func (r *Request) SetMethod(method jps.Method) error {
	r.MethodStr = GetMethodStr(method)
	if r.MethodStr == "" {
		return jps.ErrMethodNotSupported
	}
	return nil
}

func (r *Request) HTTPHeader() *header.Header {
	//TODO: Format cache-control here?
	return r.Header
}

//Ranges returns the ranges specified by the Range header.
//As per RFC 7233, it returns nil, nil if the method is not GET or the range unit is not known.
//Note that the ranges returned are not validated.
func (r *Request) Ranges() ([]jps.Range, error) {
	//RFC 7233 Section 3.1:
	// A server MUST ignore a Range header field received with a request method other than GET.
	if r.MethodStr == "GET" {
		rhs := r.Header.GetValues("Range")
		if len(rhs) != 0 {
			return ParseRangeHeader(rhs[len(rhs)-1])
		}
		//TODO: Section 6.1 DOS attacks.
	}
	return nil, nil
}

//SetRanges sets the Range header such that it specifies the ranges given.
//It DOES NOT validate the ranges, and it may return and error (wrapping jps.ErrNotSupported) if any of the ranges cannot be represented by the HTTP Range header. See FormatRanges and FormatRange for more information.
func (r *Request) SetRanges(ranges []jps.Range) error {
	val, err := FormatRanges(ranges)
	if val == "" {
		r.Header.Del("Range")
	} else {
		r.Header.Set("Range", val)
	}
	return err
}

func (r *Request) Conditions() ([]jps.Condition, error) {
	return Conditions(r.Header, r.MethodStr == "GET" || r.MethodStr == "HEAD"), nil
}

//FormatRange formats the given range in to a HTTP bytes-range-spec.
// A range of 2 non-negative values gets formatted to `start "-" end`
// A starting at a non-negative value and ending at -1 (EOF) gets formatted to `start "-"`
// A starting at a negative value and ending at -1 (EOF) gets formatted to `"-" -start`
// Any other range will return an error which wraps jps.ErrNotSupported - as the HTTP range header cannot be used to represent it.
func FormatRange(r jps.Range) (string, error) {
	if r.Start < 0 {
		if r.End != -1 {
			return "", fmt.Errorf("range %v to %v %w by HTTP", r.Start, r.End, jps.ErrNotSupported)
		}
		return "-" + strconv.FormatUint(uint64(-r.Start), 10), nil
	}
	if r.End < 0 {
		if r.End != -1 {
			return "", fmt.Errorf("range %v to %v %w by HTTP", r.Start, r.End, jps.ErrNotSupported)
		}
		return strconv.FormatUint(uint64(r.Start), 10) + "-", nil
	}
	return strconv.FormatUint(uint64(r.Start), 10) + "-" + strconv.FormatUint(uint64(r.End), 10), nil
}

//FormatRanges calls FormatRange on each range, and returns a string in the correct format for the HTTP Range header - the formatted ranges, separated by a comma.
func FormatRanges(ranges []jps.Range) (string, error) {
	l := len(ranges)
	if l == 0 {
		return "", nil
	}
	out := make([]byte, 0, len(ranges)*6)
	rs, err := FormatRange(ranges[0])
	if err != nil {
		return "", err
	}
	out = append(out, rs...)
	for i := 1; i < l; i++ {
		rs, err = FormatRange(ranges[i])
		if err != nil {
			break
		}
		out = append(out, ',')
		out = append(out, rs...)
	}
	return string(out), err
}

//ParseRangeHeader takes the string from the value of a range header (with whitespace trimmed), and returns the ranges it specifies.
//If the range unit is not recognised, it returns nil, nil - as per RFC7233.
//It returns a HTTP 400 error generated by MakeErrMalformedRangeHeader if invalid syntax is encountered.
//Or a HTTP 416 error if an integer overflow occures.
//This function DOES NOT validate that ranges are valid.
func ParseRangeHeader(v string) ([]jps.Range, error) {
	//If it doesn't start with what we expect, then we don't know the units so we should ignore the header.
	if v[:6] != "bytes=" {
		return nil, nil
	}
	v = v[6:]
	out := make([]jps.Range, 0, 1)
	for {
		//Cases: (1) a-b[,...] - bytes a to b
		//       (2) a-[,...]  - bytes a to end
		//       (3) -b[,...]  - last b bytes
		i := strings.IndexByte(v, '-')
		if i == 0 { // Case (3)
			v = v[i+1:]
			i = strings.IndexByte(v, ',')
			if i < 0 {
				endOffset, err := strconv.ParseUint(v, 10, 63)
				if err != nil {
					if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
						return out, MakeErrRangeNotSatisfiable(err)
					}
					return out, MakeErrMalformedRangeHeader(err)
				}
				return append(out, jps.Range{
					Start: -int64(endOffset),
					End:   -1,
				}), nil
			}
			endOffset, err := strconv.ParseUint(v[:i], 10, 63)
			if err != nil {
				if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
					return out, MakeErrRangeNotSatisfiable(err)
				}
				return out, MakeErrMalformedRangeHeader(err)
			}
			out = append(out, jps.Range{
				Start: -int64(endOffset),
				End:   -1,
			})
			v = v[i+1:]
		} else if i > 0 { // Case (1) or (2)
			start, err := strconv.ParseUint(v[:i], 10, 63)
			if err != nil {
				if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
					return out, MakeErrRangeNotSatisfiable(err)
				}
				return out, MakeErrMalformedRangeHeader(err)
			}
			v = v[i+1:]
			i = strings.IndexByte(v, ',')
			if i < 0 {
				if len(v) == 0 { // Case (2)
					return append(out, jps.Range{
						Start: int64(start),
						End:   -1, // End is implied to be the last byte.
					}), nil
				}
				// Case (1)
				end, err := strconv.ParseUint(v, 10, 63)
				if err != nil {
					if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
						return out, MakeErrRangeNotSatisfiable(err)
					}
					return out, MakeErrMalformedRangeHeader(err)
				}
				return append(out, jps.Range{
					Start: int64(start),
					End:   int64(end),
				}), nil
			}
			if i == 0 { // Case (2)
				out = append(out, jps.Range{
					Start: int64(start),
					End:   -1,
				})
			} else { // Case (1)
				end, err := strconv.ParseUint(v[:i], 10, 63)
				if err != nil {
					if ne, ok := err.(*strconv.NumError); ok && ne.Err == strconv.ErrRange {
						return out, MakeErrRangeNotSatisfiable(err)
					}
					return out, MakeErrMalformedRangeHeader(err)
				}
				out = append(out, jps.Range{
					Start: int64(start),
					End:   int64(end),
				})
			}
			v = v[i+1:]
		} else {
			return out, MakeErrMalformedRangeHeader(ErrBytesRangeSpecMustContainDash)
		}
	}
}

type Response struct {
	StatusCode      uint16
	StatusText      string
	Version         Version
	Header          *header.Header
	hasCacheControl bool
	cacheControl    ResourceCacheControl
}

func (r *Response) ParseCacheControl() error {
	if cc := r.Header.GetValues("cache-control"); len(cc) > 0 {
		var err error
		r.cacheControl, err = ParseResourceCacheControlHeaders(cc, r.Header.Get("expires"))
		if err != nil {
			return err
		}
	} else {
		r.cacheControl = ResourceCacheControl{
			CanTransform: true,
		}
	}
	r.hasCacheControl = true
	return nil
}

func (r *Response) CacheControl() (*ResourceCacheControl, error) {
	var err error
	if !r.hasCacheControl {
		err = r.ParseCacheControl()
	}
	return &r.cacheControl, err
}

func (r *Response) FormatCacheControl() error {
	if r.hasCacheControl {
		cc, expires, err := r.cacheControl.Headers()
		r.Header.SetValues("cache-control", cc)
		if expires != "" {
			r.Header.Set("expires", expires)
		} else {
			r.Header.Del("expires")
		}
		return err
	}
	return nil
}

func (r *Response) CacheSettings() (jps.ResourceCacheSettings, error) {
	cc, err := r.CacheControl()
	return cc, err
}

func (r *Response) WasCached() (bool, error) {
	return len(r.Header.GetValues("Age")) != 0, nil
}

func (r *Response) SetWasCached(b bool) error {
	if b {
		//TODO: Redo with new header interface
		if c, _ := r.WasCached(); !c {
			r.SetCachedAge(0)
		}
	} else {
		r.Header.Del("Age")
	}
}

func (r *Response) CachedAge() (time.Duration, error) {
	ages := r.Header.GetValues("Age")
	var secs time.Duration = 0
	for ageStr := range ages {
		t, err := strconv.ParseUint(string(ageStr), 10, 63)
		ts := time.Duration(t)
		if err != nil {
			return secs * time.Second, err
		}
		if ts > secs {
			secs = ts
		}
	}
	return secs * time.Second, nil
}

func (r *Response) SetCachedAge(t time.Duration) error {
	r.Header.Set("Age", strconv.FormatInt(int64(t/time.Second), 10))
	return nil
}

func (r *Response) DiscardCacheControl() {
	r.hasCacheControl = false
}

func (r *Response) HTTPStatus() (uint16, string) {
	return r.StatusCode, r.StatusText
}

func (r *Response) SetHTTPStatus(code uint16, text string) {
	r.StatusCode = code
	r.StatusText = text
}

func (r *Response) HTTPStatusCode() uint16 {
	return r.StatusCode
}

func (r *Response) SetHTTPStatusCode(code uint16) {
	r.StatusCode = code
	r.StatusText = ""
}

func (r *Response) HTTPStatusText() string {
	return r.StatusText
}

func (r *Response) SetHTTPStatusText(text string) {
	r.StatusText = text
}

func (r *Response) SetStatus(status jps.ResponseStatus) error {
	cws := StatusToStatusCodeMap.Get(status)
	r.StatusCode = cws.Code
	r.StatusText = cws.Str
	if cws.Code == 0 {
		return jps.ErrResponseStatusNotSupported
	}
	return nil
}

func (r *Response) Status() (jps.ResponseStatus, bool) {
	return StatusCodeToStatusMap.Get(r.StatusCode), true
}

func (r *Response) HTTPHeader() *header.Header {
	return r.Header
}

func incomingTo(config *Config, headers *header.Header, con net.Conn) net.Addr {
	if config.UseJPHeaders {
		addrs := headers.GetValuesRawKey("jp-to")
		if len(addrs) == 1 {
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		} else if len(addrs) != 0 {
			//TODO: Warn
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		}
	}
	return con.LocalAddr()
}

func incomingFrom(config *Config, headers *header.Header, con net.Conn) net.Addr {
	if config.UseJPHeaders {
		addrs := headers.GetValuesRawKey("jp-from")
		if len(addrs) == 1 {
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		} else if len(addrs) != 0 {
			//TODO: Warn
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		}
	}
	return con.LocalAddr()
}

func outgoingTo(config *Config, headers *header.Header, con net.Conn) net.Addr {
	if config.UseJPHeaders {
		addrs := headers.GetValuesRawKey("jp-fwd-to")
		if len(addrs) == 1 {
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		} else if len(addrs) != 0 {
			//TODO: Warn
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		}
	}
	return con.LocalAddr()
}

func outgoingFrom(config *Config, headers *header.Header, con net.Conn) net.Addr {
	if config.UseJPHeaders {
		addrs := headers.GetValuesRawKey("jp-fwd-from")
		if len(addrs) == 1 {
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		} else if len(addrs) != 0 {
			//TODO: Warn
			return &net.IPAddr{IP: net.ParseIP(addrs[1])}
		}
	}
	return con.LocalAddr()
}

//RequestRangeWrapper wraps a Request to allow a server to advertise it's ability to accept range requests.
//When the Range method is called for the first time, if the "Accept-Ranges" response header has not been set, it is set to bytes.
type RequestRangeWrapper struct {
	*Request
	//RespHeaders is a pointer to the response headers.
	RespHeaders *header.Header
}

func (rrw RequestRangeWrapper) Range() []jps.Range {
	if len(rrw.RespHeaders.GetValues("Accept-Ranges")) == 0 {
		rrw.RespHeaders.Set("Accept-Ranges", "bytes")
	}
	return rrw.Range()
}
