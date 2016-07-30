package pgo

import "github.com/pkmngo-odi/pogo-protos"

// WildPokemon ...
type WildPokemon struct {
	Pokemon []*protos.WildPokemon
}

// NearbyPokemon is a struct for any pokemon that are nearby
// (these pokemon will not show any latitude or longitude)
type NearbyPokemon struct {
	Pokemon []*protos.NearbyPokemon
}

// CatchablePokemon is a struct for any pokemon that are catchable
type CatchablePokemon struct {
	Pokemon []*protos.MapPokemon
}

// Catch all pokemon that are catchable
func (c *CatchablePokemon) Catch() {
	for _, poke := range c.Pokemon {
		var _ = poke
	}
}
