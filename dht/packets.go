package dht

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
	"net"

	"github.com/alexbakker/tox4go/crypto"
)

type PacketType byte

const (
	PacketTypePingRequest  PacketType = 0
	PacketTypePingResponse PacketType = 1
	PacketTypeGetNodes     PacketType = 2
	PacketTypeSendNodes    PacketType = 4
)

type NodeType byte

const (
	NodeTypeUDPIP4 NodeType = 2
	NodeTypeUDPIP6 NodeType = 10
	NodeTypeTCPIP4 NodeType = 130
	NodeTypeTCPIP6 NodeType = 138
)

type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	ID() PacketType
}

// EncryptedPacket represents an encrypted DHT packet.
type EncryptedPacket struct {
	Type            PacketType
	SenderPublicKey *[crypto.PublicKeySize]byte
	Nonce           *[crypto.NonceSize]byte
	Payload         []byte /* encrypted */
}

// Node represents a node in the DHT.
type Node struct {
	Type      NodeType
	PublicKey *[crypto.PublicKeySize]byte
	IP        net.IP
	Port      int
}

// GetNodesPacket represents the encrypted portion of the GetNodes request.
type GetNodesPacket struct {
	PublicKey *[crypto.PublicKeySize]byte
	PingID    uint64
}

// SendNodesPacket represents the encrypted portion of the SendNodes packet.
type SendNodesPacket struct {
	Nodes  []*Node
	PingID uint64
}

// PingResponsePacket represents the encrypted portion of the PingResponse packet.
type PingResponsePacket struct {
	//ping type: 0x01
	PingID uint64
}

// PingRequestPacket represents the encrypted portion of the PingRequest packet.
type PingRequestPacket struct {
	//ping type: 0x00
	PingID uint64
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *GetNodesPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	_, err := buff.Write(p.PublicKey[:])
	if err != nil {
		return nil, err
	}

	err = binary.Write(buff, binary.BigEndian, p.PingID)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *GetNodesPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	p.PublicKey = new([crypto.PublicKeySize]byte)
	_, err := reader.Read(p.PublicKey[:])
	if err != nil {
		return err
	}

	return binary.Read(reader, binary.BigEndian, &p.PingID)
}

// ID returns the packet ID of this packet.
func (p GetNodesPacket) ID() PacketType {
	return PacketTypeGetNodes
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *EncryptedPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, p.Type)
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(p.SenderPublicKey[:])
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(p.Nonce[:])
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(p.Payload)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *EncryptedPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.BigEndian, &p.Type)
	if err != nil {
		return err
	}

	p.SenderPublicKey = new([crypto.PublicKeySize]byte)
	_, err = reader.Read(p.SenderPublicKey[:])
	if err != nil {
		return err
	}

	p.Nonce = new([crypto.NonceSize]byte)
	_, err = reader.Read(p.Nonce[:])
	if err != nil {
		return err
	}

	p.Payload = make([]byte, reader.Len())
	_, err = reader.Read(p.Payload)
	return err
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *SendNodesPacket) MarshalBinary() ([]byte, error) {
	if len(p.Nodes) > 4 {
		return nil, errors.New("too many nodes, the max is 4")
	}

	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, byte(len(p.Nodes)))
	if err != nil {
		return nil, err
	}

	for _, node := range p.Nodes {
		nodeBytes, err2 := node.MarshalBinary()
		if err2 != nil {
			return nil, err2
		}

		_, err2 = buff.Write(nodeBytes)
		if err2 != nil {
			return nil, err2
		}
	}

	err = binary.Write(buff, binary.BigEndian, p.PingID)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *SendNodesPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var count byte
	err := binary.Read(reader, binary.BigEndian, &count)
	if err != nil {
		return err
	}

	if count > 4 {
		return errors.New("too many nodes, the max is 4")
	}

	p.Nodes = make([]*Node, int(count))
	for i := range p.Nodes {
		var nodeType NodeType
		var ipSize int

		p.Nodes[i] = &Node{}

		err = binary.Read(reader, binary.BigEndian, &nodeType)
		if err != nil {
			return err
		}

		switch nodeType {
		case NodeTypeUDPIP4, NodeTypeTCPIP4:
			ipSize = net.IPv4len
		case NodeTypeUDPIP6, NodeTypeTCPIP6:
			ipSize = net.IPv6len
		default:
			return fmt.Errorf("bad address family: %d", nodeType)
		}

		nodeBytes := make([]byte, 1+ipSize+2+crypto.PublicKeySize)
		nodeBytes[0] = byte(nodeType)
		_, err := reader.Read(nodeBytes[1:])
		if err != nil {
			return err
		}

		err = p.Nodes[i].UnmarshalBinary(nodeBytes)
		if err != nil {
			return err
		}
	}

	return binary.Read(reader, binary.BigEndian, &p.PingID)
}

