package pgo

// This event is emitted when there is an error that will cause the bot to no longer run
type FatalErrorEvent struct {
	Err error
}

type SemiErrorEvent struct {
	Err error
}
