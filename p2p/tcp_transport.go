package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represent the remote node over a TCP established connection.
type TCPPeer struct {
	// the undrliyng conn of the peer. Whici in this case
	// is a TCP connection.
	net.Conn
	// if we dial a conn => outbound == true
	// if we accept and retriee a conn => outbound == false
	outbound bool

	Wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		Wg:       &sync.WaitGroup{},
	}
}

// Send implemetns the Peer interface and will
// write bytes to remote.
func (p *TCPPeer) Send(b []byte) (int, error) {
	return p.Conn.Write(b)
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

func DefaultOnPeer(p Peer) error {
	fmt.Printf("[TCP] new peer %s\n", p.RemoteAddr())
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
	// fmt.Println("[TCP] channel consumed")
	return t.rpcch
}

// CLose implements the Transport interface.
func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

// Dial implements the Transport interface.
func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[TCP] dial error %s\n", err)
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		log.Printf("[TCP] listen error %s\n", err)
		return err
	}

	go t.startAcceptLoop()

	log.Printf("[TCP] transport listening on %s\n", t.ListenAddr)

	return nil
}

func (t *TCPTransport) startAcceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			log.Printf("[TCP] connection closed %s\n", err)
			return
		}

		if err != nil {
			fmt.Printf("[TCP] accept error: %s", err)
		}

		go t.handleConn(conn, false)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool) {
	fmt.Printf("[TCP] new incoming connection %+v\n", conn.RemoteAddr())
	peer := NewTCPPeer(conn, outbound)
	var err error

	defer func() {
		fmt.Printf("[TCP] dropping peer conn: %s\n", conn.RemoteAddr())
		conn.Close()
	}()

	if err = t.HandshakeFunc(peer); err != nil {
		conn.Close()
		fmt.Printf("[TCP] handshake error: %s\n", err)
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

		rpc.From = conn.RemoteAddr().String()

		peer.Wg.Add(1)
		t.rpcch <- rpc
		// waiting till stream is done
		peer.Wg.Wait()
		// stream done, continuening read loop
	}
}
