package p2p

// Interface that represents the remote node
type Peer interface{}

// Anything that handles the communication between nodes in the network.
// This can use any protocol (UDP, TCP, Websockets, ...)
type Transport interface {
	ListenAndAccept() error
}
