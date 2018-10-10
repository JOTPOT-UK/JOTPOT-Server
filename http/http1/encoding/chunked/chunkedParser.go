package chunked

import (
	"errors"
	"fmt"
	"io"
	"jotpot/net/util"
	"net"
	"strconv"
	"time"
	"unsafe"
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

func (r *Reader) readNextExtensionMaybe(buf []byte, exts [][2][]byte) ([][2][]byte, error) {
	buf = buf[:2]
	err := r.read2(buf)
	if err != nil {
		return nil, err
	}
	//buf will contain 1 of:
	// ;x
	// \r\n
	if buf[0] == ';' {
		return r.readNextExtension(true, buf[1], buf, exts)
	}
	if buf[0] != '\r' || buf[1] != '\n' {
		return exts, ErrMalformedChunks
	}
	return exts, nil
}

func (r *Reader) readNextExtension(useStartWith bool, startWith byte, buf []byte, exts [][2][]byte) ([][2][]byte, error) {
	//Create a buffer to store the extnesion in, and start it with startWith maybe?
	ext := make([]byte, 0, 16)
	if useStartWith {
		ext = append(ext, startWith)
	}
	//If we suddenly hit the end, then it will be "\r\n", so we can't risk reading more than 2 bytes.
	buf = buf[:2]
	var err error

	//Read the name
	err = r.read2(buf)
	if err != nil {
		return exts, err
	}
	//buf will contain 1 of:
	// 1) x=
	// 2) x\r
	// 3) x;
	// 4) xx
	if buf[1] == '=' {
		return append(exts, [2][]byte{
			append(ext, buf[0]), //Append the character
			/*READ*/ nil,
		}), nil
	} else if buf[1] == '\r' {
		//Read and check the \n
		err = r.read1(buf[1:2])
		if err != nil {
			return exts, err
		}
		if buf[1] != '\n' {
			return exts, ErrMalformedChunks
		}
		return append(exts, [2][]byte{
			append(ext, buf[0]), //Append the character
			nil,                 //No data
		}), nil
	} else if buf[1] == ';' {
		return r.readNextExtension(
			false, 0, buf,
			append(exts, [2][]byte{
				append(ext, buf[0]), //Append the character
				nil,                 //No data
			}),
		)
	} else {
		for {
			ext = append(ext, buf...)
			err = r.read2(buf)
			if err != nil {
				return exts, err
			}
			//buf will now contain 1 of:
			// 1) =d
			// 2) ;x
			// 3) \r\n
			// 4) x=
			// 5) x;
			// 6) x\r
			// 7) xx
			if buf[0] == '=' {

			} else if buf[0] == ';' {
				return r.readNextExtension(
					true, buf[1], buf,
					append(exts, [2][]byte{
						ext,
						nil, //No data
					}),
				)
			} else if buf[0] == '\r' && buf[1] == '\n' {
				return append(exts, [2][]byte{
					ext, //Append the character
					nil, //No data
				}), nil
			} else if buf[1] == '=' {

			} else if buf[1] == ';' {
				return r.readNextExtension(
					false, 0, buf,
					append(exts, [2][]byte{
						append(ext, buf[0]),
						nil,
					}),
				)
			} else if buf[1] == '\r' {

			}
		}
	}
}

func (r *Reader) readBytesLeft() (err error) {
	//For commenting, bytes that are part of the length will be refered to as n,
	// and bytes that are part of the extension will be referd to as x.
	buf := r.buf[2:5]
	var ext [2][]byte
	//Read 3 bytes...
	err = r.fill(buf)
	if err != nil {
		return
	}
	//The buf with be one of:
	// 1) n\r\n
	// 2) n;x
	// 3) nn\r
	// 4) nn;
	// 5) nn
	if buf[1] == '\r' {
		//Case 1
		if buf[2] != '\n' {
			return ErrMalformedChunks
		}
		buf = buf[:1]
	} else if buf[1] == ';' {
		//Case 2
		ext, err = r.readExtension(true, buf[2])
		if err != nil {
			return
		}
		buf = buf[:1]
	} else if buf[2] == '\r' {
		//Case 3
		//Read and check \n
		err = r.read1(buf[2:3])
		if err != nil {
			return
		}
		if buf[2] != '\n' {
			return ErrMalformedChunks
		}
		buf = buf[:2]
	} else if buf[2] == ';' {
		//Case 4
		ext, err = r.readExtension(false, 0)
		if err != nil {
			return
		}
		buf = buf[:2]
	} else {
		//Case 5, so, read more... Until we are at an end...
		for {
			buf2 := r.buf[:2]
			err = r.read2(buf2)
			if err != nil {
				return
			}
			//buf2 will contain 1 of:
			// 1) \r\n
			// 2) ;x
			// 3) n\r
			// 4) n;
			// 5) nn
			if buf2[0] == '\r' {
				//Case 1
				if buf2[1] != '\n' {
					return ErrMalformedChunks
				}
				break
			} else if buf2[0] == ';' {
				//Case 2
				ext, err = r.readExtension(true, buf2[1])
				if err != nil {
					return
				}
				break
			} else if buf2[1] == '\r' {
				//Case 3
				//Append the character
				buf = append(buf, buf2[0])
				//Read and check the \n
				err = r.read1(buf2[1:2])
				if err != nil {
					return
				}
				if buf2[1] != '\n' {
					return ErrMalformedChunks
				}
				break
			} else if buf2[1] == ';' {
				//Case 4
				//Append the character
				buf = append(buf, buf2[0])

				ext, err = r.readExtension(false, 0)
				if err != nil {
					return
				}
				break
			} else {
				//Case 5 - Just append and carry on
				buf = append(buf, buf2...)
			}
		}
	}
	var temp uint64
	//Since we know that the buffer won't change, we can convert it using unsafe, as this won't copy it...
	temp, err = strconv.ParseUint(*(*string)(unsafe.Pointer(&buf)), 16, 31)
	r.bytesLeft = int(temp)
	fmt.Println("ext:", ext)
	return
}

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
	fmt.Println("Reading")
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
			//Consume the CRLF after the 0 sized chunk
			err = r.ConsumeCRLF()
			//And if there wasn't an error doing that, then the error is EOF
			if err == nil {
				err = io.EOF
			}
			return 0, err
		}
	}
	fmt.Println(r.bytesLeft, "bytes left")
	//Read up to how many bytes are in this chunk, and substract how many we read.
	n, err = r.readSource(buf[:util.MinInt(r.bytesLeft, len(buf))])
	fmt.Println("Read", n, "bytes")
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
