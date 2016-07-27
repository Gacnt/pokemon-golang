package pgo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/Gacnt/gpsoaauth"
	"github.com/pkmngo-odi/pogo-protos"
)

const (
	loginURL     string = "https://sso.pokemon.com/sso/login?service=https%3A%2F%2Fsso.pokemon.com%2Fsso%2Foauth2.0%2FcallbackAuthorize"
	loginOAuth   string = "https://sso.pokemon.com/sso/oauth2.0/accessToken"
	androidID    string = "9774d56d682e549c"
	oAuthService string = "audience:server:client_id:848232511240-7so421jotr2609rmqakceuu1luuq0ptb.apps.googleusercontent.com"
	app          string = "com.nianticlabs.pokemongo"
	clientSIG    string = "321187995bc7cdc2b5fc91b11a96e2baa8602c62"
)

type Auth struct {
	client *Client

	AuthType string
	Token    string
}

type LogOnDetails struct {
	Username string
	Password string
	AuthType string
}

// Helper function to set the authentication token
func (a *Auth) SetAuthToken(token string) {
	a.client.Auth.Token = token
}

func (a *Auth) Login() {
	req := []*protos.Request{
		&protos.Request{RequestType: 2},
		&protos.Request{RequestType: 126},
		&protos.Request{RequestType: 4},
		&protos.Request{RequestType: 129},
		&protos.Request{RequestType: 5},
	}

	a.client.Write(&Msg{
		RequestURL: "https://pgorelease.nianticlabs.com/plfe/rpc",
		Requests:   req,
	})
}

func (a *Auth) GetToken(details *LogOnDetails) {
	if details.AuthType == "ptc" {
		data, err := authWithPTC(details)
		if err != nil {
			a.client.Emit(&FatalErrorEvent{err})
		}
		a.AuthType = "ptc"
		a.client.Emit(&AuthedEvent{data}) // Login Was Successful
	} else if details.AuthType == "google" {
		data, err := authWithGoogle(details)
		if err != nil {
			a.client.Emit(&FatalErrorEvent{err})
		}
		a.AuthType = "google"
		a.client.Emit(&AuthedEvent{data}) // Login Was Successful
	} else {
		log.Printf("[!] For LogOnDetails, you must set AuthType to either `ptc` or `google`, recieved: %s", details.AuthType)
	}
}

func authWithPTC(details *LogOnDetails) (string, error) {
	// Initiate HTTP Client / Cookie JAR

	jar, err := cookiejar.New(nil)
	if err != nil {
		return "", fmt.Errorf("Failed to create new cookiejar for client")
	}
	newClient := &http.Client{Jar: jar, Timeout: 15 * time.Second}

	// First Request

	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return "", fmt.Errorf("Failed to authenticate with Pokemon Trainers Club\n Details: \n\n Username: %s\n Password: %s\n AuthType: %s\n", details.Username, details.Password, details.AuthType)
	}
	req.Header.Set("User-Agent", "niantic")
	firstResp, err := newClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to send intial handshake: Possible wrong Username or Password", err)
	}
	respJSON := make(map[string]string)
	err = json.NewDecoder(firstResp.Body).Decode(&respJSON)
	if err != nil {
		return "", fmt.Errorf("Failed to decode JSON Body: %v", err)
	}

	defer firstResp.Body.Close()

	// Second Request

	form := url.Values{}
	form.Add("lt", respJSON["lt"])
	form.Add("execution", respJSON["execution"])
	form.Add("_eventId", "submit")
	form.Add("username", details.Username)
	form.Add("password", details.Password)
	req, err = http.NewRequest("POST", loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("Failed to send second request authing with PTC: %v", err)
	}
	req.Header.Set("User-Agent", "niantic")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	secResp, err := newClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to send second request authing with PTC: %v", err)
	}

	ticket := secResp.Request.URL.String()

	if strings.Contains(ticket, "ticket") {
		ticket = strings.Split(ticket, "ticket=")[1]
	} else {
		return "", fmt.Errorf("Failed could not get the Ticket from the second request\n.. Possible wrong Username or Password")
	}
	defer secResp.Body.Close()

	// Third Request

	form = url.Values{}
	form.Add("client_id", "mobile-app_pokemon-go")
	form.Add("redirect_uri", "https://www.nianticlabs.com/pokemongo/error")
	form.Add("client_secret", "w8ScCUXJQc6kXKw8FiOhd8Fixzht18Dq3PEVkUCP5ZPxtgyWsbTvWHFLm2wNY0JR")
	form.Add("grant_type", "refresh_token")
	form.Add("code", ticket)
	req, err = http.NewRequest("POST", loginOAuth, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("Failed to send the third request authing with PTC: %v", err)
	}
	req.Header.Add("User-Agent", "niantic")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	thirdResp, err := newClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Failed to send the third request authing with PTC: %v", err)
	}
	defer thirdResp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(thirdResp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to decode the body of the third request")
	}

	body := string(bodyBytes)

	if strings.Contains(body, "token=") {
		token := strings.Split(body, "token=")[1]
		token = strings.Split(token, "&")[0]
		return token, nil
	} else {
		return "", fmt.Errorf("Failed to get the token on the third request \nBody:\n\n%v", body)
	}

}

func authWithGoogle(details *LogOnDetails) (string, error) {
	_, masterToken, err := gpsoauth.Login(details.Username, details.Password, androidID)
	if err != nil {
		return "", fmt.Errorf("[!] Failed to Login with Google\nUsername: %s\nPassword: %s\nAndroidID: %s", details.Username, details.Password, androidID)
	}
	body, err := gpsoauth.OAuth(details.Username, masterToken, androidID, oAuthService, app, clientSIG)
	if err != nil {
		return "", fmt.Errorf("[!] Failed to Login with Google\nUsername: %s\nPassword: %s\nAndroidID: %s", details.Username, details.Password, androidID)
	}

	if _, ok := body["Auth"]; !ok {
		return "", fmt.Errorf("[!] Missing AUTH. Could be an incorrect Email or Password, or 2 step authentication failure. (This package does not support 2 step auth)")
	}

	return body["Auth"], nil
}
