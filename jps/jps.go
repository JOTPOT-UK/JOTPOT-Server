package jps

import (
	"io"
	"net"
)

type Protocol interface {
	Protocol() string
	Version() interface{}
	String() string
}

//Hop describes a proxy that a connection may have gone through. For example, a HTTP cache.
type Hop interface {
	Addr() net.Addr
	Label() string
	String() string
}

//ConnectionDetails is an interface which provides details about where the connection is from and its protocol.
type ConnectionDetails interface {
	//Protocol returns the higest level Protocol which is in use.
	Protocol() Protocol
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	//Via returns, if known, any hops that the connection has gone through.
	//These are intermediate proxies such as HTTP caches.
	//The boolean result will be true if the hops are known, or false otherwise.
	Via() ([]Hop, bool)
}

type Flusher interface {
	Flush() error
}

//WriteFlushCloser is an interface which has a Flush method on top of the standard io.WriteCloser methods.
type WriteFlushCloser interface {
	io.WriteCloser
	Flusher
}
