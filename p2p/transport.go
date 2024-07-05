package p2p

// Interface that represents the remote node
type Peer interface {
	Close() error
}

// Anything that handles the communication between nodes in the network.
// This can use any protocol (UDP, TCP, Websockets, ...)
type Transport interface {
	ListenAndAccept() error
	Consume() <-chan RPC
	Close() error
}
