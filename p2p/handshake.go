package p2p

// Handshake function is
type HandshakeFunc func(Peer) error

// No handshake would be done
func DefaultHandshakeFunc(Peer) error { return nil }
