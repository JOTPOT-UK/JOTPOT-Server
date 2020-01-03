package pipe

import (
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/util"
)

type ReaderID uint
type WriterID uint
type ReaderPipeGenerator struct {
	Generator func(io.Reader) (io.ReadCloser, error)
	id        ReaderID
}
type WriterPipeGenerator struct {
	Generator func(io.Writer) (io.WriteCloser, error)
	id        WriterID
}

var readerRegister ReaderID
var writerRegister WriterID
var ErrGeneratorRegisterFull = "generator register full"

func (g *ReaderPipeGenerator) Register() {
	readerRegister++
	if readerRegister == 0 {
		panic(ErrGeneratorRegisterFull)
	}
	g.id = readerRegister
}
func (g *WriterPipeGenerator) Register() {
	writerRegister++
	if writerRegister == 0 {
		panic(ErrGeneratorRegisterFull)
	}
	g.id = writerRegister
}
func (g *ReaderPipeGenerator) ID() ReaderID { return g.id }
func (g *WriterPipeGenerator) ID() WriterID { return g.id }

type ReaderPipe struct {
	readers []io.ReadCloser
}

//From pipes r through all the pipes generated from the pipe readers. Pipes are in the order they appear in the slice.
func From(r io.ReadCloser, pipes ...[]*ReaderPipeGenerator) (*ReaderPipe, error) {
	c := 1
	for pipesI := range pipes {
		c += len(pipes[pipesI])
	}
	out := &ReaderPipe{make([]io.ReadCloser, c, c)}
	c--
	out.readers[c] = r
	var err error
	for pipesI := range pipes {
		for _, p := range pipes[pipesI] {
			r, err = p.Generator(r)
			if err != nil {
				return nil, err
			}
			c--
			out.readers[c] = r
		}
	}
	return out, nil
}

func (r *ReaderPipe) Read(buf []byte) (int, error) {
	return r.readers[0].Read(buf)
}

func (r *ReaderPipe) Close() (err error) {
	//TODO: Multiple close errors?
	if len(r.readers) == 0 {
		return
	}
	i := 0
	for {
		err = r.readers[i].Close()
		i++
		//TODO: -1 to not close connection?
		if i == len(r.readers) || err != nil {
			return
		}
	}
}

type WriterPipe struct {
	writers []io.WriteCloser
}

func To(w io.WriteCloser, pipes ...[]*WriterPipeGenerator) (*WriterPipe, error) {
	c := 1
	for pipesI := range pipes {
		c += len(pipes[pipesI])
	}
	out := &WriterPipe{make([]io.WriteCloser, c, c)}
	c--
	out.writers[c] = w
	var err error
	for pipesI := range pipes {
		for _, p := range pipes[pipesI] {
			w, err = p.Generator(w)
			if err != nil {
				return nil, err
			}
			c--
			out.writers[c] = w
		}
	}
	return out, nil
}

func (w *WriterPipe) Write(buf []byte) (int, error) {
	return w.writers[0].Write(buf)
}

func (w *WriterPipe) Close() (err error) {
	if len(w.writers) == 0 {
		return
	}
	i := 0
	for {
		err = w.writers[i].Close()
		i++
		if i == len(w.writers) || err != nil {
			return
		}
	}
}

func (w *WriterPipe) Flush() (err error) {
	for i := range w.writers {
		if flusher, ok := w.writers[i].(util.WriteFlushCloser); ok {
			err = flusher.Flush()
			if err != nil {
				return
			}
		}
	}
	return
}
