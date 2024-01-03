package transport

import "net"

// PacketHandler is a handler function for Tox packets. The backing buffer of
// the given data slice may be reused after the function returns, potentially
// overwriting the contents of the slice.
type PacketHandler func(data []byte, addr *net.UDPAddr)

type Packet interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	ID() byte
}

type Transport interface {
	SendPacket(data []byte, addr *net.UDPAddr) error
	HandlePacket(data []byte, addr *net.UDPAddr)
	Listen() error
	Close() error
}
