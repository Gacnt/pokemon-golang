package pgo

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/pkmngo-odi/pogo-protos"
)

type Forts struct {
	Fort []*protos.FortData
}

// Search all forts
func (f *Forts) Search(client *Client) {
	for _, fort := range f.Fort {
		log.Println("FORT", fort)

		// If fort is not lootable, skip it
		if !fort.Enabled {
			continue
		}
		// Move over to the fort
		client.Location.Move(&Location{
			Latitude:  fort.Latitude,
			Longitude: fort.Longitude,
		}, RUNNING_SPEED)
		fortData := &protos.FortSearchMessage{
			FortId:          fort.Id,
			PlayerLatitude:  client.Location.GetLatitudeF(),
			PlayerLongitude: client.Location.GetLongitudeF(),
			FortLatitude:    fort.Latitude,
			FortLongitude:   fort.Longitude,
		}

		fortProto, err := proto.Marshal(fortData)
		if err != nil {
			client.Emit(&SemiErrorEvent{err})
		}

		// At fort now search the fort
		req := []*protos.Request{
			&protos.Request{
				RequestType:    101,
				RequestMessage: fortProto,
			},
		}
		resp, err := client.Write(&Msg{
			RequestURL: client.APIUrl,
			Requests:   req,
		})
		if err != nil {
			client.Emit(&SemiErrorEvent{err})
		}
		fortRespMessage := &protos.FortSearchResponse{}

		if len(resp.Returns) > 0 {
			err = proto.Unmarshal(resp.Returns[0], fortRespMessage)
			if err != nil {
				client.Emit(&SemiErrorEvent{err})
			}

			client.Emit(&FortSearchedEvent{fortRespMessage})
		} else {
			client.Emit(&ResponseEnvelope{resp})
		}

	}
}
