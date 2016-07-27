package pgo

// This Event is fired once a successful authentication has been made
// either by connecting with Google, or Pokemon Trainer Club
type LoggedOnEvent struct {
	APIUrl string
}

type AuthedEvent struct {
	AuthToken string
}
