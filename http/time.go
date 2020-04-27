package http

import (
	gohttp "net/http"
	"time"
)

//TODO: Remove duplicate Format/ParseDate
func ParseTime(text string) (time.Time, error) {
	return gohttp.ParseTime(text)
}

func FormatTime(t time.Time) string {
	return t.Format(gohttp.TimeFormat)
}
