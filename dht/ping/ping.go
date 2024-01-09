package ping

import (
	"time"

	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/dht"
)

const (
	DefaultTimeout = time.Second * 20
)

type Ping struct {
	publicKey *dht.PublicKey
	id        uint64
	time      time.Time
}

func New(publicKey *dht.PublicKey) (*Ping, error) {
	pingID, err := crypto.GeneratePingID()
	if err != nil {
		return nil, err
	}

	return &Ping{
		publicKey: publicKey,
		id:        pingID,
		time:      time.Now(),
	}, nil
}

func (p *Ping) Expired(timeout time.Duration) bool {
	return time.Since(p.time) > timeout
}

func (p *Ping) ID() uint64 {
	return p.id
}
