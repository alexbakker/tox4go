package ping

import (
	"crypto/rand"
	"testing"
	"time"

	"github.com/alexbakker/tox4go/dht"
)

func generatePublicKey(t *testing.T) *dht.PublicKey {
	var res dht.PublicKey
	if _, err := rand.Read(res[:]); err != nil {
		t.Fatal(err)
	}

	return &res
}

func addPing(t *testing.T, set *Set) (*dht.PublicKey, *Ping) {
	pk := generatePublicKey(t)
	p, err := set.Add(pk)
	if err != nil {
		t.Fatal(err)
	}
	return pk, p
}

func assertSetSize(t *testing.T, set *Set, size int) {
	if set.Size() != size {
		t.Fatalf("bad ping set size: expected: %d, actual: %d", size, set.Size())
	}

	if set.Size() != set.pingList.Len() {
		t.Fatalf("bad ping set size (list): expected: %d, actual: %d", size, set.pingList.Len())
	}
}

func TestPingExpiry(t *testing.T) {
	const timeout = 10 * time.Second

	st := time.Now()
	now = func() time.Time { return st }

	p, err := New(generatePublicKey(t))
	if err != nil {
		t.Fatal(err)
	}
	if p.Expired(timeout) {
		t.Fatal("unexpected ping expiry")
	}

	now = func() time.Time { return st.Add(time.Second * 8) }
	if p.Expired(timeout) {
		t.Fatal("unexpected ping expiry")
	}

	now = func() time.Time { return st.Add(time.Second * 11) }
	if !p.Expired(timeout) {
		t.Fatal("expected ping expiry")
	}
}

func TestPingSetExpirySingle(t *testing.T) {
	const timeout = 10 * time.Second

	set := NewSet(timeout)
	st := time.Now()
	now = func() time.Time { return st }

	assertSetSize(t, set, 0)
	pk1, p1 := addPing(t, set)
	assertSetSize(t, set, 1)
	if _, err := set.Pop(pk1, ^p1.id); err == nil {
		t.Fatal("popped bad ping id")
	}
	assertSetSize(t, set, 1)
	now = func() time.Time { return st.Add(time.Second * 8) }
	if pp, err := set.Pop(pk1, p1.id); err != nil || pp != p1 {
		t.Fatal("unable to pop ping")
	}
	assertSetSize(t, set, 0)
	if _, err := set.Pop(pk1, p1.id); err == nil {
		t.Fatal("popped a second time")
	}
	assertSetSize(t, set, 0)

	pk2, p2 := addPing(t, set)
	assertSetSize(t, set, 1)
	now = func() time.Time { return st.Add(time.Second * 11) }
	if pp, err := set.Pop(pk2, p2.id); err != nil || pp != p2 {
		t.Fatal("unable to pop ping")
	}
	assertSetSize(t, set, 0)
	if _, err := set.Pop(pk2, p2.id); err == nil {
		t.Fatal("popped a second time")
	}
	assertSetSize(t, set, 0)
}

func TestPingSetExpiryMulti(t *testing.T) {
	const timeout = 10 * time.Second

	set := NewSet(timeout)
	st := time.Now()
	now = func() time.Time { return st }

	addPing(t, set)
	addPing(t, set)
	assertSetSize(t, set, 2)
	now = func() time.Time { return st.Add(time.Second * 11) }
	addPing(t, set)
	assertSetSize(t, set, 1)
	pk4, p4 := addPing(t, set)
	pk5, p5 := addPing(t, set)
	assertSetSize(t, set, 3)

	if pp, err := set.Pop(pk4, p4.id); err != nil || pp != p4 {
		t.Fatal("unable to pop ping")
	}
	assertSetSize(t, set, 2)

	now = func() time.Time { return st.Add(time.Second * 22) }
	if _, err := set.Pop(pk5, p5.id); err == nil {
		t.Fatal("popped expired ping")
	}
	assertSetSize(t, set, 0)
}
