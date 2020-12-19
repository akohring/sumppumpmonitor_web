package main

import "time"

// Pit is the sump pump well
type Pit struct {
	PitID       int
	Healthy     bool
	LastUpdated time.Time
	PitLevels   []PitLevel
}

// PitLevel is the water level of a pit at a given point of time
type PitLevel struct {
	PitID       int
	DateCreated time.Time
	Level       float64
}
