package encoding

import (
	"compress/gzip"
	"compress/lzw"
	"compress/zlib"
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1/encoding/chunked"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
)

//ChunkedEncoding is the Encoding for HTTP chunked encoding.
var ChunkedEncoding = Encoding{
	Name:   "chunked",
	Reader: pipe.ReaderPipeGenerator{Generator: chunked.NewPipe},
}

//NewLzwReader calls lzw.NewReader with order as x and litWidth as y.
func NewLzwReader(r io.Reader) (io.ReadCloser, error) {
	return lzw.NewReader(r, 0, 0), nil
}

//NewLzwWriter calls lzw.NewWriter with order as x and litWidth as y.
func NewLzwWriter(r io.Writer) (io.WriteCloser, error) {
	return lzw.NewWriter(r, 0, 0), nil
}

//CompressEncoding is the HTTP "compress" encoding
var CompressEncoding = Encoding{
	Name:   "compress",
	Reader: pipe.ReaderPipeGenerator{Generator: NewLzwReader},
	Writer: pipe.WriterPipeGenerator{Generator: NewLzwWriter},
}

//XCompressEncoding is identicle to CompressEncoding, as per the HTTP Spec
var XCompressEncoding = Encoding{
	Name:   "x-compress",
	Reader: pipe.ReaderPipeGenerator{Generator: NewLzwReader},
	Writer: pipe.WriterPipeGenerator{Generator: NewLzwWriter},
}

//DeflateEncoding implements zlib compression, which is the "deflate" HTTP transfer encoding.
var DeflateEncoding = Encoding{
	Name:   "deflate",
	Reader: pipe.ReaderPipeGenerator{Generator: zlib.NewReader},
	Writer: pipe.WriterPipeGenerator{Generator: func(w io.Writer) (io.WriteCloser, error) {
		return zlib.NewWriter(w), nil
	}},
}

//NewGzipReader calls gzip.NewReader, but returns the result as an interface
func NewGzipReader(r io.Reader) (io.ReadCloser, error) {
	return gzip.NewReader(r)
}

//NewGzipWriter calls gzip.NewWriter, but returns the result as an interface
func NewGzipWriter(w io.Writer) io.WriteCloser {
	return gzip.NewWriter(w)
}

//GzipEncoding implements gzip compression
var GzipEncoding = Encoding{
	Name:   "gzip",
	Reader: pipe.ReaderPipeGenerator{Generator: NewGzipReader},
	Writer: pipe.WriterPipeGenerator{Generator: func(w io.Writer) (io.WriteCloser, error) {
		return NewGzipWriter(w), nil
	}},
}

//XGzipEncoding is identicle to GzipEncoding, apart from the name
var XGzipEncoding = Encoding{
	Name:   "x-gzip",
	Reader: pipe.ReaderPipeGenerator{Generator: NewGzipReader},
	Writer: pipe.WriterPipeGenerator{Generator: func(w io.Writer) (io.WriteCloser, error) {
		return NewGzipWriter(w), nil
	}},
}
