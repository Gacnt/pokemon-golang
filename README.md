# Pokemon-Golang
Pokemon-Golang is not necessarily a bot, but it provides the set of tools in a very easy to use way to create your own 
bot that is custom to your needs. Examples will be released a long the way, and my own version of a bot built using this 
very simple API

Driven by Go's channels, concurrency, and amazingly fast protobuf interacting, this is aimed to be the fastest 
implementation of Pokemon-Gos API for easily building a fast, reliable, automated bot.

- [x] Authentication with Google
- [x] Authentication with Pokemon Trainers Club
- [x] View Pokemon / Forts Nearby You
- [x] Visit Pokestops
- [x] Farm Pokestops
- [x] GPS Spoofing interpolated as Human
- [x] Tasking system
- [ ] Catch Pokemon
- [ ] Catch Only Certain Pokemon
- [ ] Remove Excess Items When Over X Quantities
- [ ] Transfer Flagged Pokemon
- [ ] Auto Hatch Eggs
- [ ] Evolve Pokemon
- [ ] Make Errors More Descriptive For Easier Debugging

TODO Extras:

- Add priority system to the tasking system to deem certain things more important so they run before other things (e.g. catch a rare pokemon over visiting a pokestop)

# Example Usage


```go
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
			client.Location.SetByLocation("New York")
			client.Auth.Login()
		case *pgo.LoggedOnEvent:
			client.SetAPIUrl(e.APIUrl)
			/*go func() {
				pgo.GetMapData(client)
			}()*/
		case *pgo.LocationSet:
			log.Println("Location has been set")
			log.Printf("%+v", *e.Location)
		case *pgo.MovingUpdateEvent:
			log.Println("Bot is walking")
			log.Println(e.DistanceTravelled, e.DistanceTotal)
		case *pgo.MovingDirectionChangedEvent:
			log.Println("Changed")
		case *pgo.FortSearchedEvent:
			log.Printf("%+v", e.Result)
		case *pgo.FortEvent:
			log.Println("Fort Event----------------------")

			// Task will get pushed to the tasker and executed in the order they are recieved
			client.Task.AddFunc(func() {
				e.Forts.Search(client)
			})
		case *pgo.FatalErrorEvent:
			log.Println(e.Err)
		default:
		}
	}
}
```

For documentation please visit https://godoc.org/github.com/Gacnt/pokemon-golang

This API structure was heavily inspired by [Philipp15b's Go Steam API](https://github.com/Philipp15b/go-steam)
