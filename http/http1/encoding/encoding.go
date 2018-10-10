package encoding

import (
	"jotpot/net/jps/pipe"
)

//Encoding represents an encoding - It has a writer and a reader generator,
// along with a name, which is what should be used in the header.
type Encoding struct {
	Name   string
	Reader pipe.ReaderPipeGenerator
	Writer pipe.WriterPipeGenerator
}

//HTTPTransferEncodings returns a List, which contains:
// DeflateEncoding
// GzipEncoding
// XGzipEncoding
// ChunkedEncoding
func HTTPTransferEncodings() List {
	l := NewList(4)
	l.Add(&DeflateEncoding)
	l.Add(&GzipEncoding)
	l.Add(&XGzipEncoding)
	l.Add(&ChunkedEncoding)
	return l
}

//HTTPContentEncodings returns a List, contains:
// DeflateEncoding
// GzipEncoding
// XGzipEncoding
func HTTPContentEncodings() List {
	l := NewList(3)
	l.Add(&DeflateEncoding)
	l.Add(&GzipEncoding)
	l.Add(&XGzipEncoding)
	return l
}
