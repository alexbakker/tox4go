package math

import "github.com/alexbakker/tox4go/crypto"

func DistanceBetween(publicKey1 *[crypto.PublicKeySize]byte, publicKey2 *[crypto.PublicKeySize]byte) *[crypto.PublicKeySize]byte {
	dist := new([crypto.PublicKeySize]byte)

	for i := 0; i < crypto.PublicKeySize; i++ {
		dist[i] = publicKey1[i] ^ publicKey2[i]
	}

	return dist
}

func Closest(target *[crypto.PublicKeySize]byte, publicKey1 *[crypto.PublicKeySize]byte, publicKey2 *[crypto.PublicKeySize]byte) *[crypto.PublicKeySize]byte {
	dist1 := DistanceBetween(target, publicKey1)
	dist2 := DistanceBetween(target, publicKey2)

	for i := 0; i < crypto.PublicKeySize; i++ {
		if dist1[i] < dist2[i] {
			return publicKey1
		}

		if dist1[i] > dist2[i] {
			return publicKey2
		}
	}

	//both are the same distance away from our target public key
	//at this point it doesn't really matter which one we return
	//so just return the first one..
	return publicKey1
}
