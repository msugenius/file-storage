package p2p

// Handshake function is
type HandshakeFunc func(Peer) error

// No handshake would be done
func NOPHandshakeFunc(Peer) error { return nil }
