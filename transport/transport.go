package transport

import "net"

// Handler is a handler function for Tox packets. The backing buffer of the
// given data slice may be reused after the function returns, overwriting its
// contents.
type Handler func(data []byte, addr *net.UDPAddr)

type Packet interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	ID() byte
}

type Transport interface {
	Send(data []byte, addr *net.UDPAddr) error
	Handle(handler Handler)
	HandlePacket(packetID byte, handler Handler)
	Listen() error
	Close() error
}
