package dht

import (
	"fmt"

	"github.com/Impyy/tox4go/crypto"
	"github.com/Impyy/tox4go/transport"
)

// Ident represents a DHT identity.
type Ident struct {
	PublicKey *[crypto.PublicKeySize]byte
	SecretKey *[crypto.SecretKeySize]byte
}

// NewIdent creates a new DHT identity and generates a new keypair for it.
func NewIdent() (*Ident, error) {
	publicKey, secretKey, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	inst := &Ident{
		PublicKey: publicKey,
		SecretKey: secretKey,
	}

	return inst, nil
}

// EncryptPacket encrypts the given packet.
func (i *Ident) EncryptPacket(packet transport.Packet, publicKey *[crypto.PublicKeySize]byte) (*Packet, error) {
	base := Packet{}
	base.Type = packet.ID()
	base.SenderPublicKey = i.PublicKey

	payload, err := packet.MarshalBinary()
	if err != nil {
		return nil, err
	}

	sharedKey := crypto.PrecomputeKey(publicKey, i.SecretKey)
	encryptedPayload, nonce, err := crypto.Encrypt(payload, sharedKey)
	if err != nil {
		return nil, err
	}

	base.Nonce = nonce
	base.Payload = encryptedPayload

	return &base, nil
}

// DecryptPacket decrypts the given packet.
func (i *Ident) DecryptPacket(p *Packet) (transport.Packet, error) {
	var tPacket transport.Packet

	switch p.Type {
	case PacketIDGetNodes:
		tPacket = &GetNodesPacket{}
	case PacketIDSendNodes:
		tPacket = &SendNodesPacket{}
	case PacketIDPingRequest:
		tPacket = &PingRequestPacket{}
	case PacketIDPingResponse:
		tPacket = &PingResponsePacket{}
	default:
		return nil, fmt.Errorf("unknown packet type: %d", p.Type)
	}

	sharedKey := crypto.PrecomputeKey(p.SenderPublicKey, i.SecretKey)
	decryptedData, err := crypto.Decrypt(p.Payload, sharedKey, p.Nonce)
	if err != nil {
		return nil, err
	}

	err = tPacket.UnmarshalBinary(decryptedData)
	if err != nil {
		return nil, err
	}

	return tPacket, nil
}
