package main

import (
	"encoding/json"
	"io/ioutil"
	"math"
)

// Point is a geo co-ordinate
type Point struct {
	lat float64
	lng float64
}

// BusStop describes a Singaporean (LTA) bus stop
type BusStop struct {
	BusStopCode string  `json:"BusStopCode"`
	RoadName    string  `json:"RoadName"`
	Description string  `json:"Description"`
	Latitude    float64 `json:"Latitude"`
	Longitude   float64 `json:"Longitude"`
}

// BusStops are many bus stops
type BusStops []BusStop

func loadBusJSON(jsonfile string) (bs BusStops, err error) {
	content, err := ioutil.ReadFile(jsonfile)
	if err != nil {
		return
	}
	err = json.Unmarshal(content, &bs)
	if err != nil {
		return
	}

	return
}

func (BusStops BusStops) closest(location Point) BusStop {
	c := -1
	// fmt.Println(c)
	closestSoFar := math.Inf(1)

	// log.Println(c.Description, closestSoFar)
	for i := range bs {
		distance := location.distance(Point{bs[i].Latitude, bs[i].Longitude})
		// log.Printf("'%s' %.1f\n", p.Description, distance)
		if distance < closestSoFar {
			// Set the return
			c = i
			// Record closest distance
			closestSoFar = distance
		}
	}

	return bs[c]

}

func (BusStops BusStops) nameBusStopID(busid string) (description string) {
	for _, p := range BusStops {
		if busid == p.BusStopCode {
			return p.Description
		}
	}
	return ""
}

// distance calculates the distance between two points
func (p Point) distance(p2 Point) float64 {
	latd := p2.lat - p.lat
	lngd := p2.lng - p.lng
	return latd*latd + lngd*lngd
}
