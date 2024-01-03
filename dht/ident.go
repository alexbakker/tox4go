package dht

import (
	"fmt"

	"github.com/alexbakker/tox4go/crypto"
)

// Identity represents a DHT identity.
type Identity struct {
	PublicKey *[crypto.PublicKeySize]byte
	SecretKey *[crypto.SecretKeySize]byte
}

// NewIdentity creates a new DHT identity and generates a new keypair for it.
func NewIdentity() (*Identity, error) {
	publicKey, secretKey, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("new dht identity: %w", err)
	}

	inst := &Identity{
		PublicKey: publicKey,
		SecretKey: secretKey,
	}

	return inst, nil
}

// EncryptPacket encrypts the given packet.
func (i *Identity) EncryptPacket(packet Packet, publicKey *[crypto.PublicKeySize]byte) (*EncryptedPacket, error) {
	base := EncryptedPacket{}
	base.Type = packet.ID()
	base.SenderPublicKey = i.PublicKey

	payload, err := packet.MarshalBinary()
	if err != nil {
		return nil, err
	}

	encryptedPayload, nonce, err := i.EncryptBlob(payload, publicKey)
	if err != nil {
		return nil, err
	}

	base.Nonce = nonce
	base.Payload = encryptedPayload

	return &base, nil
}

// DecryptPacket decrypts the given packet.
func (i *Identity) DecryptPacket(p *EncryptedPacket) (Packet, error) {
	var tPacket Packet

	switch p.Type {
	case PacketTypeGetNodes:
		tPacket = &GetNodesPacket{}
	case PacketTypeSendNodes:
		tPacket = &SendNodesPacket{}
	case PacketTypePingRequest:
		tPacket = &PingRequestPacket{}
	case PacketTypePingResponse:
		tPacket = &PingResponsePacket{}
	default:
		return nil, fmt.Errorf("unknown packet type: %d", p.Type)
	}

	decryptedData, err := i.DecryptBlob(p.Payload, p.SenderPublicKey, p.Nonce)
	if err != nil {
		return nil, err
	}

	err = tPacket.UnmarshalBinary(decryptedData)
	if err != nil {
		return nil, err
	}

	return tPacket, nil
}

// EncryptBlob encrypts the given slice of data.
func (i *Identity) EncryptBlob(data []byte, publicKey *[crypto.PublicKeySize]byte) ([]byte, *[crypto.NonceSize]byte, error) {
	sharedKey := crypto.PrecomputeKey(publicKey, i.SecretKey)
	encryptedPayload, nonce, err := crypto.Encrypt(data, sharedKey)
	if err != nil {
		return nil, nil, err
	}

	return encryptedPayload, nonce, nil
}

// DecryptBlob decrypts the given slice of data.
func (i *Identity) DecryptBlob(data []byte, publicKey *[crypto.PublicKeySize]byte, nonce *[crypto.NonceSize]byte) ([]byte, error) {
	sharedKey := crypto.PrecomputeKey(publicKey, i.SecretKey)
	decryptedData, err := crypto.Decrypt(data, sharedKey, nonce)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}
