package jps

import "errors"

var ErrNotSupported = errors.New("not supported")
var ErrResponseStatusNotSupported = errors.New("response status not supported")
var ErrMethodNotSupported = errors.New("method not supported")

//IgnoreNotSupported either returns the error passed or returns nil if errors.Is(err, ErrNotSupported).
//This should be used to wrap errors
func IgnoreNotSupported(err error) error {
	if errors.Is(err, ErrNotSupported) {
		return nil
	}
	return err
}
