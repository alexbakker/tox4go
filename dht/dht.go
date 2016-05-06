package dht

import (
	"fmt"

	"github.com/Impyy/tox4go/crypto"
	"github.com/Impyy/tox4go/transport"
)

// DHT represents a DHT identity.
type DHT struct {
	PublicKey *[crypto.PublicKeySize]byte
	SecretKey *[crypto.SecretKeySize]byte
}

// NewDHT creates a new DHT identity and generates a new keypair for it.
func NewDHT() (*DHT, error) {
	publicKey, secretKey, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	inst := &DHT{
		PublicKey: publicKey,
		SecretKey: secretKey,
	}

	return inst, nil
}

// EncryptPacket encrypts the given packet.
func (i *DHT) EncryptPacket(packet transport.Packet, publicKey *[crypto.PublicKeySize]byte) (*Packet, error) {
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
func (i *DHT) DecryptPacket(p *Packet) (transport.Packet, error) {
	var tPacket transport.Packet

	switch p.Type {
	case packetIDGetNodes:
		tPacket = &GetNodesPacket{}
	case packetIDSendNodes:
		tPacket = &SendNodesPacket{}
	case packetIDPingRequest:
		tPacket = &PingRequestPacket{}
	case packetIDPingResponse:
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
