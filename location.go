package pgo

import (
	"math"
	"strconv"
	"time"

	"github.com/golang/geo/s2"
	"github.com/kellydunn/golang-geo"
)

const (
	WALKING_SPEED float64 = 1.3  // meters per second
	RUNNING_SPEED float64 = 6.4  // meters per second
	BIKING_SPEED  float64 = 12.3 // meters per second
	DRIVING_SPEED float64 = 16.7 // meters per second
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

func (l *Location) Move(newLoc *Location, speed float64) {
	R := 6371000.0 // Kilometers
	lat1 := l.Latitude * math.Pi / 180
	lat2 := l.Longitude * math.Pi / 180
	diffLatRad := (newLoc.Latitude - l.Latitude) * math.Pi / 180
	diffLonRad := (newLoc.Longitude - l.Longitude) * math.Pi / 180

	a := math.Sin(diffLatRad/2)*math.Sin(diffLatRad/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(diffLonRad/2)*math.Sin(diffLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distanceToMove := R * c // Distance to travel in a straight line, in Kilometers
	distanceActuallyTravelled := 0.0
	ticker := time.Tick(time.Second * 1)
	for stop := false; !stop; {
		select {
		case <-ticker:
			deltaLat := (newLoc.Latitude - l.Latitude) * (distanceActuallyTravelled / distanceToMove)
			deltaLng := (newLoc.Longitude - l.Longitude) * (distanceActuallyTravelled / distanceToMove)
			newLat := l.GetLatitudeF() + deltaLat
			newLng := l.GetLongitudeF() + deltaLng
			l.SetLatitude(Locnum(newLat))
			l.SetLongitude(Locnum(newLng))
			if distanceActuallyTravelled >= distanceToMove {
				// Bot may be traveling too fast to get an accurate landing and
				// might overshoot the location, once overshot
				// default to the destination.
				l.SetLatitude(Locnum(newLoc.Latitude))
				l.SetLongitude(Locnum(newLoc.Longitude))
				stop = true
			}
			distanceActuallyTravelled = distanceActuallyTravelled + speed
		}
	}

}
