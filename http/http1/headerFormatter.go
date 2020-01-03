package http1

import (
	"io"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/util"
)

func LengthHeuristic(h *header.Header) int {
	return len(h.Values) * 256
}

func WriteHeadersNoError(w util.StringWriter, h *header.Header) error {
	return h.ForEach(func(key string, values []string) error {
		w.WriteString(key)
		w.Write([]byte{':', ' '})
		i := 0
		for ; i < len(values)-1; i++ {
			w.WriteString(values[i])
			w.Write([]byte{',', ' '})
		}
		w.WriteString(values[i])
		w.Write([]byte{'\r', '\n'})
		return nil
	}, false)
}

func WriteHeaders(w io.Writer, h *header.Header) error {
	return h.ForEach(func(key string, values []string) (err error) {
		_, err = io.WriteString(w, key)
		if err != nil {
			return
		}
		_, err = w.Write([]byte{':', ' '})
		if err != nil {
			return
		}
		i := 0
		for ; i < len(values)-1; i++ {
			_, err = io.WriteString(w, values[i])
			if err != nil {
				return
			}
			_, err = w.Write([]byte{',', ' '})
			if err != nil {
				return
			}
		}
		_, err = io.WriteString(w, values[i])
		if err != nil {
			return
		}
		_, err = w.Write([]byte{'\r', '\n'})
		return
	}, false)
}

func FormatHeaders(h *header.Header) []byte {
	s := make([]byte, 0, LengthHeuristic(h))
	h.ForEach(func(key string, values []string) error {
		s = append(s, key...)
		s = append(s, ':', ' ')
		i := 0
		for lm1 := len(values) - 1; i < lm1; i++ {
			s = append(s, values[i]...)
			s = append(s, ',', ' ')
		}
		s = append(s, values[i]...)
		s = append(s, '\r', '\n')
		return nil
	}, false)
	return s
}
