package http

import (
	"github.com/JOTPOT-UK/JOTPOT-Server/jps"
)

/*type ResponseStatusMap interface {
	ResponseStatus(code uint16) jps.ResponseStatus
	CodeAndText(jps.ResponseStatus) (uint16, string)
	Code(jps.ResponseStatus) uint16
	Text(jps.ResponseStatus) string
}

var DefaultResponseStatusMap = TheMapUnion{
	CodeToStatus: &StatusCodeToStatusMap,
	StatusToCode: &StatusToStatusCodeMap,
}

type TheMapUnion struct {
	CodeToStatus *Uint16SteppedResponseCodeMap
	StatusToCode *jps.ResponseStatusToCodeWithStringMap
}

func (mu *TheMapUnion) ResponseStatus(code uint16) jps.ResponseStatus {
	return mu.CodeToStatus.Get(code)
}

func (mu *TheMapUnion) CodeAndText(status jps.ResponseStatus) (uint16, string) {
	rv := mu.StatusToCode.Get(status)
	return rv.Code, rv.Str
}

func (mu *TheMapUnion) Code(status jps.ResponseStatus) uint16 {
	return mu.StatusToCode.Get(status).Code
}

func (mu *TheMapUnion) Text(status jps.ResponseStatus) string {
	return mu.StatusToCode.Get(status).Str
}*/

type Uint16SteppedResponseCodeMap struct {
	M       [][]jps.ResponseStatus
	Step    uint16
	Default jps.ResponseStatus
}

func (m Uint16SteppedResponseCodeMap) Get(code uint16) jps.ResponseStatus {
	i := code / m.Step
	ii := int(i)
	j := int(code - m.Step*i)
	if ii >= len(m.M) {
		return m.Default
	}
	slice := m.M[ii]
	if j >= len(slice) {
		return m.Default
	}
	return slice[j]
}

//StatusCodeToStatusMap is a map of HTTP status codes to jps.ResponseStatus values.
var StatusCodeToStatusMap = Uint16SteppedResponseCodeMap{
	M: [][]jps.ResponseStatus{
		{}, // 0xx
		{}, // 1xx
		{
			/*200*/ jps.ResponseStatusOK,
			/*201*/ jps.ResponseStatusCreated,
			/*202*/ jps.ResponseStatusAccepted,
			/*203*/ jps.ResponseStatusNonAuthoritative,
			/*204*/ jps.ResponseStatusNoContent,
			/*205*/ jps.ResponseStatusResetContent,
		},
		{
			/*300*/ jps.ResponseStatusMultipleChoices,
			/*301*/ jps.ResponseStatusMovedPermanently,
			/*302*/ jps.ResponseStatusFound,
			/*303*/ jps.ResponseStatusSeeOther,
			/*305*/ jps.ResponseStatusUseProxy,
			/*306*/ 0,
			/*307*/ jps.ResponseStatusTemporaryRedirect,
		},
		{
			/*400*/ jps.ResponseStatusBadMessage,
			/*401*/ 0,
			/*402*/ jps.ResponseStatusPaymentRequired,
			/*403*/ jps.ResponseStatusForbidden,
			/*404*/ jps.ResponseStatusNotFound,
			/*405*/ jps.ResponseStatusMethodNotAllowed,
			/*406*/ jps.ResponseStatusNotAcceptable,
			/*407*/ 0,
			/*408*/ jps.ResponseStatusRequestTimeout,
			/*409*/ jps.ResponseStatusConflict,
			/*410*/ jps.ResponseStatusGone,
			/*411*/ jps.ResponseStatusLengthRequired,
			/*412*/ 0,
			/*413*/ jps.ResponseStatusPayloadTooLarge,
			/*414*/ jps.ResponseStatusURITooLong,
			/*415*/ jps.ResponseStatusUnsupportedMediaType,
			/*416*/ 0,
			/*417*/ jps.ResponseStatusExpectationFailed,
			/*417*/ jps.ResponseStatusExpectationFailed,
			/*418-420*/ 0, 0, 0,
			/*420-425*/ 0, 0, 0, 0, 0,
			/*426*/ jps.ResponseStatusUpgradeRequired,
			/*427-430*/ 0, 0, 0, 0,
			/*431-440*/ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			/*441-450*/ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			/*451*/ jps.ResponseStatusUnavailableForLegalReasons,
		},
		{
			/*500*/ jps.ResponseStatusInternalServerError,
			/*501*/ jps.ResponseStatusNotImplemented,
			/*502*/ jps.ResponseStatusBadGateway,
			/*503*/ jps.ResponseStatusServiceUnavailable,
			/*504*/ jps.ResponseStatusGatewayTimeout,
			/*505*/ jps.ResponseStatusVersionNotSupported,
		},
	},
}

//StatusToStatusCodeMap is a map of jps.ResponseStatus values to jps.CodeWithStrings values containing the corresponding HTTP status code and text.
//It is initialized by init.
var StatusToStatusCodeMap jps.ResponseStatusToCodeWithStringMap

func init() {
	StatusToStatusCodeMap = jps.NewResponseStatusToCodeWithStringMapFromValuePairs(
		jps.CodeWithString{0, "Unknown"},
		[]jps.ResponseStatusCodeWithStringPair{
			/*Default OK*/ {0, jps.CodeWithString{200, "OK"}},
			{jps.ResponseStatusOK, jps.CodeWithString{200, "OK"}},
			{jps.ResponseStatusCreated, jps.CodeWithString{201, "Created"}},
			{jps.ResponseStatusAccepted, jps.CodeWithString{202, "Accepted"}},
			{jps.ResponseStatusNonAuthoritative, jps.CodeWithString{203, "None Authoritative"}},
			{jps.ResponseStatusNoContent, jps.CodeWithString{204, "No Content"}},
			{jps.ResponseStatusResetContent, jps.CodeWithString{205, "Reset Content"}},
			{jps.ResponseStatusNotFound, jps.CodeWithString{404, "Not Found"}},
		},
	)
}
