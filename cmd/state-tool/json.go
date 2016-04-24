package main

import (
	"encoding/hex"
	"encoding/json"
	"net"

	"github.com/Impyy/tox4go/crypto"
	"github.com/Impyy/tox4go/dht"
	"github.com/Impyy/tox4go/state"
)

type (
	stateAlias  state.State
	friendAlias state.Friend
	nodeAlias   dht.Node
)

// stateJSON is a JSON-friendly version of the state.State struct
type stateJSON struct {
	PublicKey     string           `json:"public_key"`
	SecretKey     string           `json:"secret_key"`
	Nospam        uint32           `json:"nospam"`
	Name          string           `json:"name"`
	StatusMessage string           `json:"status_message"`
	Status        state.UserStatus `json:"status"`
	Friends       []*friendJSON    `json:"friends"`
	Nodes         []*nodeJSON      `json:"nodes"`
	TCPRelays     []*nodeJSON      `json:"tcp_relays"`
	PathNodes     []*nodeJSON      `json:"path_nodes"`
}

// friendJSON is a JSON-friendly version of the state.Friend struct
type friendJSON struct {
	Status         state.FriendStatus `json:"status"`
	UserStatus     state.UserStatus   `json:"user_status"`
	PublicKey      string             `json:"public_key"`
	RequestMessage string             `json:"request_message"`
	Name           string             `json:"name"`
	StatusMessage  string             `json:"status_message"`
	Nospam         uint32             `json:"nospam"`
	LastSeen       uint64             `json:"last_seen"`
}

// nodeJSON is a JSON-friendly version of the dht.Node struct
type nodeJSON struct {
	Type      dht.NodeType `json:"type"`
	PublicKey string       `json:"public_key"`
	IP        net.IP       `json:"ip"`
	Port      uint16       `json:"port"`
}

func (s *stateAlias) MarshalJSON() ([]byte, error) {
	return json.Marshal(stateJSON{
		PublicKey:     hex.EncodeToString(s.PublicKey[:]),
		SecretKey:     hex.EncodeToString(s.SecretKey[:]),
		Nospam:        s.Nospam,
		Name:          s.Name,
		StatusMessage: s.StatusMessage,
		Status:        s.Status,
		Friends:       convertFriends(s.Friends),
		Nodes:         convertNodes(s.Nodes),
		TCPRelays:     convertNodes(s.TCPRelays),
		PathNodes:     convertNodes(s.PathNodes),
	})
}

func (s *stateAlias) UnmarshalJSON(data []byte) error {
	temp := stateJSON{}
	err := json.Unmarshal(data, &temp)

	if err != nil {
		return err
	}

	s.Name = temp.Name
	s.Nospam = temp.Nospam
	s.Status = temp.Status
	s.StatusMessage = temp.StatusMessage

	friends, err := convertFriendsBack(temp.Friends)
	if err != nil {
		return err
	}
	s.Friends = friends

	pathNodes, err := convertNodesBack(temp.PathNodes)
	if err != nil {
		return err
	}
	s.PathNodes = pathNodes

	nodes, err := convertNodesBack(temp.Nodes)
	if err != nil {
		return err
	}
	s.Nodes = nodes

	tcpRelays, err := convertNodesBack(temp.TCPRelays)
	if err != nil {
		return err
	}
	s.TCPRelays = tcpRelays

	publicKey, err := hex.DecodeString(temp.PublicKey)
	if err != nil {
		return err
	} else if len(publicKey) != crypto.PublicKeySize {
		return errorKeyLength{
			kind:     "public",
			expected: crypto.PublicKeySize,
			actual:   len(publicKey),
		}
	}
	s.PublicKey = new([crypto.PublicKeySize]byte)
	copy(s.PublicKey[:], publicKey)

	secretKey, err := hex.DecodeString(temp.SecretKey)
	if err != nil {
		return err
	} else if len(secretKey) != crypto.SecretKeySize {
		return errorKeyLength{
			kind:     "secret",
			expected: crypto.SecretKeySize,
			actual:   len(secretKey),
		}
	}
	s.SecretKey = new([crypto.SecretKeySize]byte)
	copy(s.SecretKey[:], secretKey)

	return nil
}
