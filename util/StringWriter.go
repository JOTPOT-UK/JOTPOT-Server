package util

type StringWriter interface {
	Write([]byte) (int, error)
	WriteString(string) (int, error)
}
