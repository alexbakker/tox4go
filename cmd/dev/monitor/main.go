package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/alexbakker/tox4go/bootstrap"
	"github.com/alexbakker/tox4go/crypto"
	"github.com/alexbakker/tox4go/toxstatus"
	"github.com/alexbakker/tox4go/transport"
)

type node struct {
	Addr      *net.UDPAddr
	PublicKey *[crypto.PublicKeySize]byte
}

type jsonOutput struct {
	Self       []*selfInfo
	DHTFriends []*dhtFriend
	Pings      []*ping
}

type dhtFriend struct {
	PublicKey string
	Addr      string
	LastPing  string
}

type ping struct {
	PublicKey string
	ID        uint64
	Age       string
}

type selfInfo struct {
	Key   string
	Value string
}

var (
	instance *bootstrap.Node
	assets   = GetAssets()
)

func main() {
	transport, err := transport.NewUDPTransport("udp", ":33450")
	if err != nil {
		panic(err)
	}

	instance, err = bootstrap.NewNode(transport)
	if err != nil {
		panic(err)
	}
	instance.IsBootstrap = true

	bootStrapNodes, err := toxstatus.Fetch()
	if err != nil {
		panic(err)
	}

	for _, bn := range bootStrapNodes {
		err = instance.Bootstrap(bn)
		if err != nil {
			fmt.Printf("bad node: %s:%d, %s\n", bn.IP.String(), bn.Port, err.Error())
		}
	}

	//handle stop signal
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt)

	go func() {
		transErr := transport.Listen()
		if transErr != nil {
			panic(transErr)
		}
	}()

	//setup http server
	listener, err := net.Listen("tcp", ":8081")
	if err != nil {
		log.Fatalf("error in net.Listen: %s", err.Error())
	}
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", handleHTTPRequest)
	serveMux.HandleFunc("/json", handleJSONRequest)
	go func() {
		err := http.Serve(listener, serveMux)
		if err != nil {
			log.Printf("http server error: %s\n", err.Error())
			interruptChan <- os.Interrupt
		}
	}()

	for _ = range interruptChan {
		fmt.Printf("killing node\n")
		transport.Stop()
		listener.Close()
		break
	}
}

func handleJSONRequest(w http.ResponseWriter, r *http.Request) {
	/*friends := core.DHT.Friends()
	newFriends := make([]*dhtFriend, len(friends))
	for i, friend := range friends {
		newFriends[i] = &dhtFriend{
			PublicKey: hex.EncodeToString(friend.PublicKey[:]),
			Addr:      friend.Addr.String(),
			LastPing:  fmt.Sprintf("%.0fs", getSecondsSince(friend.LastPing)),
		}
	}*/

	pings := instance.Pings()
	newPings := make([]*ping, len(pings.List))
	for i, p := range pings.List {
		newPings[i] = &ping{
			PublicKey: hex.EncodeToString(p.PublicKey[:]),
			ID:        p.ID,
			Age:       fmt.Sprintf("%.0fs", getSecondsSince(p.Time)),
		}
	}

	output := jsonOutput{DHTFriends: []*dhtFriend{}, Pings: newPings, Self: getSelfInfo()}

	bytes, err := json.Marshal(output)
	if err != nil {
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Write(bytes)
}

func handleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path[1:]
	if r.URL.Path == "/" {
		urlPath = "index.html"
	}

	asset, exists := assets[urlPath]
	if !exists {
		http.Error(w, http.StatusText(404), 404)
	} else {
		w.Write(asset)
	}
}

func getSelfInfo() []*selfInfo {
	info := []*selfInfo{
		&selfInfo{"PublicKey", hex.EncodeToString(instance.Ident.PublicKey[:])},
		&selfInfo{"Routines", fmt.Sprintf("%d", runtime.NumGoroutine())},
		&selfInfo{"Version", runtime.Version()},
	}

	return info
}

func getSecondsSince(stamp time.Time) float64 {
	return time.Now().Sub(stamp).Hours() * 60 * 60
}
