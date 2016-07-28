# Pokemon-Golang
Pokemon-Golang is not necessarily a bot, but it provides the set of tools in a very easy to use way to create your own 
bot that is custom to your needs. Examples will be released a long the way, and my own version of a bot built using this 
very simple API

Driven by Go's channels, concurrency, and amazingly fast protobuf interacting, this is aimed to be the fastest 
implementation of Pokemon-Gos API for easily building a fast, reliable, automated bot.

- [x] Authentication with Google
- [x] Authentication with Pokemon Trainers Club
- [x] View Pokemon / Forts Nearby You
- [ ] Visit Pokestops
- [ ] Farm Pokestops
- [ ] GPS Spoofing interpolated as Human
- [ ] Catch Pokemon
- [ ] Catch Only Certain Pokemon
- [ ] Remove Excess Items When Over X Quantities
- [ ] Transfer Flagged Pokemon
- [ ] Auto Hatch Eggs
- [ ] Evolve Pokemon
- [ ] Make Errors More Descriptive For Easier Debugging

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
	loginInfo.AuthType = "ptc"

	client.Auth.GetToken(loginInfo)
	for event := range client.Events() {
		switch e := event.(type) {
		case *pgo.AuthedEvent:
			client.Auth.SetAuthToken(e.AuthToken)
			client.Location.SetByName("New York")
			client.Auth.Login()
		case *pgo.LoggedOnEvent:
			client.SetAPIUrl(e.APIUrl)
			go func() {
				pgo.GetMapData(client)
			}()
		case *pgo.LocationSet:
			log.Println("Location has been set")
			log.Printf("%+v", *e.Location)
		case *pgo.NearbyPokemonEvent:
			log.Println(e)
		case *pgo.WildPokemonEvent:
			log.Println(e)
		case *pgo.CatchablePokemonEvent:
			log.Println(e)
		case *pgo.FatalErrorEvent:
			log.Println(e.Err)
		default:
			log.Printf("Uncaught Event was fired: \nType: %v\n Value: %+v", reflect.TypeOf(e), e)
		}
	}
}
```

For documentation please visit https://godoc.org/github.com/Gacnt/pokemon-golang
