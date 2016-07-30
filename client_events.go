package pgo

import "github.com/pkmngo-odi/pogo-protos"

// This event is emitted when there is an error that will cause the bot to no longer run
type FatalErrorEvent struct {
	Err error
}

type SemiErrorEvent struct {
	Err error
}

type ResponseEnvelope struct {
	Resp *protos.ResponseEnvelope
}
