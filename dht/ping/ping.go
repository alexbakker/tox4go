package ping

import (
	"time"

	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/dht"
)

const (
	Timeout = time.Second * 20
)

type Ping struct {
	PublicKey *dht.PublicKey
	ID        uint64
	Time      time.Time
}

func NewPing(publicKey *dht.PublicKey) (*Ping, error) {
	pingID, err := crypto.GeneratePingID()
	if err != nil {
		return nil, err
	}

	return &Ping{
		PublicKey: publicKey,
		ID:        pingID,
		Time:      time.Now(),
	}, nil
}

func (p *Ping) Expired() bool {
	return time.Since(p.Time) > Timeout
}
