package p2p

import "net"

// Interface that represents the remote node
type Peer interface {
	net.Conn
	Send([]byte) (int, error)
}

// Anything that handles the communication between nodes in the network.
// This can use any protocol (UDP, TCP, Websockets, ...)
type Transport interface {
	Close() error
	Consume() <-chan RPC
	Dial(string) error
	ListenAndAccept() error
}
