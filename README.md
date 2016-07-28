# Pokemon-Golang
Pokemon-Golang is soon to be a fully automated Pokemon-GO bot.

Driven by Go's channels, concurrency, and amazingly fast protobuf interacting, this is aimed to be the fastest 
implementation of Pokemon-Gos API for easily building a fast, reliable, automated bot.

- [x] Authentication with Google
- [x] Authentication with Pokemon Trainers Club
- [ ] Visit Pokestops
- [ ] Farm Pokestops
- [ ] GPS Spoofing interpolated as Human
- [ ] Catch Pokemon
- [ ] Catch Only Certain Pokemon
- [ ] Remove Excess Items When Over X Quantities
- [ ] Transfer Flagged Pokemon
- [ ] Auto Hatch Eggs
- [ ] Evolve Pokemon

# Example Usage


```
package main

import (
	"log"
	"reflect"

	"github.com/Gacnt/pokemon-go"
)

func main() {
	client := pgo.NewClient()
	loginInfo := new(pgo.LogOnDetails)
	loginInfo.Username = "Username"
	loginInfo.Password = "Password"
	loginInfo.AuthType = "google"

        // Send Initial Request
	client.Auth.GetToken(loginInfo)

        // Begin Event Listener
	for event := range client.Events() {
		switch e := event.(type) {
		case *pgo.AuthedEvent:
			client.Auth.SetAuthToken(e.AuthToken)
			client.Auth.Login()
			log.Println("Set Token")
		case *pgo.LoggedOnEvent:
			client.SetAPIUrl(e.APIUrl)
			log.Println(client.GetAPIUrl())
		case *pgo.FatalErrorEvent:
			log.Println(e.Err)
		default:
			log.Printf("Uncaught Event was fired: \nType: %v\n Value: %+v", reflect.TypeOf(e), e)
		}
	}
}
```
