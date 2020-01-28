package http

import (
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor uint16
}

func (v Version) Protocol() string {
	return "http"
}

func (v Version) Format() string {
	return "HTTP/" + strconv.FormatUint(uint64(v.Major), 10) + "." + strconv.FormatUint(uint64(v.Minor), 10)
}

func (v Version) String() string {
	return v.Format()
}

func (v Version) Version() interface{} {
	return v
}

//ParseHTTPVersion parses: "HTTP/" DIGIT "." DIGIT
func ParseHTTPVersion(s string) (Version, error) {
	if !strings.HasPrefix(s, "HTTP/") {
		return Version{0, 0}, ErrMalformedHTTPVersionDoesNotStartWithHTTP
	}
	s = s[5:]
	if len(s) == 0 {
		return Version{0, 0}, ErrMalformedHTTPVersionNoVersion
	}
	switch s[0] {
	case '1':
		switch s[1:] {
		case ".1":
			return Version{1, 1}, nil
		case ".0":
			return Version{1, 0}, nil
		}
	case '2':
		if s[1:] == ".0" {
			return Version{2, 0}, nil
		}
	case '3':
		if s[1:] == ".0" {
			return Version{3, 0}, nil
		}
	}
	dot := strings.Index(s, ".")
	if dot < 0 {
		return Version{0, 0}, ErrMalformedHTTPVersionNoDot
	}
	major, err := strconv.ParseUint(s[:dot], 10, 16)
	if err != nil {
		return Version{0, 0}, MakeErrMalformedHTTPVersion(err)
	}
	minor, err := strconv.ParseUint(s[dot+1:], 10, 16)
	if err != nil {
		return Version{uint16(major), 0}, MakeErrMalformedHTTPVersion(err)
	}
	return Version{uint16(major), uint16(minor)}, nil
}
