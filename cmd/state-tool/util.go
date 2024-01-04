package main

import (
	"encoding/hex"

	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/dht"
	"github.com/alexbakker/tox4go/state"
)

// to anyone who's reading these convert functions: send help
func convertFriends(f1 []*state.Friend) []*friendJSON {
	friends := make([]*friendJSON, len(f1))

	for i, f := range f1 {
		friends[i] = &friendJSON{
			Status:         f.Status,
			UserStatus:     f.UserStatus,
			PublicKey:      hex.EncodeToString(f.PublicKey[:]),
			RequestMessage: f.RequestMessage,
			Name:           f.Name,
			StatusMessage:  f.StatusMessage,
			Nospam:         f.Nospam,
			LastSeen:       f.LastSeen,
		}
	}

	return friends
}

func convertFriendsBack(f1 []*friendJSON) ([]*state.Friend, error) {
	friends := make([]*state.Friend, len(f1))

	for i, f := range f1 {
		friends[i] = &state.Friend{
			Status:         f.Status,
			UserStatus:     f.UserStatus,
			RequestMessage: f.RequestMessage,
			Name:           f.Name,
			StatusMessage:  f.StatusMessage,
			Nospam:         f.Nospam,
			LastSeen:       f.LastSeen,
		}

		publicKey, err := hex.DecodeString(f.PublicKey)
		if err != nil {
			return nil, err
		} else if len(publicKey) != crypto.PublicKeySize {
			return nil, errorKeyLength{
				kind:     "public",
				expected: crypto.PublicKeySize,
				actual:   len(publicKey),
			}
		}
		friends[i].PublicKey = new([crypto.PublicKeySize]byte)
		copy(friends[i].PublicKey[:], publicKey)
	}

	return friends, nil
}

func convertNodes(n1 []*dht.Node) []*nodeJSON {
	nodes := make([]*nodeJSON, len(n1))

	for i, n := range n1 {
		nodes[i] = &nodeJSON{
			Type:      n.Type,
			PublicKey: hex.EncodeToString(n.PublicKey[:]),
			IP:        n.IP,
			Port:      n.Port,
		}
	}

	return nodes
}

func convertNodesBack(n1 []*nodeJSON) ([]*dht.Node, error) {
	nodes := make([]*dht.Node, len(n1))

	for i, n := range n1 {
		nodes[i] = &dht.Node{
			Type: n.Type,
			IP:   n.IP,
			Port: n.Port,
		}

		publicKey, err := hex.DecodeString(n.PublicKey)
		if err != nil {
			return nil, err
		} else if len(publicKey) != crypto.PublicKeySize {
			return nil, errorKeyLength{
				kind:     "public",
				expected: crypto.PublicKeySize,
				actual:   len(publicKey),
			}
		}
		nodes[i].PublicKey = new(dht.PublicKey)
		copy(nodes[i].PublicKey[:], publicKey)
	}

	return nodes, nil
}
