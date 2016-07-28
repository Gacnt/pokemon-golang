package pgo

import "github.com/pkmngo-odi/pogo-protos"

type MapDataEvent struct{}

type NearbyPokemonEvent struct {
	Pokemons []*protos.NearbyPokemon
}

type WildPokemonEvent struct {
	Pokemons []*protos.WildPokemon
}

type CatchablePokemonEvent struct {
	Pokemons []*protos.MapPokemon
}

type FortEvent struct {
	Forts []*protos.FortData
}

type FortSummariesEvent struct {
	Summaries []*protos.FortSummary
}

type MapObjectsEvent struct {
	MapCells []*protos.MapCell
}
