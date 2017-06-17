package toxstatus

import (
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"

	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/dht"
)

func Fetch() ([]*dht.Node, error) {
	type toxNode struct {
		Ipv4Address string `json:"ipv4"`
		Ipv6Address string `json:"ipv6"`
		Port        int    `json:"port"`
		TCPPorts    []int  `json:"tcp_ports"`
		PublicKey   string `json:"public_key"`
	}

	type toxStatus struct {
		Nodes []*toxNode `json:"nodes"`
	}

	res, err := http.Get("https://nodes.tox.chat/json")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	statusObj := toxStatus{}
	err = json.NewDecoder(res.Body).Decode(&statusObj)
	if err != nil {
		return nil, err
	}

	nodes := []*dht.Node{}
	for _, node := range statusObj.Nodes {
		publicKey := new([crypto.PublicKeySize]byte)
		tempKey, err := hex.DecodeString(node.PublicKey)
		if err != nil {
			continue
		}
		copy(publicKey[:], tempKey)

		nodes = append(nodes, &dht.Node{
			IP:        net.ParseIP(node.Ipv4Address),
			Port:      node.Port,
			PublicKey: publicKey,
			Type:      dht.NodeTypeUDP,
		})
	}
	return nodes, nil
}
