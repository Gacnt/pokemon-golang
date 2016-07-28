package pgo

import (
	"strconv"

	"github.com/golang/geo/s2"
	"github.com/kellydunn/golang-geo"
)

type Location struct {
	client *Client

	Name string

	Latitude  float64
	Longitude float64
	Altitude  float64
}

type Locnum float64

func (l *Locnum) String() string {
	return strconv.FormatFloat(float64(*l), 'f', 7, 64)
}

func (l *Location) SetLatitude(lat Locnum) {
	l.Latitude = float64(lat)
}

func (l *Location) SetLongitude(lon Locnum) {
	l.Longitude = float64(lon)
}

func (l *Location) SetAltitude(alt Locnum) {
	l.Altitude = float64(alt)
}

func (l *Location) GetLatitude() Locnum {
	return Locnum(l.Latitude)
}

func (l *Location) GetLongitude() Locnum {
	return Locnum(l.Longitude)
}

func (l *Location) GetAltitude() Locnum {
	return Locnum(l.Altitude)
}

func (l *Location) GetLatitudeF() float64 {
	return l.Latitude
}

func (l *Location) GetLongitudeF() float64 {
	return l.Longitude
}

func (l *Location) GetAltitudeF() float64 {
	return l.Altitude
}

func (l *Location) SetByLocation(name string) {
	geoLoc := &geo.GoogleGeocoder{}

	p, err := geoLoc.Geocode(name)
	if err != nil {
		l.client.Emit(&SemiErrorEvent{err})
	}

	l.Name = name
	l.Latitude = p.Lat()
	l.Longitude = p.Lng()
	l.Altitude = 0.0

	l.client.Emit(&LocationSet{l})
}

func (l *Location) SetByCoords(lat, lng, alt float64) {
	p := geo.NewPoint(lat, lng)

	geo := &geo.GoogleGeocoder{}
	name, err := geo.ReverseGeocode(p)
	if err != nil {
		l.client.Emit(&SemiErrorEvent{err})
	}

	l.Name = name
	l.Altitude = alt
	l.Longitude = lng
	l.Latitude = lat

	l.client.Emit(&LocationSet{l})
}

func (l *Location) GetNeighbors() []uint64 {
	ll := s2.LatLngFromDegrees(l.Latitude, l.Longitude)
	cid := s2.CellIDFromLatLng(ll).Parent(15)

	walker := []uint64{uint64(cid)}
	next := cid.Next()
	prev := cid.Prev()
	for i := 0; i < 10; i++ {
		walker = append(walker, uint64(next))
		walker = append(walker, uint64(prev))
		next = next.Next()
		prev = prev.Prev()
	}

	return walker
}
