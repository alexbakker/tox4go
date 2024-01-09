package dht

import (
	"fmt"

	"github.com/alexbakker/tox4go/crypto"
	lru "github.com/hashicorp/golang-lru/v2"
)

// Identity represents a DHT identity.
type Identity struct {
	PublicKey *PublicKey
	SecretKey *[crypto.SecretKeySize]byte

	cache *lru.Cache[PublicKey, *[crypto.SharedKeySize]byte]
}

type IdentityOptions struct {
	// SharedKeyCacheSize is the size of the LRU cache for precomputed shared keys.
	SharedKeyCacheSize int
}

// NewIdentity creates a new DHT identity and generates a new keypair for it.
func NewIdentity(opts IdentityOptions) (*Identity, error) {
	publicKey, secretKey, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("new dht identity: %w", err)
	}

	var cache *lru.Cache[PublicKey, *[crypto.SharedKeySize]byte]
	if opts.SharedKeyCacheSize > 0 {
		cache, err = lru.New[PublicKey, *[crypto.SharedKeySize]byte](opts.SharedKeyCacheSize)
		if err != nil {
			return nil, fmt.Errorf("lru cache: %w", err)
		}
	}

	ident := &Identity{
		PublicKey: (*PublicKey)(publicKey),
		SecretKey: secretKey,
		cache:     cache,
	}

	return ident, nil
}

// EncryptPacket encrypts the given packet.
func (i *Identity) EncryptPacket(packet Packet, publicKey *PublicKey) (*EncryptedPacket, error) {
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
		return nil, fmt.Errorf("unknown dht packet type: %d", p.Type)
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
func (i *Identity) EncryptBlob(data []byte, publicKey *PublicKey) ([]byte, *[crypto.NonceSize]byte, error) {
	sharedKey := i.precomputeKey(publicKey)
	encryptedPayload, nonce, err := crypto.Encrypt(data, sharedKey)
	if err != nil {
		return nil, nil, err
	}

	return encryptedPayload, nonce, nil
}

// DecryptBlob decrypts the given slice of data.
func (i *Identity) DecryptBlob(data []byte, publicKey *PublicKey, nonce *[crypto.NonceSize]byte) ([]byte, error) {
	sharedKey := i.precomputeKey(publicKey)
	decryptedData, err := crypto.Decrypt(data, sharedKey, nonce)
	if err != nil {
		return nil, err
	}

	return decryptedData, nil
}

func (i *Identity) precomputeKey(publicKey *PublicKey) *[crypto.SharedKeySize]byte {
	if i.cache != nil {
		if sharedKey, ok := i.cache.Get(*publicKey); ok {
			return sharedKey
		}
	}

	sharedKey := crypto.PrecomputeKey((*[PublicKeySize]byte)(publicKey), i.SecretKey)
	if i.cache != nil {
		i.cache.Add(*publicKey, sharedKey)
	}

	return sharedKey
}
