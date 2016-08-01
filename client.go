package pgo

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkmngo-odi/pogo-protos"
)

type Client struct {
	Auth   *Auth
	APIUrl string

	Location *Location

	Task *Task

	events   chan interface{}
	writeBuf *bytes.Buffer
}

type Msg struct {
	RequestURL string
	Requests   []*protos.Request
}

func NewClient() *Client {
	client := &Client{
		events:   make(chan interface{}, 6),
		writeBuf: new(bytes.Buffer),
	}

	client.Auth = &Auth{client: client}
	client.Location = &Location{client: client}
	client.Location.Moving = &Moving{Stop: make(chan interface{})}
	client.Task = &Task{add: make(chan JobToRun, 10), stop: make(chan interface{})}

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
	return c.Auth.Token
}

func (c *Client) SetAPIUrl(url string) {
	c.APIUrl = "https://" + url + "/rpc"
}

func (c *Client) GetAPIUrl() string {
	return c.APIUrl
}

// Send messages to the server
func (c *Client) Write(msg *Msg) (*protos.ResponseEnvelope, error) {
	jwt := &protos.RequestEnvelope_AuthInfo_JWT{
		c.Auth.Token,
		59,
	}
	auth := &protos.RequestEnvelope_AuthInfo{
		Provider: c.Auth.AuthType,
		Token:    jwt,
	}

	request := &protos.RequestEnvelope{
		StatusCode: 2,
		RequestId:  1469378659230941192,

		Requests: msg.Requests,

		Latitude:  c.Location.GetLatitudeF(),
		Longitude: c.Location.GetLongitudeF(),
		Altitude:  c.Location.GetAltitudeF(),

		AuthInfo: auth,

		Unknown12: 989,
	}
	reqProto, err := proto.Marshal(request)
	if err != nil {
		return nil, err
	}

	cookieJar, _ := cookiejar.New(nil)
	client := &http.Client{Timeout: 15 * time.Second, Jar: cookieJar}
	req, err := http.NewRequest("POST", msg.RequestURL, bytes.NewReader(reqProto))
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "niantic")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	bodyData := &protos.ResponseEnvelope{}
	err = proto.Unmarshal(body, bodyData)
	if err != nil {
		return nil, err
	}

	return bodyData, err

}
