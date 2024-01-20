package bootstrap

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"errors"
	"fmt"
)

type PacketType byte

const (
	PacketTypeBootstrapInfo PacketType = 0xF0
)

const (
	requestPacketLength = 78
	maxMOTDLength       = 256
)

var ErrUnknownPacketType = errors.New("unknown packet type")

type Packet interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
	ID() PacketType
}

// RawPacket represents the base of all bootstrap node packets.
type RawPacket struct {
	Type    PacketType
	Payload []byte
}

// InfoResponsePacket represents the structure of a packet that is sent in
// response to a bootstrap node info request.
type InfoResponsePacket struct {
	Version uint32
	MOTD    string
}

// InfoRequestPacket represents the structure of the packet used to request
// info from a bootstrap node. It contains infoRequestPacketLength - 1 useless
// bytes.
type InfoRequestPacket struct{}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *RawPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, p.Type)
	if err != nil {
		return nil, err
	}

	_, err = buff.Write(p.Payload)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *RawPacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.BigEndian, &p.Type)
	if err != nil {
		return err
	}

	p.Payload = make([]byte, reader.Len())
	_, err = reader.Read(p.Payload)
	return err
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *InfoResponsePacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	err := binary.Write(buff, binary.BigEndian, p.Version)
	if err != nil {
		return nil, err
	}

	motdBytes := []byte(p.MOTD)
	if len(motdBytes) > maxMOTDLength {
		return nil, errors.New("MOTD too long")
	}

	_, err = buff.Write(motdBytes)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *InfoResponsePacket) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)

	err := binary.Read(reader, binary.BigEndian, &p.Version)
	if err != nil {
		return err
	}

	motdBytes := make([]byte, reader.Len())
	if len(motdBytes) > maxMOTDLength {
		return errors.New("MOTD too long")
	}

	_, err = reader.Read(motdBytes)
	if err != nil {
		return err
	}

	p.MOTD = string(bytes.Trim(motdBytes, "\x00"))
	return nil
}

// ID returns the packet ID of this packet.
func (p InfoResponsePacket) ID() PacketType {
	return PacketTypeBootstrapInfo
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (p *InfoRequestPacket) MarshalBinary() ([]byte, error) {
	buff := new(bytes.Buffer)

	zeroes := make([]byte, requestPacketLength-1)
	_, err := buff.Write(zeroes)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), err
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *InfoRequestPacket) UnmarshalBinary(data []byte) error {
	dataLen := len(data)
	expected := requestPacketLength - 1

	if dataLen != expected {
		return fmt.Errorf("invalid packet length: %d, expected: %d", dataLen, expected)
	}

	return nil
}

// ID returns the packet ID of this packet.
func (p InfoRequestPacket) ID() PacketType {
	return PacketTypeBootstrapInfo
}

func UnmarshalBinary(data []byte) (Packet, error) {
	var raw RawPacket
	if err := raw.UnmarshalBinary(data); err != nil {
		return nil, err
	}

	return UnmarshalPacket(&raw)
}

func UnmarshalPacket(p *RawPacket) (Packet, error) {
	var res Packet

	switch p.Type {
	case PacketTypeBootstrapInfo:
		packetLen := len(p.Payload)

		if packetLen == requestPacketLength-1 && sliceIsZero(p.Payload) {
			res = &InfoRequestPacket{}
		} else {
			res = &InfoResponsePacket{}
		}
	default:
		return nil, fmt.Errorf("%w: %d", ErrUnknownPacketType, p.Type)
	}

	if err := res.UnmarshalBinary(p.Payload); err != nil {
		return nil, err
	}

	return res, nil
}

func MarshalPacket(packet Packet) (*RawPacket, error) {
	payload, err := packet.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return &RawPacket{
		Type:    packet.ID(),
		Payload: payload,
	}, nil
}

func sliceIsZero(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}
