package cache

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"
	"github.com/JOTPOT-UK/JOTPOT-Server/mediatype"
)

type BytesCache interface {
	Get(string) []byte
	Put(string, []byte)
}

//FileReader is an interface to a  which you can read files from.
type FileReader interface {
	//OpenRead returns an IncomingBody which reads from `path` through the range, `r`.
	OpenRead(path string, r jps.Range) (*jps.IncomingBody, error)
}

//FileWriter is an interface to a  which you can create, write and remove files in.
type FileWriter interface {
	//TODO: Create?
	//OpenRead returns a body which writes to the range, `r`, in `path`.
	OpenWrite(path string, r jps.Range) (*jps.OutgoingBody, error)
	//Remove deletes a file at `path`.
	Rmfile(path string) error
}

//FileReadWriter is an interface to a  which you can create, write, read and remove files in.
//It notably doesn't have to have a directory structure - see DirReadWriter for a directory interface.
type FileReadWriter interface {
	FileReader
	FileWriter
}

//StatReader is an interface to a  in which you can read metadata about files from.
type StatReader interface {
	//Stat returns metadata for file at `path`.
	Stat(path string) (jps.MetadataGetter, error)
}

//StatWriter is an interface to a  in which you can set metadata for files.
type StatWriter interface {
	Chown(path, user, group string) error
	Chmod(path string, permissions uint64) error //TODO: Change interface
	Chtimes(path string, CTime, MTime, ATime time.Time) error
}

type StatReadWriter interface {
	StatReader
	StatWriter
}

type DirReader interface {
	//Readdir returns a list of enteries in the directory at `path`.
	Readdir(path string) ([]string, error)
}

type DirWriter interface {
	//Mkdir creates a directory at `path`.
	Mkdir(path string) error
	//Rmdir removes a directory at `path`.
	Rmdir(path string) error
	//Move moves or renames `oldpath` to `newpath`.
	Move(oldpath, newpath string) error

	//TODO: MkdirAll?
	//MkdirAll(path string) error
}

type DirReadWriter interface {
	DirReader
	DirWriter
}

type Remover interface {
	Remove(path string) error
	RemoveAll(path string) error
}

//LinkReader is an interface to a  in which you can read links.
type LinkReader interface {
	//Readlink returns the path which `path` links to or "" if it is not a link.
	Readlink(path string) (string, error)
}

//LinkWriter is an interface to a  in which you can create links.
type LinkWriter interface {
	//Link creates a link named `LinkName` to an existing `Dest`.
	Link(Dest, LinkName string) error
	//TODO: Rmlink?
}

//LinkReadWriter is an interface to a  in which you can set and read links.
type LinkReadWriter interface {
	LinkReader
	LinkWriter
}

type ReadOnly interface {
	FileReader
	DirReader
	StatReader
	LinkReader
}

type Filesystem interface {
	FileReadWriter
	DirReadWriter
	StatReadWriter
	LinkReadWriter
	Remover
}

type LimitReadCloser struct {
	R    io.ReadCloser
	i, j int64
}

func NewLimitReadCloser(r io.ReadCloser, limit int64) io.ReadCloser {
	return &LimitReadCloser{
		R: r,
		j: limit,
	}
}

func (r *LimitReadCloser) Read(p []byte) (int, error) {
	pl := int64(len(p))
	left := r.j - r.i
	if left == 0 {
		return 0, io.EOF
	}
	if left < pl {
		p = p[:left]
	}
	n, err := r.R.Read(p)
	r.i += int64(n)
	return n, err
}

func (r *LimitReadCloser) Close() error {
	return r.R.Close()
}

type OSFile struct {
	F      *os.File
	stats  os.FileInfo
	fRange jps.Range
	len    int64
	reader io.ReadCloser
}

func NewOSFile(F *os.File, r jps.Range) *OSFile {
	return &OSFile{
		F:      F,
		fRange: r,
		len:    -1,
	}
}

func (f *OSFile) setReader() (err error) {
	r := f.fRange
	if r.End == -1 {
		if r.Start > 0 {
			_, err = f.F.Seek(r.Start, 0)
		} else if r.Start < 0 {
			_, err = f.F.Seek(r.Start, 2)
		}
		f.reader = f.F
	} else {
		l, lSizeMul := r.Length()
		if lSizeMul != 0 {
			var size int64
			size, err = f.Size()
			if err != nil {
				return err
			}
			r = r.Abs(size)
			l, _ = r.Length()
			_, err = f.F.Seek(r.Start, 0)
		} else if r.Start > 0 {
			_, err = f.F.Seek(r.Start, 0)
		} else if r.Start < 0 {
			_, err = f.F.Seek(r.Start, 2)
		}
		f.len = l
		f.reader = NewLimitReadCloser(f.F, l)
	}
	return
}

