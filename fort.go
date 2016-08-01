package pgo

import (
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/pkmngo-odi/pogo-protos"
)

type Forts struct {
	Forts []*Fort
}

type Fort struct {
	*protos.FortData
}

// Search all forts
func (f *Fort) Search(client *Client) {
	log.Println("FORT", f)

	// If fort is not lootable, skip it
	if !f.Enabled {
		return
	}
	// Move over to the fort
	client.Location.Move(&Location{
		Latitude:  f.Latitude,
		Longitude: f.Longitude,
	}, RUNNING_SPEED)
	fortData := &protos.FortSearchMessage{
		FortId:          f.Id,
		PlayerLatitude:  client.Location.GetLatitudeF(),
		PlayerLongitude: client.Location.GetLongitudeF(),
		FortLatitude:    f.Latitude,
		FortLongitude:   f.Longitude,
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
