package http

import (
	//"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	//"github.com/JOTPOT-UK/JOTPOT-Server/http/httpchars"
	"strconv"
	"strings"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/httpchars"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

//pastTime is a time in the past
var pastTime = time.Date(2001, time.May, 8, 12, 0, 0, 0, time.UTC)

//pastTimeStr is a http-date string in the past
const pastTimeStr = "Tue, 08 May 2001 12:00:00 GMT"

//CacheControlDirective represents a HTTP Cache-Control directive.
type CacheControlDirective struct {
	//Name is the name of the directive.
	//Since directives are case-insensitive, they should be stored in lower case here
	// for comparison.
	//When parsed, they are converted to lower case, although not when formatted.
	Name string
	//Arg is the argument for the directive.
	//An empty string represents no argument (since the HTTP spec doesn't permit an empty argument).
	Arg string
}

//HeaderString returns the string to be used in the HTTP Cache-Control header to represent this directive.
func (ccd *CacheControlDirective) HeaderString() (string, error) {
	if len(ccd.Arg) == 0 {
		return strings.ToLower(ccd.Name), nil
	}
	arg, err := httpchars.FormatTokenOrQuotedString(ccd.Arg)
	return ccd.Name + "=" + string(arg), err
}

//String can panic... Use HeaderString() for things; this is just to help with debugging.
func (ccd *CacheControlDirective) String() string {
	str, err := ccd.HeaderString()
	if err != nil {
		panic(err)
	}
	return str
}

//ParseCacheControlDirective takes a HTTP Cache-Control directive as a string and parses it to a
//CacheControlDirective. The input string is assumed to be trimmed
//(so to have no leading or trailing spaces/tabs).
func ParseCacheControlDirective(s string) (CacheControlDirective, error) {
	i := strings.IndexByte(s, '=')
	if i == -1 {
		return CacheControlDirective{Name: s}, nil
	}
	var err error
	arg := s[i+1:]
	if len(arg) != 0 && arg[0] == '"' {
		var argBytes []byte
		argBytes, err = httpchars.ParseFullQuotedStringNoStartingQuote(arg[1:])
		arg = string(argBytes)
	}
	//arg, err := httpchars.Parse
	return CacheControlDirective{
		Name: s[:i],
		Arg:  s[i+1:],
	}, err
}

//ParseCacheControlDirectives calls ParseCacheControlDirective on each of the strings in the given slice
//and returns the result. Note that each string is expected to be trimmed.
func ParseCacheControlDirectives(dstrs []string) ([]CacheControlDirective, error) {
	ds := make([]CacheControlDirective, len(dstrs))
	for _, dstr := range dstrs {
		d, err := ParseCacheControlDirective(dstr)
		ds = append(ds, d)
		if err != nil {
			return ds, err
		}
	}
	return ds, nil
}

//ExpiryTime represents when a cached resource expires.
//It's not too inefficient and can be set either by a max age or an expiry time.
type ExpiryTime struct {
	kind    expiryTimeKind
	maxAge  time.Duration
	expires time.Time
}

type expiryTimeKind uint8

const (
	expiryTimeNone expiryTimeKind = iota
	expiryTimeMaxAge
	expiryTimeExpires
	expiryTimeBoth
)

func (et *ExpiryTime) SetMaxAge(maxAge time.Duration) {
	et.maxAge = maxAge
	et.kind = expiryTimeMaxAge
}
func (et *ExpiryTime) SetExpires(expires time.Time) {
	et.expires = expires
	et.kind = expiryTimeExpires
}

func (et *ExpiryTime) makeBothFromExpires() {
	s := time.Now().Sub(et.expires)
	if s < 0 {
		s = 0
	}
	et.maxAge = s
	et.kind = expiryTimeBoth
}
func (et *ExpiryTime) makeBothFromMaxAge() {
	if et.maxAge < 0 {
		et.expires = time.Time{}
	} else if et.maxAge == 0 {
		et.expires = pastTime
	} else {
		et.expires = time.Now().Add(et.maxAge)
	}
	et.kind = expiryTimeBoth
}

//MaxAge returns the time until expiry.
func (et *ExpiryTime) MaxAge() time.Duration {
	switch et.kind {
	case expiryTimeBoth, expiryTimeMaxAge:
		return et.maxAge
	case expiryTimeExpires:
		et.makeBothFromExpires()
		return et.maxAge
	default:
		return -1
	}
}

//Expires returns the expiry time!
func (et *ExpiryTime) Expires() (time.Time, bool) {
	switch et.kind {
	case expiryTimeBoth:
		return et.expires, et.maxAge >= 0
	case expiryTimeExpires:
		return et.expires, true
	case expiryTimeMaxAge:
		et.makeBothFromMaxAge()
		return et.expires, et.maxAge >= 0
	default:
		return time.Time{}, false
	}
}

//ExpiresString returns Expires(), formatted in http-date format.
func (et *ExpiryTime) ExpiresString() string {
	if t, ok := et.Expires(); ok {
		return FormatTime(t)
	}
	return ""
}

func (et *ExpiryTime) MaxAgeAndExpires() (time.Duration, time.Time) {
	switch et.kind {
	case expiryTimeBoth:
		return et.maxAge, et.expires
	case expiryTimeMaxAge:
		et.makeBothFromMaxAge()
		return et.maxAge, et.expires
	case expiryTimeExpires:
		et.makeBothFromExpires()
		return et.maxAge, et.expires
	default:
		return -1, time.Time{}
	}
}

func (et *ExpiryTime) MaxAgeAndExpiresString() (time.Duration, string) {
	maxAge, expires := et.MaxAgeAndExpires()
	if maxAge < 0 {
		return maxAge, ""
	}
	return maxAge, FormatTime(expires)
}

//ResourceCacheControl provides the ResourceCacheSettings interface for HTTP Responses.
type ResourceCacheControl struct {
	PrivateMode, PublicMode     jps.ResourceCacheMode
	CanTransform                bool
	PublicExpiry, PrivateExpiry ExpiryTime
	//Extras are directives that were not understood.
	Extras []CacheControlDirective
}

func (*ResourceCacheControl) CacheSupported() bool { return true }

func (rcc *ResourceCacheControl) CacheTransformAllowed() (bool, error) {
	return rcc.CanTransform, nil
}
func (rcc *ResourceCacheControl) SetCacheTransformAllowed(b bool) error {
	rcc.CanTransform = b
	return nil
}

var (
	noStore = CacheControlDirective{
		Name: "no-store",
	}
	noCache = CacheControlDirective{
		Name: "no-cache",
	}
	private = CacheControlDirective{
		Name: "private",
	}
	public = CacheControlDirective{
		Name: "public",
	}
	mustRevalidate = CacheControlDirective{
		Name: "must-revalidate",
	}
	proxyRevalidate = CacheControlDirective{
		Name: "proxy-revalidate",
	}
	sMaxAge0 = CacheControlDirective{
		Name: "s-maxage",
		Arg:  "0",
	}
	maxAge0 = CacheControlDirective{
		Name: "max-age",
		Arg:  "0",
	}
)

func maxAge(d time.Duration) CacheControlDirective {
	return CacheControlDirective{
		Name: "max-age",
		Arg:  strconv.FormatInt(int64(d/time.Second), 10),
	}
}

func sMaxAge(d time.Duration) CacheControlDirective {
	return CacheControlDirective{
		Name: "s-maxage", // Whhyyyyyy isn't this s-max-age?????
		Arg:  strconv.FormatInt(int64(d/time.Second), 10),
	}
}

type ccds = []CacheControlDirective

//Headers returns the values of the Cache-Control HTTP header
//and the value of the Expires HTTP header (or "" if the header should not have a value).
func (rcc *ResourceCacheControl) Headers() ([]string, string, error) {
	ds, expires := rcc.headers2()
	sds := make([]string, len(ds), len(ds)+1)
	var err error
	for i, d := range ds {
		sds[i], err = d.HeaderString()
		if err != nil {
			return nil, "", err
		}
	}
	if !rcc.CanTransform {
		sds = append(sds, "no-transform")
	}
	return sds, expires, nil
}

func (rcc *ResourceCacheControl) headers2() ([]CacheControlDirective, string) {
	ds, expires := rcc.headers1()
	return append(ds, rcc.Extras...), expires
}

func (rcc *ResourceCacheControl) headers1() ([]CacheControlDirective, string) {
	switch rcc.PublicMode {
	case jps.ResourceCacheModeNoStore:
		switch rcc.PrivateMode {
		case jps.ResourceCacheModeCanCache, jps.ResourceCacheModeUnspecified: // TODO: Is this the correct unspecified behaviour?
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge >= 0 {
				return ccds{private, maxAge(privateMaxAge)}, pastTimeStr
			}
			return ccds{private}, pastTimeStr

		case jps.ResourceCacheModeMustRevalidate:
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge > 0 {
				return ccds{private, mustRevalidate, maxAge(privateMaxAge)}, pastTimeStr
			} else if privateMaxAge == 0 {
				return ccds{private, noCache}, pastTimeStr
			} else {
				return ccds{private}, pastTimeStr
			}
		case jps.ResourceCacheModeNoCache:
			return ccds{private, noCache}, pastTimeStr
		default:
			return ccds{noStore}, pastTimeStr
		}
	case jps.ResourceCacheModeNoCache:
		switch rcc.PrivateMode {
		case jps.ResourceCacheModeCanCache, jps.ResourceCacheModeMustRevalidate:
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge > 0 {
				return ccds{mustRevalidate, maxAge(privateMaxAge), sMaxAge0}, pastTimeStr
			} else if privateMaxAge == 0 {
				return ccds{noCache}, pastTimeStr
			} else {
				return ccds{mustRevalidate, sMaxAge0}, pastTimeStr
			}
		case jps.ResourceCacheModeNoCache, jps.ResourceCacheModeUnspecified:
			return ccds{noCache}, pastTimeStr
		default:
			return ccds{noStore}, pastTimeStr
		}
	case jps.ResourceCacheModeMustRevalidate:
		publicMaxAge, expires := rcc.PublicExpiry.MaxAgeAndExpiresString()
		switch rcc.PrivateMode {
		case jps.ResourceCacheModeCanCache:
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge == publicMaxAge {
				return ccds{proxyRevalidate, maxAge(publicMaxAge)}, expires
			}
			return ccds{proxyRevalidate, maxAge(privateMaxAge), sMaxAge(publicMaxAge)}, expires
		case jps.ResourceCacheModeMustRevalidate, jps.ResourceCacheModeUnspecified:
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge == publicMaxAge {
				return ccds{mustRevalidate, maxAge(privateMaxAge)}, expires
			}
			return ccds{mustRevalidate, maxAge(privateMaxAge), sMaxAge(publicMaxAge)}, expires
		case jps.ResourceCacheModeNoCache:
			if publicMaxAge > 0 {
				return ccds{mustRevalidate, maxAge0, sMaxAge(publicMaxAge)}, pastTimeStr
			} else if publicMaxAge == 0 {
				return ccds{noCache}, pastTimeStr
			}
		default:
			return ccds{noStore}, pastTimeStr
		}
	case jps.ResourceCacheModeCanCache:
		publicMaxAge, expires := rcc.PublicExpiry.MaxAgeAndExpiresString()
		switch rcc.PrivateMode {
		case jps.ResourceCacheModeCanCache, jps.ResourceCacheModeUnspecified:
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge == publicMaxAge {
				return ccds{public, maxAge(privateMaxAge)}, expires
			}
			return ccds{public, maxAge(privateMaxAge), sMaxAge(publicMaxAge)}, expires
		case jps.ResourceCacheModeMustRevalidate:
			privateMaxAge := rcc.PrivateExpiry.MaxAge()
			if privateMaxAge == publicMaxAge {
				return ccds{mustRevalidate, maxAge(privateMaxAge)}, expires
			}
			return ccds{mustRevalidate, maxAge(privateMaxAge), sMaxAge(publicMaxAge)}, expires
		case jps.ResourceCacheModeNoCache:
			return ccds{mustRevalidate, maxAge0, sMaxAge(publicMaxAge)}, pastTimeStr
		case jps.ResourceCacheModeNoStore:
			return ccds{noStore}, pastTimeStr
		}
	}
	return nil, ""
}

func (rcc *ResourceCacheControl) PublicCacheMode() (jps.ResourceCacheMode, error) {
	return rcc.PublicMode, nil
}

func (rcc *ResourceCacheControl) PrivateCacheMode() (jps.ResourceCacheMode, error) {
	return rcc.PrivateMode, nil
}

func (rcc *ResourceCacheControl) SetPublicCacheMode(m jps.ResourceCacheMode) error {
	rcc.PublicMode = m
	return nil
}

func (rcc *ResourceCacheControl) SetPrivateCacheMode(m jps.ResourceCacheMode) error {
	rcc.PrivateMode = m
	return nil
}

func (rcc *ResourceCacheControl) SetCacheMode(m jps.ResourceCacheMode) error {
	rcc.PrivateMode = m
	rcc.PublicMode = m
	return nil
}

func (rcc *ResourceCacheControl) SetCacheMaxAge(age time.Duration) error {
	rcc.PrivateExpiry.SetMaxAge(age)
	rcc.PublicExpiry.SetMaxAge(age)
	return nil
}

func (rcc *ResourceCacheControl) SetPublicCacheMaxAge(age time.Duration) error {
	rcc.PublicExpiry.SetMaxAge(age)
	return nil
}

func (rcc *ResourceCacheControl) SetPrivateCacheMaxAge(age time.Duration) error {
	rcc.PrivateExpiry.SetMaxAge(age)
	return nil
}

func (rcc *ResourceCacheControl) PublicCacheMaxAge() (time.Duration, error) {
	return rcc.PublicExpiry.MaxAge(), nil
}

func (rcc *ResourceCacheControl) PrivateCacheMaxAge() (time.Duration, error) {
	return rcc.PrivateExpiry.MaxAge(), nil
}

func (rcc *ResourceCacheControl) SetCacheExpiry(t time.Time) error {
	rcc.PrivateExpiry.SetExpires(t)
	rcc.PublicExpiry.SetExpires(t)
	return nil
}

func (rcc *ResourceCacheControl) SetPublicCacheExpires(t time.Time) error {
	rcc.PublicExpiry.SetExpires(t)
	return nil
}

func (rcc *ResourceCacheControl) SetPrivateCacheExpires(t time.Time) error {
	rcc.PrivateExpiry.SetExpires(t)
	return nil
}

func (rcc *ResourceCacheControl) SetCacheExpires(t time.Time) error {
	rcc.PrivateExpiry.SetExpires(t)
	rcc.PublicExpiry.SetExpires(t)
	return nil
}

func (rcc *ResourceCacheControl) PublicCacheExpires() (time.Time, bool, error) {
	t, b := rcc.PublicExpiry.Expires()
	return t, b, nil
}

func (rcc *ResourceCacheControl) PrivateCacheExpires() (time.Time, bool, error) {
	t, b := rcc.PrivateExpiry.Expires()
	return t, b, nil
}

func doExpiryStuff(rcc *ResourceCacheControl) {
	if rcc.PrivateMode == jps.ResourceCacheModeMustRevalidate && rcc.PrivateExpiry.MaxAge() == 0 {
		rcc.PrivateMode = jps.ResourceCacheModeNoCache
	}
	if rcc.PrivateMode == jps.ResourceCacheModeMustRevalidate && rcc.PrivateExpiry.MaxAge() == 0 {
		rcc.PrivateMode = jps.ResourceCacheModeNoCache
	}
}

func ParseResourceCacheControlDirectives(ds []CacheControlDirective, expires string) (ResourceCacheControl, error) {
	out := ResourceCacheControl{
		CanTransform: true,
	}
	hasExtras := false
	hasMaxAge := false
	var maxAge time.Duration
	hasSMaxAge := false
	var sMaxAge time.Duration
	private := false
	public := false
	noCache := false
	noStore := false
	mustRevalidate := false
	proxyRevalidate := false
	for _, d := range ds {
		switch d.Name {
		case "no-transform":
			out.CanTransform = false
		case "max-age":
			maxAgeSecs, err := strconv.ParseInt(d.Arg, 10, 64)
			if err != nil {
				return out, err
			}
			maxAge = time.Duration(maxAgeSecs) * time.Second
			hasMaxAge = true
		case "s-maxage":
			sMaxAgeSecs, err := strconv.ParseUint(d.Arg, 10, 64)
			if err != nil {
				return out, err
			}
			sMaxAge = time.Duration(sMaxAgeSecs) * time.Second
			hasSMaxAge = true
		case "private":
			private = true
		case "public":
			public = true
		case "no-cache":
			noCache = true
		case "no-store":
			noStore = true
		case "must-revalidate":
			mustRevalidate = true
		case "proxy-revalidate":
			proxyRevalidate = true
		default:
			if !hasExtras {
				out.Extras = make([]CacheControlDirective, 0, 1)
				hasExtras = true
			}
			out.Extras = append(out.Extras, d)
		}
	}
	if private {
		out.PublicMode = jps.ResourceCacheModeNoStore
		if noStore {
			out.PrivateMode = jps.ResourceCacheModeNoStore
		} else if noCache {
			out.PrivateMode = jps.ResourceCacheModeNoCache
		} else if mustRevalidate {
			out.PrivateMode = jps.ResourceCacheModeMustRevalidate
		}
	} else {
		if noStore {
			out.PrivateMode = jps.ResourceCacheModeNoStore
			out.PublicMode = jps.ResourceCacheModeNoStore
		} else if noCache {
			out.PrivateMode = jps.ResourceCacheModeNoCache
			out.PublicMode = jps.ResourceCacheModeNoStore
		} else if mustRevalidate {
			out.PrivateMode = jps.ResourceCacheModeMustRevalidate
			out.PublicMode = jps.ResourceCacheModeNoStore
		} else if proxyRevalidate {
			out.PublicMode = jps.ResourceCacheModeMustRevalidate
			if public {
				out.PrivateMode = jps.ResourceCacheModeCanCache
			}
		} else if public {
			out.PublicMode = jps.ResourceCacheModeCanCache
			out.PrivateMode = jps.ResourceCacheModeCanCache
		}
	}
	if hasSMaxAge {
		out.PublicExpiry.SetMaxAge(sMaxAge)
		if hasMaxAge {
			out.PrivateExpiry.SetMaxAge(maxAge)
		} else if len(expires) > 0 {
			if et, err := ParseTime(expires); err != nil {
				out.PrivateExpiry.SetExpires(et)
			}
		}
		doExpiryStuff(&out)
	} else if hasMaxAge {
		out.PublicExpiry.SetMaxAge(maxAge)
		out.PrivateExpiry.SetMaxAge(maxAge)
		doExpiryStuff(&out)
	} else if len(expires) > 0 {
		if et, err := ParseTime(expires); err != nil {
			out.PrivateExpiry.SetExpires(et)
			doExpiryStuff(&out)
		}
	}
	return out, nil
}

func ParseResourceCacheControlHeaders(cacheControl []string, expires string) (ResourceCacheControl, error) {
	ds, err := ParseCacheControlDirectives(cacheControl)
	if err != nil {
		return ResourceCacheControl{}, err
	}
	return ParseResourceCacheControlDirectives(ds, expires)
}

type PragmaNoCache uint8

const (
	PragmaNoCacheDoesntMatter PragmaNoCache = iota
	PragmaNoCacheNo
	PragmaNoCacheYes
)

type RequestCacheControl struct {
	Mode                          jps.RequestCacheMode
	_MaxAge, _MinFresh, _MaxStale time.Duration
	CanTransform                  bool
	//Extras are directives that were not understood.
	Extras []CacheControlDirective
}

var defaultRequestCacheControl = RequestCacheControl{
	_MaxAge: -1, _MinFresh: -1, _MaxStale: -1,
	CanTransform: true,
}

func (rcc *RequestCacheControl) Headers() ([]string, PragmaNoCache, error) {
	ds, prag := rcc.modeHeaders()
	if !rcc.CanTransform {
		ds = append(ds, "no-transform")
	}
	for _, extra := range rcc.Extras {
		extraStr, err := extra.HeaderString()
		if err != nil {
			return ds, prag, err
		}
		ds = append(ds, extraStr)
	}
	return ds, prag, nil
}

func (rcc *RequestCacheControl) modeHeaders() ([]string, PragmaNoCache) {
	switch rcc.Mode {
	case jps.RequestCacheModeNoCache:
		return []string{"no-cache"}, PragmaNoCacheYes
	case jps.RequestCacheModeNoStore:
		return []string{"no-store"}, PragmaNoCacheNo
	case jps.RequestCacheModeOnlyIfCached:
		return []string{"only-if-cached"}, PragmaNoCacheNo
	default:
		return nil, PragmaNoCacheDoesntMatter
	}
}

func (_ *RequestCacheControl) CacheSupported() bool { return true }

func (rcc *RequestCacheControl) CacheMode() (jps.RequestCacheMode, error) {
	return rcc.Mode, nil
}
func (rcc *RequestCacheControl) SetCacheMode(mode jps.RequestCacheMode) error {
	rcc.Mode = mode
	return nil
}

func (rcc *RequestCacheControl) CacheMaxAge() (time.Duration, error) {
	return rcc._MaxAge, nil
}
func (rcc *RequestCacheControl) CacheMinFresh() (time.Duration, error) {
	return rcc._MinFresh, nil
}
func (rcc *RequestCacheControl) CacheMaxStale() (time.Duration, error) {
	return rcc._MaxStale, nil
}

func (rcc *RequestCacheControl) SetCacheMaxAge(d time.Duration) error {
	rcc._MaxAge = d
	return nil
}
func (rcc *RequestCacheControl) SetCacheMinFresh(d time.Duration) error {
	rcc._MinFresh = d
	return nil
}
func (rcc *RequestCacheControl) SetCacheMaxStale(d time.Duration) error {
	rcc._MaxStale = d
	return nil
}

func (rcc *RequestCacheControl) CacheTransformAllowed() (bool, error) {
	return rcc.CanTransform, nil
}
func (rcc *RequestCacheControl) SetCacheTransformAllowed(b bool) error {
	rcc.CanTransform = b
	return nil
}

func ParseRequestCacheControlDirectives(ds []CacheControlDirective) (RequestCacheControl, error) {
	out := defaultRequestCacheControl
	hasExtras := false
	noStore := false
	noCache := false
	onlyIfCached := false
	for _, d := range ds {
		switch d.Name {
		case "no-transform":
			out.CanTransform = false
		case "max-age":
			secs, err := strconv.ParseInt(d.Arg, 10, 64) // TODO: Handle overflows
			out._MaxAge = time.Duration(secs) * time.Second
			if err != nil {
				return out, err
			}
		case "min-fresh":
			secs, err := strconv.ParseInt(d.Arg, 10, 64) // TODO: Handle overflows
			out._MinFresh = time.Duration(secs) * time.Second
			if err != nil {
				return out, err
			}
		case "max-stale":
			secs, err := strconv.ParseInt(d.Arg, 10, 64) // TODO: Handle overflows
			out._MaxStale = time.Duration(secs) * time.Second
			if err != nil {
				return out, err
			}
		case "no-cache":
			noCache = true
		case "no-store":
			noStore = true
		case "only-if-cached":
			onlyIfCached = true
		default:
			if !hasExtras {
				out.Extras = make([]CacheControlDirective, 0, 1)
				hasExtras = true
			}
			out.Extras = append(out.Extras, d)
		}
	}
	if onlyIfCached {
		out.Mode = jps.RequestCacheModeOnlyIfCached
	} else if noStore {
		out.Mode = jps.RequestCacheModeNoStore
	} else if noCache {
		out.Mode = jps.RequestCacheModeNoCache
	}
	return out, nil
}

//TODO: Pragma
func ParseRequestCacheControlHeaders(cacheControl []string, pragma []string) (RequestCacheControl, error) {
	ds, err := ParseCacheControlDirectives(cacheControl)
	if err != nil {
		return defaultRequestCacheControl, err
	}
	return ParseRequestCacheControlDirectives(ds)
}
