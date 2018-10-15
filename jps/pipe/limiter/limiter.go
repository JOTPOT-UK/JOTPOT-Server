package limiter

import (
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps/pipe"
)

func GenerateLimiterGenerator(n int64) pipe.ReaderPipeGenerator {
	return pipe.ReaderPipeGenerator{
		Generator: func(r io.Reader) (io.ReadCloser, error) {
			return &pipe.ReadCloserReader{io.LimitReader(r, n)}, nil
		},
	}
}
