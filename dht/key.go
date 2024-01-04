package dht

import (
	"encoding/hex"

	"github.com/alexbakker/tox4go/crypto"
)

const PublicKeySize = crypto.PublicKeySize

type PublicKey [PublicKeySize]byte

func (pk *PublicKey) Closest(pk1 *PublicKey, pk2 *PublicKey) *PublicKey {
	dist1 := pk.DistanceTo(pk1)
	dist2 := pk.DistanceTo(pk2)

	for i := 0; i < PublicKeySize; i++ {
		if dist1[i] < dist2[i] {
			return pk1
		}

		if dist1[i] > dist2[i] {
			return pk2
		}
	}

	// Return pk1 in case both public keys are the same distance away from the
	// target public key
	return pk1
}

func (pk1 *PublicKey) DistanceTo(pk2 *PublicKey) *[PublicKeySize]byte {
	dist := new([PublicKeySize]byte)

	for i := 0; i < PublicKeySize; i++ {
		dist[i] = pk1[i] ^ pk2[i]
	}

	return dist
}

func (pk *PublicKey) String() string {
	return hex.EncodeToString(pk[:])
}