// ID returns the packet ID of this packet.
func (p SendNodesPacket) ID() PacketType {
	return PacketTypeSendNodes
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *PingResponsePacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, PacketTypePingResponse)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buff, binary.BigEndian, p.PingID)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *PingResponsePacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var pingType PacketType
	err := binary.Read(reader, binary.BigEndian, &pingType)
	if err != nil {
		return err
	} else if pingType != PacketTypePingResponse {
		return fmt.Errorf("incorrect ping type: %d! is this a replay attack?", pingType)
	}

	return binary.Read(reader, binary.BigEndian, &p.PingID)
}

// ID returns the packet ID of this packet.
func (p PingResponsePacket) ID() PacketType {
	return PacketTypePingResponse
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *PingRequestPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, PacketTypePingRequest)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buff, binary.BigEndian, p.PingID)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *PingRequestPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	var pingType PacketType
	err := binary.Read(reader, binary.BigEndian, &pingType)
	if err != nil {
		return err
	} else if pingType != PacketTypePingRequest {
		return fmt.Errorf("incorrect ping type: %d! is this a replay attack?", pingType)
	}

	return binary.Read(reader, binary.BigEndian, &p.PingID)
}

// ID returns the packet ID of this packet.
func (p PingRequestPacket) ID() PacketType {
	return PacketTypePingRequest
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (n *Node) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)
	var nodeType NodeType
	var ipSize int

	err := binary.Read(reader, binary.BigEndian, &nodeType)
	if err != nil {
		return err
	}

	switch nodeType {
	case NodeTypeUDPIP4, NodeTypeTCPIP4:
		ipSize = net.IPv4len
	case NodeTypeUDPIP6, NodeTypeTCPIP6:
		ipSize = net.IPv6len
	default:
		return fmt.Errorf("unknown address family: %d", nodeType)
	}
	n.Type = nodeType

	n.IP = make([]byte, ipSize)
	_, err = reader.Read(n.IP)
	if err != nil {
		return err
	}

	var port uint16
	err = binary.Read(reader, binary.BigEndian, &port)
	if err != nil {
		return err
	}
	n.Port = int(port)

	n.PublicKey = new([crypto.PublicKeySize]byte)
	_, err = reader.Read(n.PublicKey[:])
	return err
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (n *Node) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)

	err := binary.Write(buf, binary.BigEndian, n.Type)
	if err != nil {
		return nil, err
	}

	if _, err = buf.Write(n.IP); err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.BigEndian, uint16(n.Port))
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(n.PublicKey[:])
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (n *Node) Addr() net.Addr {
	switch n.Type {
	case NodeTypeUDPIP4, NodeTypeTCPIP4:
		return &net.UDPAddr{IP: n.IP, Port: n.Port}
	case NodeTypeUDPIP6, NodeTypeTCPIP6:
		return &net.TCPAddr{IP: n.IP, Port: n.Port}
	default:
		panic(fmt.Sprintf("unsupported node type: %d", n.Type))
	}
}

func (t NodeType) Net() string {
	switch t {
	case NodeTypeUDPIP4:
		return "udp4"
	case NodeTypeUDPIP6:
		return "udp6"
	case NodeTypeTCPIP4:
		return "tcp4"
	case NodeTypeTCPIP6:
		return "tcp6"
	default:
		panic(fmt.Sprintf("bad node type: %d", t))
	}
}

func (t PacketType) String() string {
	var name string
	switch t {
	case PacketTypePingRequest:
		name = "PING_REQUEST"
	case PacketTypePingResponse:
		name = "PING_RESPONSE"
	case PacketTypeGetNodes:
		name = "GET_NODES"
	case PacketTypeSendNodes:
		name = "SEND_NODES"
	default:
		name = "UNKNOWN"
	}

	return fmt.Sprintf("%s(%d)", name, t)
}
