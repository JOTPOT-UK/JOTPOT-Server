package http1

import (
	"errors"
	gohttp "net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/mediatype"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

var ErrRangeTooLarge = errors.New("range too large")
var ErrMultipleContentRangeHeaders = errors.New("multiple content-range headers")
var ErrMalformedContentRangeHeader = errors.New("malformed content-range header")
var ErrMultipleContentTypeHeaders = errors.New("multiple content-type headers")
var ErrMultipleContentLocationHeaders = errors.New("multiple content-location headers")
var ErrMultipleLastModifiedHeaders = errors.New("multiple last-modified headers")

type HeaderMetadata struct {
	H *header.Header
}

func (m HeaderMetadata) WantSize() jps.MetadataWanted {
	return jps.MetadataHasEffect
}

func (m HeaderMetadata) Size() (int64, error) {
	//Get the size if specified by the Content-Range header.
	if hasContentRange, rangeStr, sizeStr, err := m.splitContentRange(); hasContentRange {
		if sizeStr == "*" || err != nil {
			if rangeStr == "*" && err == nil {
				err = ErrMalformedContentRangeHeader
			}
			return -1, err
		}
		size, err := strconv.ParseUint(sizeStr, 10, 63)
		return int64(size), err
	}
	lengths := m.H.GetValues("Content-Length")
	if len(lengths) == 0 {
		return -1, nil
	}
	if len(lengths) == 1 {
		length, err := strconv.ParseUint(lengths[0], 10, 63)
		if err != nil {
			//TODO: Negative lengths?????
			return int64(length), http.MakeErrMalformedContentLength(err)
		}
		return int64(length), nil
	}
	return -1, http.ErrTooManyContentLength
}

func (m HeaderMetadata) removeContentLength() {
	m.H.Del("Content-Length")
	codes := m.H.GetValues("Transfer-Encoding")
	last := len(codes) - 1
	if last < 0 || codes[last] != "chunked" {
		m.H.Add("Transfer-Encoding", "chunked")
	}
}
func (m HeaderMetadata) setContentLength(l int64) {
	m.H.Set("Content-Length", strconv.FormatInt(l, 10))
	codes := m.H.GetValues("Transfer-Encoding")
	last := len(codes) - 1
	if last > -1 && codes[last] == "chunked" {
		m.H.SetValues("Transfer-Encoding", codes[:last])
	}
}

func (m HeaderMetadata) SetSize(size int64) error {
	hasContentRange, rangeStr, _, err := m.splitContentRange()
	if hasContentRange {
		if size < 0 {
			if size != -1 {
				panic("size must be >= -1")
			}
			m.setContentRange(rangeStr, "*")
		} else {
			m.setContentRange(rangeStr, strconv.FormatInt(size, 10))
		}
		return err
	}
	if size < 0 {
		if size != -1 {
			panic("size must be >= -1")
		}
		m.removeContentLength()
	} else {
		m.setContentLength(size)
	}
	return nil
}

func (m HeaderMetadata) WantRange() jps.MetadataWanted {
	return jps.MetadataHasEffect
}

func (m HeaderMetadata) setContentRange(r, size string) {
	m.H.Set("Content-Range", "bytes "+r+"/"+size)
}

func (m HeaderMetadata) splitContentRange() (bool, string, string, error) {
	crs := m.H.GetValues("Content-Range")
	if len(crs) < 1 {
		return false, "", "", nil
	} else if len(crs) > 1 {
		return true, "", "", ErrMultipleContentRangeHeaders
	}
	cr := crs[0]
	if cr[:6] != "bytes " {
		return true, "", "", nil
	}
	cr = cr[6:]
	i := strings.IndexByte(cr, '/')
	if i == -1 {
		return true, "", "", ErrMalformedContentRangeHeader
	}
	return true, cr[:i], cr[i+1:], nil
}

func (m HeaderMetadata) Range() (jps.Range, error) {
	hasContentRange, rangeStr, _, err := m.splitContentRange()
	if !hasContentRange || rangeStr == "*" || err != nil {
		return jps.Range{0, -1}, err
	}
	i := strings.IndexByte(rangeStr, '-')
	if i == -1 {
		return jps.Range{0, -1}, ErrMalformedContentRangeHeader
	}
	start, err := strconv.ParseUint(rangeStr[:i], 10, 63)
	if err != nil {
		//Don't return yet if we overflowed...
		if ne, ok := err.(*strconv.NumError); ok && ne.Err != strconv.ErrRange {
			return jps.Range{int64(start), -1}, ErrMalformedContentRangeHeader
		}
	}
	end, err2 := strconv.ParseUint(rangeStr[i+1:], 10, 63)
	//If there wasn't an error here, carry the previous error forwards.
	if err2 == nil {
		err2 = err
	}
	return jps.Range{
		Start: int64(start),
		End:   int64(end),
	}, err2
}

var ErrCannotDetermineByteRangePositions = errors.New("cannot determine byte range positions")

func (m HeaderMetadata) SetRange(r jps.Range) error {
	size, err := m.Size()
	if size > -1 {
		start, end := r.StartEnd(size)
		if start < 0 || end < 0 {
			return ErrRangeTooLarge
		}
		m.setContentRange(
			strconv.FormatUint(uint64(start), 10)+"-"+
				strconv.FormatUint(uint64(end), 10),
			strconv.FormatUint(uint64(size), 10),
		)
		m.setContentLength(end - start + 1)
	} else if r.Start < 0 || r.End < 0 {
		if err != nil {
			return err
		}
		m.setContentRange("*", "*")
		return http.MakeErrRangeNotSatisfiable(ErrCannotDetermineByteRangePositions)
	} else {
		m.setContentRange(
			strconv.FormatUint(uint64(r.Start), 10)+"-"+
				strconv.FormatUint(uint64(r.End), 10),
			"*",
		)
		m.setContentLength(r.End - r.Start + 1)
	}
	return err
}

func (m HeaderMetadata) WantType() jps.MetadataWanted {
	return jps.MetadataHasEffect | jps.MetadataRecomended | jps.MetadataNoDefault
}

func (m HeaderMetadata) Type() (*mediatype.Type, error) {
	types := m.H.GetValues("Content-Type")
	if len(types) == 0 {
		return nil, nil
	}
	var err error
	if len(types) != 1 {
		err = ErrMultipleContentTypeHeaders
	}
	mt, err := ParseMediaType(types[0])
	return mt, err
}
func (m HeaderMetadata) TypeString() (string, error) {
	types := m.H.GetValues("Content-Type")
	if len(types) == 0 {
		return "", nil
	}
	if len(types) != 1 {
		return types[0], ErrMultipleContentTypeHeaders
	}
	return types[0], nil
}

func (m HeaderMetadata) SetType(mt *mediatype.Type) error {
	str, err := FormatMediaType(mt)
	if err == nil {
		m.H.Set("Content-Type", string(str))
	}
	return err
}
func (m HeaderMetadata) SetTypeString(mt string) error {
	m.H.Set("Content-Type", mt)
	return nil
}

//TODO: Content-Coding?

func (m HeaderMetadata) WantEncodings() jps.MetadataWanted {
	return jps.MetadataHasEffect | jps.MetadataRecomended
}

func (m HeaderMetadata) Encodings() ([]string, error) {
	return m.H.GetValues("Content-Encoding"), nil
}

func (m HeaderMetadata) SetEncodings(codings []string) error {
	m.H.SetValues("Content-Encoding", codings)
	return nil
}

func (m HeaderMetadata) WantLanguages() jps.MetadataWanted {
	return jps.MetadataHasEffect
}

func (m HeaderMetadata) Languages() ([]string, error) {
	return m.H.GetValues("Content-Language"), nil
}

func (m HeaderMetadata) SetLanguages(langs []string) error {
	m.H.SetValues("Content-Language", langs)
	return nil
}

func (m HeaderMetadata) WantLocation() jps.MetadataWanted {
	return jps.MetadataHasEffect
}

func (m HeaderMetadata) Location() (string, error) {
	locations := m.H.GetValues("Content-Location")
	switch len(locations) {
	case 0:
		return "", nil
	case 1:
		return locations[0], nil
	default:
		return locations[0], ErrMultipleContentLocationHeaders
	}
}

func (m HeaderMetadata) SetLocation(location string) error {
	m.H.Set("Content-Location", location)
	return nil
}

func (m HeaderMetadata) WantMTime() jps.MetadataWanted {
	return jps.MetadataHasEffect
}

func (m HeaderMetadata) MTime() (time.Time, error) {
	mtimes := m.H.GetValues("Last-Modified")
	if len(mtimes) == 0 {
		return time.Time{}, nil
	}
	mtime, err := gohttp.ParseTime(mtimes[0])
	if err == nil && len(mtimes) != 1 {
		err = ErrMultipleLastModifiedHeaders
	}
	return mtime, err
}

func (m HeaderMetadata) SetMTime(mtime time.Time) error {
	m.H.Set("Last-Modified", mtime.Format(gohttp.TimeFormat))
	return nil
}

func (m HeaderMetadata) WantETag() jps.MetadataWanted {
	return jps.MetadataHasEffect
}

//TODO: Implement strong

func (m HeaderMetadata) ETag() (string, bool, error) {
	return m.H.Get("etag"), true, nil
}

func (m HeaderMetadata) SetETag(etag string, strong bool) error {
	m.H.Set("etag", etag)
	return nil
}
