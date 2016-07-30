package pgo

import "github.com/pkmngo-odi/pogo-protos"

type MapDataEvent struct{}

type NearbyPokemonEvent struct {
	Pokemons *NearbyPokemon
}

type WildPokemonEvent struct {
	Pokemons *WildPokemon
}

type CatchablePokemonEvent struct {
	Pokemons *CatchablePokemon
}

type FortEvent struct {
	Forts *Forts
}

type GymEvent struct {
	Gyms *Gyms
}

type FortSummariesEvent struct {
	Summaries []*protos.FortSummary
}

type MapObjectsEvent struct {
	MapCells []*protos.MapCell
}
