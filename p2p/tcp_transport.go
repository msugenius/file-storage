package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer represent the remote node over a TCP established connection.
type TCPPeer struct {
	// the undrliyng conn of the peer
	conn net.Conn
	// if we dial a conn => outbound == true
	// if we accept and retriee a conn => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}
}

type TCPTransport struct {
	listenAddress string
	listener      net.Listener
	handshakeFunc HandshakeFunc
	decoder       Decoder

	peerLock sync.RWMutex
	peers    map[net.Addr]Peer
}

func NewTCPTransport(listenAddr string) *TCPTransport {
	return &TCPTransport{
		listenAddress: listenAddr,
		handshakeFunc: NOPHandshakeFunc,
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.listenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s", err)
		}

		go t.handleConn(conn)
	}
}

type Temp struct{}

func (t *TCPTransport) handleConn(conn net.Conn) {
	peer := NewTCPPeer(conn, true)

	if err := t.handshakeFunc(peer); err != nil {
		fmt.Printf("[TCP] handshake error: %s", err)
	}

	// Read loop
	msg := &Temp{}
	for {
		if err := t.decoder.Decode(conn); err != nil {
			fmt.Printf("[TCP] decoding error %s", err)
			continue
		}
	}

	fmt.Printf("new incoming connection %+v\n", conn)
}
