package main

import (
	"fmt"

	"github.com/Impyy/tox4go/bootstrap"
	"github.com/Impyy/tox4go/transport"
)

func main() {
	transport, err := transport.NewUDPTransport("udp", ":33450")
	if err != nil {
		panic(err)
	}

	node, err := bootstrap.NewNode(transport)
	if err != nil {
		panic(err)
	}
	node.IsBootstrap = true

	bootStrapNodes, err := grabNodes()
	if err != nil {
		panic(err)
	}

	for _, bn := range bootStrapNodes {
		err = node.Bootstrap(bn)
		if err != nil {
			fmt.Printf("bad node: %s:%d, %s\n", bn.IP.String(), bn.Port, err.Error())
		}
	}

	err = transport.Listen()
	if err != nil {
		panic(err)
	}
}
