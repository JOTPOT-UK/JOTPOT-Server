package chunked

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"
	"unsafe"

	"github.com/JOTPOT-UK/JOTPOT-Server/util"
)

//IsTemp returns true if the error us a temporary network error
func IsTemp(err error) bool {
	ne, ok := err.(net.Error)
	return ok && ne.Temporary()
}

//ErrMalformedChunks will be returned if the raw data stream is invalid encoding
var ErrMalformedChunks = errors.New("malformed chunks")

//ZeroReadSleep is the duration of which the parser will sleep after readSource returns with a read of 0.
// If this happens, there will have already been a wait of about a minute, so this may be long.
var ZeroReadSleep = 2 * time.Second

//Reader provides a way of reading data streamed using chunked encoding
type Reader struct {
	source    io.Reader
	bytesLeft int
	buf       [5]byte
}

//Close does absolutially nothing! And returns a nil error!
func (r *Reader) Close() error {
	return nil
}

//readSource reads data from the source into the buf. If an error is temporary, it will wait.
func (r *Reader) readSource(buf []byte) (n int, err error) {
	timeout := time.Millisecond
	var waits uint8
	for {
		//Read
		n, err = r.source.Read(buf)

		if err == nil {
			//If there wasn't an error, then return if we read something. Otherwise, wait...
			if n != 0 {
				return
			}
		} else if IsTemp(err) {
			//If there was an error, and it was temp...
			//If we did read something, then just give them that, so that hopefully the error will be gone next time.
			//Otherwise, let's wait
			if n != 0 {
				return n, nil
			}
		} else {
			//Return because there was a perminant error!!!!!!!!! :(
			return
		}

		//We want to wait for timeout
		time.Sleep(timeout)
		//If we are waiting for more than 250ms, then don't increase the timeout, but do count it.
		if timeout > time.Second/4 {
			//If we've waited for 255 times, then we can return with the error :(
			if waits == 255 {
				return
			}
			waits++
		} else {
			//Double the timeout for next time
			timeout *= 2
		}
	}
}

func (r *Reader) read1(buf []byte) (err error) {
	var n int
	for {
		n, err = r.readSource(buf)
		if n == 1 || err != nil {
			return
		}
		time.Sleep(ZeroReadSleep)
	}
}

func (r *Reader) read2(buf []byte) (err error) {
	var n int
	for {
		//Let's read! And return if there was an error
		n, err = r.readSource(buf)
		if n == 2 || err != nil {
			return
		} else if n == 1 {
			//If we've read only 1 byte, then make sure we read 1 more!
			return r.read1(buf[1:])
		} else if n == 0 {
			//If we didn't read anything, then can we please try again after a sleep?
			time.Sleep(ZeroReadSleep)
		}
	}
}

func (r *Reader) fill(buf []byte) (err error) {
	var n, c int
	for {
		n, err = r.readSource(buf[c:])
		c += n
		//Return if we have filled the buffer, or there was an error
		if c == len(buf) || err != nil {
			return
		} else if n == 0 {
			//If we didn't read anything, then can we please try again after a sleep?
			time.Sleep(ZeroReadSleep)
		}
	}
}

func (r *Reader) readLine() (line []byte, err error) {
	//First, there is a good chance that we are actually a bufio.Reader... So that would make this more efficient!
	bufReader, ok := r.source.(*bufio.Reader)
	if ok {
		//TODO: What about a temp error?
		line, err = bufReader.ReadBytes('\r')
		if err != nil {
			return
		}
		line[len(line)-1], err = bufReader.ReadByte()
		if err != nil {
			return
		}
		if line[len(line)-1] != '\n' {
			err = ErrMalformedChunks
		}
		return line[:len(line)-1], err
	}

	buf := r.buf[:2]
	line = make([]byte, 0, 4)
	for {
		err = r.read2(buf)
		if err != nil {
			return
		}
		if buf[0] == '\r' {
			if buf[1] != '\n' {
				return line, ErrMalformedChunks
			}
			return
		} else if buf[1] == '\r' {
			err = r.read1(buf[1:2])
			if err != nil {
				return
			}
			if buf[1] != '\n' {
				return line, ErrMalformedChunks
			}
			return append(line, buf[0]), nil
		} else {
			line = append(line, buf...)
		}
	}
}

func (r *Reader) readBytesLeft() error {
	var size []byte
	line, err := r.readLine()
	if err != nil {
		return err
	}
	i := bytes.IndexByte(line, ';')
	if i == -1 {
		size = line
	} else {
		size = line[:i]
		fmt.Println("exts: ", line[i+1:])
	}
	var temp uint64
	//Since we know that the buffer won't change, we can convert it using unsafe, as this won't copy it...
	temp, err = strconv.ParseUint(*(*string)(unsafe.Pointer(&size)), 16, 31)
	r.bytesLeft = int(temp)
	return err
}

//ConsumeCRLF reads 2 bytes from the source, and returns ErrMalformedChunks if they are not "\r\n".
func (r *Reader) ConsumeCRLF() (err error) {
	err = r.read2(r.buf[:2])
	if err != nil {
		return
	}
	if r.buf[0] != '\r' || r.buf[1] != '\n' {
		return ErrMalformedChunks
	}
	return
}

func (r *Reader) Read(buf []byte) (n int, err error) {
	//If it's negative, then we should return EOF
	if r.bytesLeft < 0 {
		return 0, io.EOF
	} else if r.bytesLeft == 0 {
		//We've consumed a chunk, so lets find out how many bytes are in the next one
		err = r.readBytesLeft()
		//If there was an error, set the length to -1, so that we give EOF next call, and return with the error.
		if err != nil {
			r.bytesLeft = -1
			return
		}
		//If there is a length of 0, then its the end!
		if r.bytesLeft == 0 {
			r.bytesLeft = -1
			//TODO: Trailers
			//Consume the CRLF after the 0 sized chunk
			err = r.ConsumeCRLF()
			//And if there wasn't an error doing that, then the error is EOF
			if err == nil {
				err = io.EOF
			}
			return 0, err
		}
	}
	//Read up to how many bytes are in this chunk, and substract how many we read.
	n, err = r.readSource(buf[:util.MinInt(r.bytesLeft, len(buf))])
	//EOF and then return if there was an error
	if err != nil {
		r.bytesLeft = -1
		return
	}
	r.bytesLeft -= n
	//If there are still some bytes left to read,
	// then we have filled the buffer, or exhaused the stream for now,
	// so return.
	if r.bytesLeft != 0 {
		return
	}
	//Because it's the end of the chunk, we need to consume the CRLF
	err = r.ConsumeCRLF()
	if err != nil {
		return
	}
	//At this point, we've ran out of bytes left in this chunk,
	// so we should call read again to start on the next chunk.
	n2, err := r.Read(buf[n:])
	//Add how much we've read to n, and then return.
	n += n2
	return
}

//NewPipe creates a new pipe that reads data with a chunked transfer encoding and returns just the raw data.
func NewPipe(source io.Reader) (io.ReadCloser, error) {
	return &Reader{
		source: source,
	}, nil
}
