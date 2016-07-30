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

	Moving *Moving
}

type Moving struct {
	IsMoving          bool
	Distance          float64
	DistanceTravelled float64
	Stop              chan interface{}
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

// Teleport preferably a short distance, teleporting too far
// will probably result in the bot getting a soft lock which
// may result in unexpected and unidentifed behaviours like
// releasing Godzilla or worse yet, a Gyrados
func (l *Location) Teleport(newLoc *Location) {

	l.SetLatitude(Locnum(newLoc.Latitude))
	l.SetLongitude(Locnum(newLoc.Longitude))

	l.client.Emit(&MovingDoneEvent{})
}

// This will call the bot to move to a new location
// Note: calling this while the bot is already moving will cause its currently set
// destination to change to the new location.
// If you want your bot to visit both spots, wait until the `MovingDoneEvent` is fired before calling
// `Move` again
// This function WILL BLOCK THE EVENT LOOP if you want to move somewhere I suggest
// running this function in its own goroutine
func (l *Location) Move(newLoc *Location, speed float64) {

	// Check if bot is already moving
	if l.Moving.IsMoving {
		// If Bot is moving block until the channel is read from
		// meaning the bot has stopped moving
		l.Moving.Stop <- true
		l.client.Emit(&MovingDirectionChangedEvent{newLoc})
	}

	// Set IsMoving to true so the bot knows it is currently moving in the world
	l.Moving.IsMoving = true

	// Find out in a straight line, roughly how far of a distance
	// the target location is from point `A` to point `B` in
	// meters
	R := 6371000.0 // Meters
	lat1 := l.Latitude * math.Pi / 180
	lat2 := l.Longitude * math.Pi / 180
	diffLatRad := (newLoc.Latitude - l.Latitude) * math.Pi / 180
	diffLonRad := (newLoc.Longitude - l.Longitude) * math.Pi / 180

	a := math.Sin(diffLatRad/2)*math.Sin(diffLatRad/2) +
		math.Cos(lat1)*math.Cos(lat2)*math.Sin(diffLonRad/2)*math.Sin(diffLonRad/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distanceTotal := R * c // Distance to travel in a straight line, in meters (meters / 1000 to get Kilometers)
	distanceMoved := 0.0

	l.Moving.Distance = distanceTotal

	// Using the found distance of `distanceToMove` calculate and move
	// to the next point with the set speed in `meters per second`
	ticker := time.Tick(time.Second * 1)
	for stop := false; !stop; {
		select {
		case <-l.Moving.Stop:
			// Bot was manually told to stop wherever it is at this time
			// bot will stand here until it receives another move event
			stop = true
			l.Moving.IsMoving = false
			l.client.Emit(&MovingDoneEvent{})
			break
		case <-ticker:
			deltaLat := (newLoc.Latitude - l.Latitude) * (distanceMoved / distanceTotal)
			deltaLng := (newLoc.Longitude - l.Longitude) * (distanceMoved / distanceTotal)
			newLat := l.GetLatitudeF() + deltaLat
			newLng := l.GetLongitudeF() + deltaLng
			l.SetLatitude(Locnum(newLat))
			l.SetLongitude(Locnum(newLng))
			l.client.Emit(&MovingUpdateEvent{Location: &Location{
				Latitude:  newLat,
				Longitude: newLng,
				Moving:    l.Moving,
			},
				DistanceTravelled: distanceMoved,
				DistanceTotal:     distanceTotal,
			})
			if distanceMoved >= distanceTotal {
				// Bot may be traveling too fast to get an accurate landing and
				// might overshoot the location, once overshot
				// default to the destination.
				l.SetLatitude(Locnum(newLoc.Latitude))
				l.SetLongitude(Locnum(newLoc.Longitude))
				l.client.Emit(&MovingDoneEvent{l})
				stop = true
			}
			distanceMoved = distanceMoved + speed
			l.Moving.DistanceTravelled = distanceMoved
		}
	}

}

// Make the bot stop moving and sit in place
func (m *Moving) Sit(client *Client) {
	if m.IsMoving {
		m.Stop <- true
	}

	client.Emit(&MovingDoneEvent{client.Location})
}
