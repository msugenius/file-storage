package p2p

import "io"

type Decoder interface {
	Decode(io.Reader, any) error
}

type Encoder interface {
	Encode(io.Reader) error
}
