package relay

import (
	"bytes"
	"errors"

	"github.com/Impyy/tox4go/crypto"
	"github.com/Impyy/tox4go/transport"
)

type Connection struct {
	PublicKey     *[crypto.PublicKeySize]byte
	SecretKey     *[crypto.SecretKeySize]byte
	PeerPublicKey *[crypto.PublicKeySize]byte
	verified      bool
	baseNonce     *[crypto.NonceSize]byte
}

func NewConnection(peerPublicKey *[crypto.PublicKeySize]byte) (*Connection, error) {
	publicKey, secretKey, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	conn := &Connection{
		PublicKey:     publicKey,
		SecretKey:     secretKey,
		PeerPublicKey: peerPublicKey,
		verified:      false,
		baseNonce:     nil,
	}

	return conn, nil
}

func (c *Connection) StartHandshake() (*HandshakePayload, error) {
	baseNonce, err := crypto.GenerateNonce()
	if err != nil {
		return nil, err
	}
	c.baseNonce = baseNonce

	return &HandshakePayload{
		PublicKey: c.PublicKey,
		BaseNonce: baseNonce,
	}, nil
}

func (c *Connection) EndHandshake(res *HandshakePayload) error {
	if c.baseNonce == nil || !bytes.Equal(res.BaseNonce[:], c.baseNonce[:]) {
		return errors.New("handshake failed: base nonces are not equal")
	}

	if !bytes.Equal(res.PublicKey[:], c.PublicKey[:]) {
		return errors.New("handshake failed: public keys are not equal")
	}

	c.verified = true
	return nil
}

func (c *Connection) Verified() bool {
	return c.verified
}

// EncryptPacket encrypts the given packet.
func (c *Connection) EncryptPacket(packet transport.Packet) (*Packet, error) {
	if !c.verified {
		return nil, errors.New("complete a handshake first")
	}

	return nil, errors.New("not implemented")
}

// DecryptPacket decrypts the given packet.
func (c *Connection) DecryptPacket(p *Packet) (transport.Packet, error) {
	if !c.verified {
		return nil, errors.New("complete a handshake first")
	}

	return nil, errors.New("not implemented")
}
