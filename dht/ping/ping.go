package ping

import (
	"container/list"
	"time"

	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/dht"
)

// now is the function used to obtain the current time. It is overridden by tests.
var now = time.Now

type Ping struct {
	publicKey *dht.PublicKey
	id        uint64
	time      time.Time

	e *list.Element
}

func New(publicKey *dht.PublicKey) (*Ping, error) {
	pingID, err := crypto.GeneratePingID()
	if err != nil {
		return nil, err
	}

	return &Ping{
		publicKey: publicKey,
		id:        pingID,
		time:      now(),
	}, nil
}

func (p *Ping) Expired(timeout time.Duration) bool {
	return now().Sub(p.time) > timeout
}

func (p *Ping) ID() uint64 {
	return p.id
}
