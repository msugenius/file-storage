package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/msugenius/file-storage/p2p"
)

type FileServerOpts struct {
	StorageRoot       string
	PathTransformFunc PathTransformFunc
	Transport         p2p.Transport
	TCPTransportsOpts p2p.TCPTransportOpts
	BootstrapNodes    []string
}

type FileServer struct {
	FileServerOpts

	peerLock sync.Mutex
	peers    map[string]p2p.Peer

	store  *Store
	quitch chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer {
	storeOpts := StoreOpts{
		Root:              opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		store:          NewStore(storeOpts),
		FileServerOpts: opts,
		quitch:         make(chan struct{}),
		peers:          make(map[string]p2p.Peer),
	}
}

func (s *FileServer) OnPeer(p p2p.Peer) error {
	log.Println("connection to peer")
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s\n", p.RemoteAddr())

	return nil
}

// func (s *FileServer) handleMessage(msg *Message) error {
// switch v := msg.Payload.(type) {
// case *DataMessage:
// 	fmt.Printf("recieved data %+v\n", v)
// }

// return nil
// }

func (s *FileServer) bootstrapNetwork() error {
	if len(s.BootstrapNodes) == 0 {
		return nil
	}

	for _, addr := range s.BootstrapNodes {
		go func() {
			fmt.Println("attempting to connect remote host: ", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Printf("dial error %s\n", err)
			}
		}()
	}

	return nil
}

type Message struct {
	From    string
	Payload any
}

type DataMessage struct {
	Key  string
	Data []byte
}

func (s *FileServer) Start() error {
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}

func (s *FileServer) StoreData(key string, r io.Reader) error {
	// 1. Store this data to disk
	// 2. Broadcast this file to all known peers

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: []byte("asdasdas"),
	}
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers {
		if n, err := peer.Send(buf.Bytes()); err != nil {
			fmt.Println(n)
			return err
		}
	}

	time.Sleep(3 * time.Second)

	payload := []byte("LARGE FILE")
	for _, peer := range s.peers {
		if n, err := peer.Send(payload); err != nil {
			fmt.Println(n)
			return err
		}
	}

	return nil

	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	// if err := s.store.Write(key, tee); err != nil {
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key:  key,
	// 	Data: buf.Bytes(),
	// }

	// fmt.Printf("[SERVER] payload %+v\n", p)

	// return s.broadcast(&Message{
	// 	From:    "todo",
	// 	Payload: p,
	// })
}

func (s *FileServer) Stop() {
	fmt.Println("stopping the server")
	close(s.quitch)
}

func (s *FileServer) broadcast(msg *Message) error {
	peers := []io.Writer{}
	for _, peer := range s.peers {
		peers = append(peers, peer)
	}

	fmt.Printf("peers %+v\n", peers)

	mw := io.MultiWriter(peers...)

	return gob.NewEncoder(mw).Encode(msg)
}

func (s *FileServer) loop() {
	defer func() {
		log.Println("file server stopped due to quit action")
		s.Transport.Close()
	}()

	for {
		select {
		case rpc := <-s.Transport.Consume():
			var msg Message
			err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg)
			if err != nil {
				log.Println("[SERVER] loop ", err)
				return
			}

			peer, ok := s.peers[rpc.From]
			if !ok {
				panic("peer not found in peer map")
			}

			b := make([]byte, 1024)
			if _, err := peer.Read(b); err != nil {
				panic(err)
			}

			fmt.Printf("recv %+v\n", string(msg.Payload.([]byte)))
			// err = s.handleMessage(&msg)
			// if err != nil {
			// 	log.Println("[SERVER] loop ", err)
			// 	return
			// }

		case <-s.quitch:
			return
		}
	}
}
