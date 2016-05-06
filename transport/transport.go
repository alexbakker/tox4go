package transport

type Packet interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
	ID() byte
}
