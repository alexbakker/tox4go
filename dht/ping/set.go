package ping

import (
	"container/list"
	"fmt"
	"maps"
	"time"

	"github.com/alexbakker/tox4go/dht"
)

const (
	DefaultTimeout = time.Second * 5
)

type pingKey struct {
	ID        uint64
	PublicKey dht.PublicKey
}

type Set struct {
	pings    map[pingKey]*Ping
	pingList *list.List
	timeout  time.Duration

	lastRealloc time.Time
}

func NewSet(timeout time.Duration) *Set {
	return &Set{
		pings:    make(map[pingKey]*Ping),
		pingList: list.New(),
		timeout:  timeout,
	}
}

func (s *Set) Size() int {
	return len(s.pings)
}

func (s *Set) Add(publicKey *dht.PublicKey) (*Ping, error) {
	s.clearExpired()

	p, err := New(publicKey)
	if err != nil {
		return nil, err
	}

	key := pingKey{PublicKey: *p.publicKey, ID: p.id}
	if _, ok := s.pings[key]; ok {
		return nil, fmt.Errorf("ping id already in set: %d", p.id)
	}

	s.pings[key] = p
	p.e = s.pingList.PushBack(p)
	return p, nil
}

// Pop tries to find the given ping ID and public key in the set of pings. If it
// is found, it is removed from the set and returned. If it is not found, an
// error is returned. Expiry can also be the reason that the given ping ID is
// not found, but this will not be reported as a separate error.
func (s *Set) Pop(publicKey *dht.PublicKey, id uint64) (*Ping, error) {
	s.clearExpired()

	key := pingKey{PublicKey: *publicKey, ID: id}
	p, ok := s.pings[key]
	if !ok {
		return nil, fmt.Errorf("ping id not in set: %d", id)
	}

	s.delete(p)
	return p, nil
}

func (s *Set) clearExpired() {
	var next *list.Element
	for e := s.pingList.Front(); e != nil; e = next {
		p := e.Value.(*Ping)
		if !p.Expired(s.timeout) {
			break
		}

		next = e.Next()
		s.delete(p)
	}

	// Reallocate the ping map every once in a while, so that the memory usage
	// of the map doesn't keep growing: https://github.com/golang/go/issues/20135
	if now().Sub(s.lastRealloc) > 1*time.Second {
		s.pings = maps.Clone(s.pings)
		s.lastRealloc = now()
	}
}

func (s *Set) delete(p *Ping) {
	delete(s.pings, pingKey{PublicKey: *p.publicKey, ID: p.id})
	s.pingList.Remove(p.e)
}
