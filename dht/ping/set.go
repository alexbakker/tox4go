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

type Set struct {
	pings   []*Ping
	pingMap map[pingKey]*Ping
	timeout time.Duration
}

func NewSet(timeout time.Duration) *Set {
	return &Set{
		pingMap: make(map[pingKey]*Ping),
		timeout: timeout,
	}
}

func (s *Set) Size() int {
	return len(s.pingMap)
}

func (s *Set) Add(p *Ping) error {
	s.ClearExpired()

	key := pingKey{PublicKey: *p.publicKey, ID: p.id}
	if _, ok := s.pingMap[key]; ok {
		return fmt.Errorf("ping id already in set: %d", p.id)
	}

	s.pingMap[key] = p
	s.pings = append(s.pings, p)
	return nil
}

func (s *Set) AddNew(publicKey *dht.PublicKey) (*Ping, error) {
	p, err := New(publicKey)
	if err != nil {
		return nil, err
	}

	if err = s.Add(p); err != nil {
		return nil, err
	}

	return p, nil
}

// Pop tries to find the given ping ID and public key in the set of pings. If it
// is found, it is removed from the set and returned. If it is not found, an
// error is returned. Expiry can also be the reason that the given ping ID is
// not found, but this will not be reported as a separate error.
func (s *Set) Pop(publicKey *dht.PublicKey, id uint64) (*Ping, error) {
	s.ClearExpired()

	key := pingKey{PublicKey: *publicKey, ID: id}
	p, ok := s.pingMap[key]
	if !ok {
		return nil, fmt.Errorf("ping id not in set: %d", id)
	}

	// Delete the ping from the map immediately, but wait for expiry before
	// deleting it from the time-sorted list, because lookups in the latter is
	// expensive for non-expired pings in large ping lists
	delete(s.pingMap, key)
	return p, nil
}

func (s *Set) ClearExpired() {
	// Iterate through the time-sorted list of pings and delete any pings from
	// the map that have expired
	var deli int
	for i, p := range s.pings {
		if !p.Expired(s.timeout) {
			deli = i
			break
		}

		delete(s.pingMap, pingKey{PublicKey: *p.publicKey, ID: p.id})
	}

	// Slice the pings that have expired away from the ping list
	s.pings = s.pings[deli:]
}
