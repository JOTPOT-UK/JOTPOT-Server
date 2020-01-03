package fileserver

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path"

	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

var ErrRequestHasNoURL = errors.New("request has no url")

type FileServer struct {
	GetRoot func(url *url.URL) (string, error)
}

func (fs *FileServer) Open(url *url.URL) (string, *os.File, os.FileInfo, error) {
	root, err := fs.GetRoot(url)
	if err != nil {
		return "", nil, nil, err
	}
	p := path.Join(root, url.Path)
	f, err := os.OpenFile(p, os.O_RDONLY, 0)
	if err != nil {
		return "", nil, nil, err
	}
	stats, err := f.Stat()
	if err != nil {
		return "", nil, nil, err
	}
	if stats.IsDir() {
		p = path.Join(p, "index.html")
		f, err = os.OpenFile(p, os.O_RDONLY, 0)
		if err != nil {
			return "", nil, nil, err
		}
		stats, err = f.Stat()
		if err != nil {
			return "", nil, nil, err
		}
		if stats.IsDir() {
			return "", nil, nil, os.ErrNotExist
		}
	}
	return p, f, stats, nil
}

func write(dst io.Writer, src io.Reader, maxBufSize, size int64) error {
	//TODO: Make sure size fits in to int
	if maxBufSize < size {
		size = maxBufSize
	}
	buf := make([]byte, size, size)
	fmt.Println("Buf size:", size)
	for {
		rn, rerr := src.Read(buf)
		if rn == 0 {
			if rerr != nil {
				if rerr == io.EOF {
					return nil
				}
				return rerr
			}
			continue // TODO: This is dangerous... Perhaps use io.ErrNoProgress?
		}
		wn, werr := dst.Write(buf[:rn])
		if werr != nil {
			return rerr
		}
		if rerr != nil {
			if rerr == io.EOF {
				return nil
			}
			return rerr
		}
		if wn != rn {
			return io.ErrShortWrite
		}
	}
}

func (fs *FileServer) Handle(frame *jps.ServerFrame) (bool, error) {
	req := frame.Request
	resp := frame.Response
	url := req.URL()
	if url == nil {
		return false, ErrRequestHasNoURL
	}
	filePath, f, stats, err := fs.Open(req.URL())
	if err != nil {
		if os.IsNotExist(err) {
			resp.SetStatus(jps.ResponseStatusNotFound)
			resp.Body.StringBody("404: Not found!")
			return true, nil
		}
		return true, err
	}
	size := stats.Size()
	resp.Body.SetSize(size)
	resp.Body.SetTypeString(mime.TypeByExtension(path.Ext(filePath)))
	resp.Body.SetMTime(stats.ModTime())
	err = write(&resp.Body, f, 16384, size)
	if err != nil {
		return false, err
	}
	resp.Body.Close()
	return true, err
}
