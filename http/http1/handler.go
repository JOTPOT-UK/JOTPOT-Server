package http1

import (
	"bufio"
	"errors"
	"io"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"

	"github.com/JOTPOT-UK/JOTPOT-Server/util"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

type HTTPErrorHandler func(sess *Session, err error)

type Session struct {
	server    *jps.Server
	variables jpvariable.Variables
	config    *http.Config
	stream    jps.Stream
	reader    *bufio.Reader
	//close should be set to true when the stream should be closed and the session ended after the current frame is completed.
	close    bool
	eos      bool
	hyjacked bool
	//conLock is to ensure that only 1 frame is using the stream at a time.
	//Due to the nature of HTTP/1, this should be locked when frame is called, and unlocked when the response has ended.
	conLock sync.Mutex
}

func (s *Session) Server() *jps.Server {
	return s.server
}
func (s *Session) Variables() jpvariable.Variables {
	return s.variables
}
func (s *Session) BufioSource() jps.BufioSource {
	return s.server.BufioSource
}
func (s *Session) End() error {
	s.close = true
	return nil
}
func (s *Session) Protocol() jps.Protocol {
	return SessionProtocol{}
}
func (s *Session) RemoteAddr() net.Addr {
	return s.stream.RemoteAddr()
}
func (s *Session) LocalAddr() net.Addr {
	return s.stream.LocalAddr()
}
func (s *Session) Via() ([]jps.Hop, bool) {
	return nil, false
}

func (s *Session) ErrorHandler() *HTTPErrorHandler {
	var rv *HTTPErrorHandler
	if rvInterface, ok := s.variables.Get("http-error-handler"); ok {
		if rv, ok = rvInterface.(*HTTPErrorHandler); !ok {
			rv = nil
		}
	}
	return rv
}
func (s *Session) HandleError(err error) {
	handler := s.ErrorHandler()
	if handler != nil {
		(*handler)(s, err)
	}
}

type SessionProtocol struct{}

func (_ SessionProtocol) Protocol() string {
	return "http"
}
func (_ SessionProtocol) Version() interface{} {
	return nil
}
func (_ SessionProtocol) String() string {
	return "http"
}

func (s *Session) Hyjack() (jps.Stream, error) {
	s.hyjacked = true
	s.close = true
	return &jps.WrappedStream{
		Reader: s.reader,
		Writer: s.stream,
		Closer: s.stream,
		Stream: s.stream,
	}, nil
}

func (s *Session) Frame() (*jps.ServerFrame, error) {
	//Wait for the stream
	s.conLock.Lock()
	//close might have been set to true while we were waiting, so check again.
	if s.close {
		s.conLock.Unlock()
		return nil, jps.EOS
	}
	//TODO: Unlock on close

	//Section 3.5 of RFC7230 says that:
	// a server that is expecting to receive
	// and parse a request-line SHOULD ignore at least one empty line (CRLF)
	// received prior to the request-line
	//We will ignore up to 5, as it would take a REALLY buggy client to go over that!
	for i := byte(0); i < 5; i++ {
		peek, err := s.reader.Peek(2)
		if err != nil {
			s.close = true
			s.conLock.Unlock()
			s.stream.Close()
			return nil, fmt.Errorf("HTTP \\r\\n consumer loop: %w", err)
		}
		if peek[0] == '\r' && peek[1] == '\n' {
			s.reader.Discard(2)
		} else {
			break
		}
	}
	//TIMER:start := time.Now()
	var httpReq *http.Request
	var err error
	var canRespond bool
	httpReq, s.close, err, canRespond = ParseRequest(s.reader, s.config.HeaderProcessor)
	if err != nil {
		s.close = true
		if canRespond {
			s.HandleError(err)
			err = jps.EOS
		} else {
			err = fmt.Errorf("Error while parsing HTTP request: %w", err)
		}
		s.conLock.Unlock()
		s.stream.Close()
		return nil, err
	}
	//TIMER:jps.ParseTimes = append(jps.ParseTimes, time.Since(start))
	//TIMER:start = time.Now()
	bodyReader := NewBodyReader(
		s, s.config,
		httpReq.Header, httpReq,
		util.ReaderClosePassOn{s.reader, s.stream},
	)
	httpResp := http.Response{
		StatusCode: 200,
		Version:    httpReq.Version,
		Header:     header.New(16, s.config.HeaderProcessor),
	}
	httpResp.Header.SetValuesRaw("Date", []string{time.Now().Format(time.RFC822)}) //TODO: Add a custom date function? Not required but optimal...
	bodyWriter := NewBodyWriter(
		s, s.config,
		httpResp.Header, httpReq,
		util.NoopFlusher{s.stream},
		func /*headerGenerator*/ () []byte {
			if httpResp.Header.Has("Connection", []string{"close"}) {
				s.close = true
				s.conLock.Unlock()
			} else if s.close {
				s.conLock.Unlock()
				httpResp.Header.Set("Connection", "close")
			} else {
				httpResp.Header.Set("Connection", "keep-alive")
			}
			return FormatResponse(&httpResp)
		}, func /*close*/ () error {
			//TODO: Allow multiple closes?
			if s.close && !s.hyjacked {
				return s.stream.Close()
			}
			s.conLock.Unlock()
			return nil
		},
	)
	//TIMER:jps.SetupTimes = append(jps.SetupTimes, time.Since(start))
	return &jps.ServerFrame{
		Session: s,
		Request: jps.IncomingRequest{
			Request: http.RequestRangeWrapper{
				Request:     httpReq,
				RespHeaders: httpResp.Header,
			},
			ConnectionDetails: nil,
			Body: jps.IncomingBody{
				Reader:         jps.NewReader(&bodyReader),
				MetadataGetter: HeaderMetadata{httpReq.Header},
			},
		},
		Response: jps.OutgoingResponse{
			Response:          &httpResp,
			ConnectionDetails: nil,
			Body: jps.OutgoingBody{
				Writer:   jps.NewWriter(&bodyWriter),
				Metadata: HeaderMetadata{httpResp.Header},
			},
		},
		ConnectionDetails: s,
	}, nil
	//s.Handlers.Call(jpsReq)
	//if req.R.Close {
	//	jpsReq.Close()
	//	con.Close()
	//	return nil, nil, nil
	//}
	//TODO: Pragma
}

//Handle takes a server and a connection to that server, and handles it as a HTTP/1.x request.
func Handle(s *jps.Server, con jps.Stream) { //TODO: use bufio reader provided
	/*defer (func() {
		err := recover()
		if err != nil {
			fmt.Println("Panic!!!!")
			fmt.Println(err)
			SendErrorResponse(con, 500, "Internal server error")
		}
	})()*/
	//Get the config from the server variables, or use the default config.
	var config *http.Config
	if configInterface, ok := s.Config.Get("http"); ok {
		if config, ok = configInterface.(*http.Config); !ok {
			config = &http.DefaultConfig
		}
	} else {
		config = &http.DefaultConfig
	}
	//Create and handle the session!
	err := s.HandleSession(&Session{
		server: s,
		//Create a new scope for variables.
		variables: jpvariable.VariablesScope{
			This:   jpvariable.MapVariables{},
			Parent: s.Config,
		},
		config: config,
		stream: con,
		reader: jps.BufioReader(con),
	})
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, err)
	}
}
