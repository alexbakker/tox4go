package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/alexbakker/tox4go/bootstrap"
	"github.com/alexbakker/tox4go/toxstatus"
	"github.com/alexbakker/tox4go/transport"
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

	bootStrapNodes, err := toxstatus.GetNodes(context.Background())
	if err != nil {
		panic(err)
	}

	for _, bn := range bootStrapNodes {
		err = node.Bootstrap(bn)
		if err != nil {
			fmt.Printf("bad node: %s:%d, %s\n", bn.IP.String(), bn.Port, err.Error())
		}
	}

	//handle stop signal
	interruptChan := make(chan os.Signal)
	signal.Notify(interruptChan, os.Interrupt)

	go func() {
		err := transport.Listen()
		if err != nil {
			panic(err)
		}
	}()

	for _ = range interruptChan {
		fmt.Printf("killing node\n")
		transport.Stop()
		break
	}
}
