package pgo

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkmngo-odi/pogo-protos"
)

func GetMapData(client *Client) {
	mo := &protos.GetMapObjectsMessage{
		CellId:           client.Location.GetNeighbors(),
		SinceTimestampMs: make([]int64, 21),
		Latitude:         client.Location.GetLatitudeF(),
		Longitude:        client.Location.GetLongitudeF(),
	}

	moProto, err := proto.Marshal(mo)
	if err != nil {
		client.Emit(&SemiErrorEvent{err})
	}

	inv := &protos.GetInventoryMessage{
		LastTimestampMs: time.Now().UnixNano() / 1000000,
	}

	invProto, err := proto.Marshal(inv)
	if err != nil {
		client.Emit(&SemiErrorEvent{err})
	}

	dl := &protos.DownloadSettingsMessage{
		Hash: "05daf51635c82611d1aac95c0b051d3ec088a930",
	}
	dlProto, err := proto.Marshal(dl)
	if err != nil {
		client.Emit(&SemiErrorEvent{err})
	}

	req := []*protos.Request{
		&protos.Request{
			RequestType:    protos.RequestType_GET_MAP_OBJECTS,
			RequestMessage: moProto,
		},
		&protos.Request{
			RequestType: protos.RequestType_GET_HATCHED_EGGS,
		},
		&protos.Request{
			RequestType:    protos.RequestType_GET_INVENTORY,
			RequestMessage: invProto,
		},
		&protos.Request{
			RequestType: protos.RequestType_CHECK_AWARDED_BADGES,
		},
		&protos.Request{
			RequestType:    protos.RequestType_DOWNLOAD_SETTINGS,
			RequestMessage: dlProto,
		},
	}

	resp, err := client.Write(&Msg{
		RequestURL: client.APIUrl,
		Requests:   req,
	})

	if err != nil || resp.StatusCode != 1 || len(resp.Returns) == 0 {
		client.Emit(&SemiErrorEvent{err})
		return
	}

	respMapObj := &protos.GetMapObjectsResponse{}
	err = proto.Unmarshal(resp.Returns[0], respMapObj)

	client.Emit(&MapObjectsEvent{respMapObj.MapCells})
	for _, m := range respMapObj.MapCells {
		if len(m.NearbyPokemons) > 0 {
			nearby := &NearbyPokemon{
				m.NearbyPokemons,
			}
			client.Emit(&NearbyPokemonEvent{nearby})
		}
		if len(m.WildPokemons) > 0 {
			wild := &WildPokemon{
				m.WildPokemons,
			}
			client.Emit(&WildPokemonEvent{wild})
		}
		if len(m.CatchablePokemons) > 0 {
			catchable := &CatchablePokemon{
				m.CatchablePokemons,
			}
			client.Emit(&CatchablePokemonEvent{catchable})
		}
		if len(m.Forts) > 0 {
			fortsSl := []*Fort{}
			gyms := []*protos.FortData{}

			for _, f := range m.Forts {
				if f.Type.String() == "GYM" {
					gyms = append(gyms, f)
				} else if f.Type.String() == "CHECKPOINT" {
					fortsSl = append(fortsSl, &Fort{FortData: f})
				}
			}
			gym := &Gyms{
				gyms,
			}
			client.Emit(&FortEvent{fortsSl})
			client.Emit(&GymEvent{gym})
		}
		if len(m.FortSummaries) > 0 {
			client.Emit(&FortSummariesEvent{m.FortSummaries})
		}
	}
}
