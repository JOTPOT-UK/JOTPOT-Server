package udp

type UDPProtocol struct{}

func (_ UDPProtocol) Protocol() string {
	return "udp"
}

func (_ UDPProtocol) String() string {
	return "udp"
}

func (_ UDPProtocol) Version() interface{} {
	return nil
}