func (f *OSFile) stat() (os.FileInfo, error) {
	var err error
	if f.stats == nil {
		f.stats, err = f.F.Stat()
	}
	return f.stats, err
}

func (f *OSFile) Size() (int64, error) {
	stats, err := f.stat()
	if err != nil {
		return -1, err
	}
	return stats.Size(), nil
}

func (f *OSFile) BodyLength() (int64, error) {
	var err error
	if f.len < 0 {
		f.len, err = f.Size()
	}
	return f.len, err
}

func (f *OSFile) Body() (io.ReadCloser, error) {
	var err error
	if f.reader == nil {
		err = f.setReader()
	}
	return f.reader, err
}

func (_ OSFile) Session() jps.Session {
	return nil // TODO: OSFileSession?
}

func (_ OSFile) Location() (string, error) {
	return "", nil
}

func (f *OSFile) Range() (jps.Range, error) {
	return f.fRange, nil
}

func (_ OSFile) Type() (*mediatype.Type, error) {
	return nil, nil
}

func (_ OSFile) TypeString() (string, error) {
	return "", nil
}

func (_ OSFile) Encodings() ([]string, error) {
	return nil, nil
}

func (_ OSFile) Languages() ([]string, error) {
	return nil, nil
}

func (f *OSFile) MTime() (time.Time, error) {
	stats, err := f.stat()
	return stats.ModTime(), err
}

func (_ OSFile) ETag() (string, bool, error) {
	return "", false, nil
}

func OpenRead(path string, r jps.Range) (*OSFileResponse, *jps.IncomingBody) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	osf := NewOSFile(f, r)
	status := jps.ResponseStatusOK
	if err != nil {
		if os.IsNotExist(err) {
			status = jps.ResponseStatusNotFound
		} else if os.IsPermission(err) {
			status = jps.ResponseStatusForbidden
		} else {
			status = jps.ResponseStatusInternalServerError
		}
	}
	return &OSFileResponse{
			status: status,
		}, &jps.IncomingBody{
			Reader:         jps.NewReader(osf),
			MetadataGetter: osf,
		}
}

type OSSession struct {
	jps.ConnectionDetails
	ended bool

	Vars    jpvariable.Variables
	BSource jps.BufioSource
}

func (s *OSSession) End() error {
	s.ended = true
	return nil
}

func (s *OSSession) Variables() jpvariable.Variables {
	return s.Vars
}

func (s *OSSession) BufioSource() jps.BufioSource {
	return s.BSource
}

func (_ *OSSession) Hyjack() (jps.Stream, error) {
	return nil, jps.ErrNotSupported
}

func (s *OSSession) Frame() (*jps.ClientFrame, error) {
	return &jps.ClientFrame{
		Request: jps.OutgoingRequest{
			Request: &OSFileRequest{},
		},
		Session: s,
	}
}

type OSFileRequest struct {
	_URL    *url.URL
	_Method jps.Method
	Range   jps.Range
}

func (r *OSFileRequest) URL() *url.URL {
	return r._URL
}

func (_ *OSFileRequest) methodSupported(method jps.Method) bool {
	switch method {
	case jps.MethodStat, jps.MethodRead, jps.MethodReadDir, jps.MethodReadFile, jps.MethodReadLink, jps.MethodMkdir, jps.MethodMkfile, jps.MethodWrite, jps.MethodMove, jps.MethodRemove, jps.MethodRemoveAll, jps.MethodLink, jps.MethodChown, jps.MethodChmod, jps.MethodChtimes:
		return true
	default:
		return false
	}
}

func (r *OSFileRequest) Method() jps.Method {
	return r._Method
}

func (r *OSFileRequest) SetMethod(method jps.Method) error {
	if r.methodSupported(method) {
		r._Method = method
		return nil
	}
	return jps.ErrNotSupported
}

func (r *OSFileRequest) Ranges() ([]jps.Range, error) {
	return []jps.Range{r.Range}, nil
}

func (r *OSFileRequest) SetRanges(ranges []jps.Range) error {
	if len(ranges) < 1 {
		//TODO: Decide on behaviour
		r.Range = jps.NewRange(0, -1)
	} else if len(ranges) == 1 {
		r.Range = ranges[0]
	} else {
		r.Range = ranges[0]
		return fmt.Errorf("multiple ranges %w for OSFile", jps.ErrNotSupported)
	}
	return nil
}

type OSFileResponse struct {
	Req *OSFileRequest

	status jps.ResponseStatus
}

func (r OSFileResponse) Status() (jps.ResponseStatus, bool) {
	return r.status, true
}
func (r *OSFileResponse) SetStatus(status jps.ResponseStatus) bool {
	r.status = status
	return true
}
