package http

import "github.com/JOTPOT-UK/JOTPOT-Server/jps"

//GetMethod returns the equivelent jps.Method for the given HTTP method string (note that it is case sensitive, so must be in upper-case).
func GetMethod(str string) jps.Method {
	switch str {
	case "GET":
		return jps.MethodRead
	case "POST":
		return jps.MethodPost
	case "PUT":
		return jps.MethodWrite
	case "HEAD":
		return jps.MethodStat
	case "OPTIONS":
		return jps.MethodOptions
	case "DELETE":
		return jps.MethodRemove
	default:
		return jps.MethodUnknown
	}
}

func GetMethodStr(method jps.Method) string {
	switch method {
	case jps.MethodRead:
		return "GET"
	case jps.MethodPost:
		return "POST"
	case jps.MethodWrite:
		return "PUT"
	case jps.MethodStat:
		return "HEAD"
	case jps.MethodOptions:
		return "OPTIONS"
	case jps.MethodRemove:
		return "DELETE"
	default:
		return ""
	}
}
