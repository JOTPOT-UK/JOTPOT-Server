package jps

type ResponseType uint8
type ResponseCode uint8
type ResponseStatus uint16

const ResponseTypeOK = ResponseType(0)
const ResponseTypeClientError = ResponseType(0x80)
const ResponseTypeServerError = ResponseType(0x40)
const ResponseTypeRedirect = ResponseType(0x20)
const ResponseCodeClientErrorMask = ResponseStatus(ResponseTypeClientError) << 8
const ResponseCodeServerErrorMask = ResponseStatus(ResponseTypeServerError) << 8
const ResponseCodeNotOKMask = ResponseCodeClientErrorMask | ResponseCodeServerErrorMask
const ResponseCodeRedirectMask = ResponseStatus(ResponseTypeRedirect) << 8

func (s ResponseStatus) OK() bool {
	return s&ResponseCodeNotOKMask == 0
}
func (s ResponseStatus) IsClientError() bool {
	return s&ResponseCodeClientErrorMask != 0
}
func (s ResponseStatus) IsServerError() bool {
	return s&ResponseCodeServerErrorMask != 0
}
func (s ResponseStatus) IsRedirect() bool {
	return s&ResponseCodeRedirectMask != 0
}

func (s ResponseStatus) Type() ResponseType {
	return ResponseType(s >> 8)
}
func (s ResponseStatus) GetCode() ResponseCode {
	return ResponseCode(s & 0xff)
}

//Failure? 4 - Other error
//NO_CONNECTION?
//CONNECTION_LOST?
//SSH_FX_INVALID_HANDLE?
//NO_SUCH_PATH? 10
//FILE_ALREADY_EXISTS? 11
//FILE_ALREADY_EXISTS? 12
//NO_MEDIA? 13
/*
   SSH_FX_NO_SPACE_ON_FILESYSTEM        14
   SSH_FX_QUOTA_EXCEEDED                15
   SSH_FX_UNKNOWN_PRINCIPAL             16
   SSH_FX_LOCK_CONFLICT                 17
   SSH_FX_DIR_NOT_EMPTY                 18
   SSH_FX_NOT_A_DIRECTORY               19
   SSH_FX_INVALID_FILENAME              20
   SSH_FX_LINK_LOOP                     21
   SSH_FX_CANNOT_DELETE                 22
   SSH_FX_INVALID_PARAMETER             23
   SSH_FX_FILE_IS_A_DIRECTORY           24
   SSH_FX_BYTE_RANGE_LOCK_CONFLICT      25
   SSH_FX_BYTE_RANGE_LOCK_REFUSED       26
   SSH_FX_DELETE_PENDING                27
   SSH_FX_FILE_CORRUPT                  28
   SSH_FX_OWNER_INVALID                 29
   SSH_FX_GROUP_INVALID                 30
       SSH_FX_NO_MATCHING_BYTE_RANGE_LOCK       31
*/
//TODO: SFTP SSH_FX_EOF (1)
//SSH_FX_FAILURE (4) -- other error

