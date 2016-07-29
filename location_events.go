package pgo

type LocationSet struct {
	Location *Location
}

// This Event is fired while the bot is in transit (using Location.Move)
type MovingUpdateEvent struct {
	Location          *Location
	DistanceTravelled float64
	DistanceTotal     float64
}

type MovingDirectionChangedEvent struct {
	Location *Location
}

// This Event is fired once the bot has reached its destination (using Location.Move)
type MovingDoneEvent struct {
	Location *Location
}
