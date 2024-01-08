package ping

import (
	"fmt"
	"time"

	"github.com/alexbakker/tox4go/dht"
)

type pingKey struct {
	ID        uint64
	PublicKey dht.PublicKey
}

type Collection struct {
	pings   []*Ping
	pingMap map[pingKey]*Ping
	timeout time.Duration
}

func NewCollection(timeout time.Duration) *Collection {
	return &Collection{
		pingMap: make(map[pingKey]*Ping),
		timeout: timeout,
	}
}

func (c *Collection) Add(p *Ping) error {
	c.ClearExpired()

	key := pingKey{PublicKey: *p.PublicKey, ID: p.ID}
	if _, ok := c.pingMap[key]; ok {
		return fmt.Errorf("ping id already in collection: %d", p.ID)
	}

	c.pingMap[key] = p
	c.pings = append(c.pings, p)
	return nil
}

func (c *Collection) AddNew(publicKey *dht.PublicKey) (*Ping, error) {
	p, err := New(publicKey)
	if err != nil {
		return nil, err
	}

	if err = c.Add(p); err != nil {
		return nil, err
	}

	return p, nil
}

// Pop tries to find the given ping ID and public key in the collection of
// pings. If it is found, it is removed from the collection and returned. If it
// is not found, an error is returned. Expiry can also be the reason that the
// given ping ID is not found, but this will not be reported as a separate
// error.
func (c *Collection) Pop(publicKey *dht.PublicKey, id uint64) (*Ping, error) {
	c.ClearExpired()

	key := pingKey{PublicKey: *publicKey, ID: id}
	p, ok := c.pingMap[key]
	if !ok {
		return nil, fmt.Errorf("ping id not in collection: %d", id)
	}

	// Delete the ping from the map immediately, but wait for expiry before
	// deleting it from the time-sorted list, because lookups in the latter is
	// expensive for non-expired pings in large ping lists
	delete(c.pingMap, key)
	return p, nil
}

func (c *Collection) ClearExpired() {
	// Iterate through the time-sorted list of pings and delete any pings from
	// the map that have expired
	var deli int
	for i, p := range c.pings {
		if !p.Expired(c.timeout) {
			deli = i + 1
			break
		}

		delete(c.pingMap, pingKey{PublicKey: *p.PublicKey, ID: p.ID})
	}

	// Slice the pings that have expired away from the ping list
	c.pings = c.pings[deli:]
}
