package p2p

import (
	"fmt"
	"log"
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

// Implementation of Peer interface
func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

func DefaultOnPeer(Peer Peer) error {
	fmt.Println("[TCP] on peer")
	return nil
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch    chan RPC

	peerLock sync.RWMutex
	peers    map[net.Addr]Peer
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch:            make(chan RPC),
	}
}

// Consume implements the Transport interface, which will
// return read-only channel for reading the incoming
// messages recieved from another peer
func (t *TCPTransport) Consume() <-chan RPC {
	fmt.Println("transport channel consumed")
	return t.rpcch
}

// CLose implemetns the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on %s\n", t.ListenAddr)

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

func (t *TCPTransport) handleConn(conn net.Conn) {
	fmt.Printf("[TCP] new incoming connection %+v\n", conn)
	peer := NewTCPPeer(conn, true)
	var err error

	defer func() {
		fmt.Printf("[TCP] dropping peer conn: %s", err)
		conn.Close()
	}()

	if err = t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("[TCP] handshake error: %s", err)
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	rpc := RPC{}
	for {
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			fmt.Printf("[TCP] decoding error %s\n", err)
			return
		}

		rpc.From = conn.RemoteAddr()
		t.rpcch <- rpc
	}
}
