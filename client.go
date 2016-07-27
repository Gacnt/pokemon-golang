package pgo

import "bytes"

type Client struct {
	Auth      *Auth
	AuthToken string

	events chan interface{}

	writeBuf *bytes.Buffer
}

func NewClient() *Client {
	client := &Client{
		events:   make(chan interface{}, 3),
		writeBuf: new(bytes.Buffer),
	}

	client.Auth = &Auth{client: client}

	return client
}

func (c *Client) Events() <-chan interface{} {
	return c.events
}

func (c *Client) Emit(event interface{}) {
	c.events <- event
}

// Helper function to return the Authentication token recieved
// at login
func (c *Client) Token() string {
	return c.AuthToken
}
