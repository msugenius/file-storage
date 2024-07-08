package main

import (
	"bytes"
	"log"
	"time"

	"github.com/msugenius/file-storage/p2p"
)

func makeServer(listenAddr string, root string, nodes []string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		Decoder:       p2p.DefaultDecoder{},
		HandshakeFunc: p2p.DefaultHandshakeFunc,
	}
	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot:       root,
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)

	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main() {
	s1 := makeServer(":3000", "data1", []string{})
	s2 := makeServer(":4000", "data2", []string{":3000"})

	go func() { log.Fatal(s1.Start()) }()
	time.Sleep(1 * time.Second)
	go func() { log.Fatal(s2.Start()) }()
	time.Sleep(3 * time.Second)

	data := bytes.NewReader([]byte("very very very large file"))
	err := s2.StoreData("my-private-data", data)
	if err != nil {
		panic(err)
	}

	select {}
}
