package toxstatus

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/http"

	"github.com/alexbakker/tox4go/dht"
)

const (
	defaultURL = "https://nodes.tox.chat/json"
)

type Client struct {
	HTTPClient          *http.Client
	URL                 string
	IncludeOfflineNodes bool
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
			Online    bool   `json:"status_udp"`
		} `json:"nodes"`
	}
	if err = json.NewDecoder(httpRes.Body).Decode(&statusObj); err != nil {
		return nil, err
	}

	var res []*dht.Node
	for _, node := range statusObj.Nodes {
		if !node.Online && !c.IncludeOfflineNodes {
			continue
		}

		publicKey := new(dht.PublicKey)
		decPublicKey, err := hex.DecodeString(node.PublicKey)
		if err != nil {
			continue
		}
		copy(publicKey[:], decPublicKey)

		var ip net.IP
		var nodeType dht.NodeType
		if node.IP4Addr != "" && node.IP4Addr != "-" {
			var r net.Resolver
			if ips, err := r.LookupIP(ctx, "ip4", node.IP4Addr); err == nil && len(ips) > 0 {
				ip = ips[0]
				nodeType = dht.NodeTypeUDPIP4
			}
		} else if node.IP6Addr != "" && node.IP6Addr != "-" {
			var r net.Resolver
			if ips, err := r.LookupIP(ctx, "ip6", node.IP6Addr); err == nil && len(ips) > 0 {
				ip = ips[0]
				nodeType = dht.NodeTypeUDPIP6
			}
		}

		if ip == nil {
			continue
		}

		res = append(res, &dht.Node{
			IP:        ip,
			Port:      node.Port,
			PublicKey: publicKey,
			Type:      nodeType,
		})
	}

	return res, nil
}
