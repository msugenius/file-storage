package main

import (
	"fmt"
	"log"

	"github.com/msugenius/file-storage/p2p"
)

func main() {
	tcpOpts := p2p.TCPTransportsOpts{
		ListenAddr:    ":3000",
		HandshakeFunc: p2p.DefaultHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
		OnPeer:        p2p.DefaultOnPeer,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	go func() {
		for {
			msg := <-tr.Consume()
			fmt.Printf("%+v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select {}
}