var (
	//ResponseStatusOK is the response status given on success with no extra details.
	//HTTP: 200, SFTP: SSH_FX_OK (0)
	ResponseStatusOK = ResponseStatus(1)
	//ResponseStatusCreated is the response status given when a resource is created.
	//HTTP: 201
	ResponseStatusCreated = ResponseStatus(2)
	//ResponseStatusAccepted ...
	//HTTP: 202
	ResponseStatusAccepted = ResponseStatus(3)
	//ResponseStatusNonAuthorative ...
	//HTTP: 203
	ResponseStatusNonAuthoritative = ResponseStatus(4)
	//ResponseStatusNoContent ...
	//HTTP: 204
	ResponseStatusNoContent = ResponseStatus(5)
	//ResponseStatusResetContent ...
	//HTTP: 205
	ResponseStatusResetContent = ResponseStatus(6)
	// -- Redirects -- //TODO: Change order
	ResponseStatusMultipleChoices   = ResponseStatus(ResponseCodeRedirectMask | 8)
	ResponseStatusMovedPermanently  = ResponseStatus(ResponseCodeRedirectMask | 1)
	ResponseStatusFound             = ResponseStatus(ResponseCodeRedirectMask | 2)
	ResponseStatusSeeOther          = ResponseStatus(ResponseCodeRedirectMask | 3)
	ResponseStatusUseProxy          = ResponseStatus(ResponseCodeRedirectMask | 5)
	ResponseStatusTemporaryRedirect = ResponseStatus(ResponseCodeRedirectMask | 7)
	ResponseStatusNotModified       = ResponseStatus(ResponseCodeRedirectMask | 14)
	//ResponseStatusBadRequest
	//HTTP: 400 Bad Request, SFTP: SSH_FX_BAD_MESSAGE (5)
	ResponseStatusBadMessage = ResponseStatus(ResponseCodeClientErrorMask | 1)
	//ResponseStatusBadRequest is an alias of ResponseStatusBadMessage to make it easier for HTTP developers.
	ResponseStatusBadRequest      = ResponseStatusBadMessage
	ResponseStatusPaymentRequired = ResponseStatus(ResponseCodeClientErrorMask | 2)
	//ResponseStatusForbidden
	//HTTP: 403 Forbidden, SFTP SSH_FX_PERMISSION_DENIED (3)
	ResponseStatusForbidden = ResponseStatus(ResponseCodeClientErrorMask | 3)
	//ResponseStatusNotFound
	//HTTP: 404 Not Found, SFTP: SSH_FX_NO_SUCH_FILE (2) / SSH_FX_NO_SUCH_PATH(10) //TODO: SSH?
	ResponseStatusNotFound             = ResponseStatus(ResponseCodeClientErrorMask | 4)
	ResponseStatusMethodNotAllowed     = ResponseStatus(ResponseCodeClientErrorMask | 5)
	ResponseStatusNotAcceptable        = ResponseStatus(ResponseCodeClientErrorMask | 6)
	ResponseStatusRequestTimeout       = ResponseStatus(ResponseCodeClientErrorMask | 8)
	ResponseStatusConflict             = ResponseStatus(ResponseCodeClientErrorMask | 9)
	ResponseStatusGone                 = ResponseStatus(ResponseCodeClientErrorMask | 10)
	ResponseStatusLengthRequired       = ResponseStatus(ResponseCodeClientErrorMask | 11)
	ResponseStatusPayloadTooLarge      = ResponseStatus(ResponseCodeClientErrorMask | 13)
	ResponseStatusURITooLong           = ResponseStatus(ResponseCodeClientErrorMask | 14)
	ResponseStatusUnsupportedMediaType = ResponseStatus(ResponseCodeClientErrorMask | 15)
	ResponseStatusPreconditionFailed   = ResponseStatus(ResponseCodeClientErrorMask | 16)
	//ResponseStatusInvalidHandle
	//SFTP: SSH_FX_INVALID_HANDLE (9)
	ResponseStatusInvalidHandle = ResponseStatus(ResponseCodeClientErrorMask | 19)
	//ResponseStatusNoConnection should really be an error... But I guess it has to be a possible response code!
	//SFTP: SSH_FX_NO_CONNECTION (6)
	//ResponseStatusNoConnection        = ResponseStatus(ResponseCodeClientErrorMask | 16)
	//ResponseStatusConnectionLost      = ResponseStatus(ResponseCodeClientErrorMask | 16)
	ResponseStatusExpectationFailed = ResponseStatus(ResponseCodeClientErrorMask | 17)
	ResponseStatusNotADirectory     = ResponseStatus(ResponseCodeClientErrorMask | 19) // NOT HTTP!
	ResponseStatusUpgradeRequired   = ResponseStatus(ResponseCodeClientErrorMask | 26)
	//ResponseStatusFileAlreadyExists
	//SFTP: SSH_FX_FILE_ALREADY_EXISTS (11)
	ResponseStatusFileAlreadyExists = ResponseStatus(ResponseCodeClientErrorMask | 21)
	//ResponseStatusUnavailableForLegalReasons
	//HTTP: 451 Unavailable For Legal Reasons
	ResponseStatusUnavailableForLegalReasons = ResponseStatus(ResponseCodeClientErrorMask | 51)
	ResponseStatusInternalServerError        = ResponseStatus(ResponseCodeServerErrorMask | 1)
	//ResponseStautsNotImplemented
	//HTTP: 501, SFTP: SSH_FX_OP_UNSUPPORTED (8)
	ResponseStatusNotImplemented      = ResponseStatus(ResponseCodeServerErrorMask | 2)
	ResponseStatusBadGateway          = ResponseStatus(ResponseCodeServerErrorMask | 3)
	ResponseStatusServiceUnavailable  = ResponseStatus(ResponseCodeServerErrorMask | 4)
	ResponseStatusGatewayTimeout      = ResponseStatus(ResponseCodeServerErrorMask | 5)
	ResponseStatusVersionNotSupported = ResponseStatus(ResponseCodeServerErrorMask | 6)
)
