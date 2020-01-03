package jps

type UnixProtocol struct{}

func (_ UnixProtocol) Protocol() string {
	return "unix"
}

func (_ UnixProtocol) String() string {
	return "unix"
}

func (_ UnixProtocol) Version() interface{} {
	return nil
}
