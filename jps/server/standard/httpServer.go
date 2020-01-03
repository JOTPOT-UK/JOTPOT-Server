package standard

import (
	"strings"

	gohttp "net/http"

	"github.com/JOTPOT-UK/JOTPOT-Server/jpvariable"

	"github.com/JOTPOT-UK/JOTPOT-Server/http/header/jpsheader"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1"

	"github.com/JOTPOT-UK/JOTPOT-Server/http"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/header"
	"github.com/JOTPOT-UK/JOTPOT-Server/http/http1/encoding"
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

func LowerCaseAll(key string, values []string, start, end int) (string, []string, int, int) {
	for i := start; i < end; i++ {
		values[i] = strings.ToLower(values[i])
	}
	return key, values, start, end
}

func NewDefaultHTTPServer() *jps.Server {
	s := jps.Server{
		Config: jpvariable.MapVariables(map[string]interface{}{
			"http": &http.Config{
				TransferEncodings: encoding.HTTPTransferEncodings(),
				ContentEncodings:  encoding.HTTPContentEncodings(),
				HeaderProcessor: &header.Processor{
					KeyTransforms:    []func(string) string{gohttp.CanonicalHeaderKey, jpsheader.Prefix},
					ValuesProcessors: []func(string, []string, int, int) (string, []string, int, int){LowerCaseAll},
				}},
		}),
		DefaultStreamHandler: http1.Handle,

		BufioSource: &jps.BasicBufioSource{4096, 4096},
	}
	/*err := s.AddStreamSessionHandler(0, 0, &jps.StreamSessionHandler{
		F: func(sess jps.StreamSession) (bool, error) {
			fmt.Println("2")
			con, err := sess.AcceptStream()
			fmt.Println("3")
			if err != nil {
				fmt.Println("4")
				if err == jps.EOS {
					fmt.Println("5a")
					err = nil
				}
				fmt.Println("5b")
				return true, err
			}
			fmt.Println("6")
			s.HandleStream(con, http1.Handle)
			return true, nil
		},
	})
	if err != nil {
		panic(err)
	}*/
	return &s
}
