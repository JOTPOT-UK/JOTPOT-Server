package http1

import (
	"io"
	"strconv"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
)

//TODO: Write Host header first in request formatting.

func WriteResponse(w io.Writer, resp *http.Response) error {
	_, err := w.Write(FormatResponse(resp))
	return err
}

func FormatResponse(resp *http.Response) []byte {
	statusTxt := resp.StatusText
	if statusTxt == "" {
		statusTxt = http.StatusText(int(resp.StatusCode))
	}
	out := make([]byte, 0, 15+len(statusTxt)+LengthHeuristic(resp.Header))
	out = append(out, resp.Version.Format()...)
	out = append(out, ' ')
	out = append(out, strconv.FormatUint(uint64(resp.StatusCode), 10)...)
	out = append(out, ' ')
	out = append(out, statusTxt...)
	out = append(out, '\r', '\n')
	out = append(out, FormatHeaders(resp.Header)...)
	return append(out, '\r', '\n')
}
