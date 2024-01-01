package toxstatus

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"

	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/dht"
)

const (
	defaultURL = "https://nodes.tox.chat/json"
)

type Client struct {
	HTTPClient *http.Client
	URL        string
}

func GetNodes(ctx context.Context) ([]*dht.Node, error) {
	return new(Client).GetNodes(ctx)
}

func (c *Client) GetNodes(ctx context.Context) ([]*dht.Node, error) {
	url := defaultURL
	if c.URL != "" {
		url = c.URL
	}

	hc := http.DefaultClient
	if c.HTTPClient != nil {
		hc = c.HTTPClient
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	httpRes, err := hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer httpRes.Body.Close()

	var statusObj struct {
		Nodes []*struct {
			IP4Addr   string `json:"ipv4"`
			IP6Addr   string `json:"ipv6"`
			Port      int    `json:"port"`
			TCPPorts  []int  `json:"tcp_ports"`
			PublicKey string `json:"public_key"`
		} `json:"nodes"`
	}
	if err = json.NewDecoder(httpRes.Body).Decode(&statusObj); err != nil {
		return nil, err
	}

	var res []*dht.Node
	for _, node := range statusObj.Nodes {
		publicKey := new([crypto.PublicKeySize]byte)
		decPublicKey, err := hex.DecodeString(node.PublicKey)
		if err != nil {
			continue
		}
		copy(publicKey[:], decPublicKey)

		res = append(res, &dht.Node{
			IP:        net.ParseIP(node.IP4Addr),
			Port:      node.Port,
			PublicKey: publicKey,
			Type:      dht.NodeTypeUDP,
		})
	}
	return res, nil
}
