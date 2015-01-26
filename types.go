package gogoogledm

import (
	"fmt"
	"time"
)

type DistanceMatrixAPI struct {
	apiKey              string
	maxElementsPerQuery int
	timeToWait          time.Duration
	languageCode        string
	unitSystem          UnitSystem
}

type ApiResponse struct {
	DestinationAddresses []string `json:"destination_addresses"`
	OriginAddresses      []string `json:"origin_addresses"`
	Rows                 []struct {
		Elements []struct {
			Distance struct {
				Text  string
				Value float64
			}
			Duration struct {
				Text  string
				Value float64
			}
			Fare struct {
				Currency string
				Value    float64
			}
			Status string
		}
	}
	Status string
}

type ApiCall struct {
	Origins      []Coordinates
	Destinations []Coordinates
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func (coordinates Coordinates) String() string {
	return fmt.Sprintf("%v,%v", coordinates.Latitude, coordinates.Longitude)
}

type UnitSystem int

const (
	MetricUnit UnitSystem = 1 + iota
	ImperialUnit
)

var unitSystems = []string{
	"metric",
	"imperial",
}

func (unitSystem UnitSystem) String() string {
	return unitSystems[unitSystem-1]
}

type AccountType int

const (
	FreeAccount AccountType = 1 + iota
	GoogleForWorkAccount
)

var accountTypes = []string{
	"FreeAccount",
	"GoogleForWorkAccount",
}

func (accountType AccountType) String() string {
	return accountTypes[accountType-1]
}

type TransportMode int

const (
	Walking TransportMode = 1 + iota
	Bicycling
	Transit
	Driving
)

var transportModes = []string{
	"walking",
	"bicycling",
	"transit",
	"driving",
}

func (transportMode TransportMode) String() string {
	return transportModes[transportMode-1]
}
