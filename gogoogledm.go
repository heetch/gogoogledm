package gogoogledm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const (
	base_url = "https://maps.googleapis.com/maps/api/distancematrix/json?"
)

// Users of the free API:
// 100 elements per query.
// 100 elements per 10 seconds.
// 2,500 elements per 24 hour period.

// Google Maps API for Work customers:
// 625 elements per query.
// 1,000 elements per 10 seconds.
// 100,000 elements per 24 hour period.

// Distance Matrix API URLs are restricted to approximately 2000 characters, after URL Encoding.

// Top-level Status Codes
// ----------------------
// OK indicates the response contains a valid result.
// INVALID_REQUEST indicates that the provided request was invalid.
// MAX_ELEMENTS_EXCEEDED indicates that the product of origins and destinations exceeds the per-query limit.
// OVER_QUERY_LIMIT indicates the service has received too many requests from your application within the allowed time period.
// REQUEST_DENIED indicates that the service denied use of the Distance Matrix service by your application.
// UNKNOWN_ERROR indicates a Distance Matrix request could not be processed due to a server error. The request may succeed if you try again.

// Element-level Status Codes
// ----------------------
// OK indicates the response contains a valid result.
// NOT_FOUND indicates that the origin and/or destination of this pairing could not be geocoded.
// ZERO_RESULTS indicates no route could be found between the origin and destination.

type Trip struct {
	Origins      []Coordinates
	Destinations []Coordinates
	Mode         string
}

type Coordinates struct {
	Latitude  float64
	Longitude float64
}

func (coordinates Coordinates) String() string {
	return fmt.Sprintf("%v,%v", coordinates.Latitude, coordinates.Longitude)
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

type DistanceMatrixAPI struct {
	apiKey       string
	languageCode string
	unitSystem   UnitSystem
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

func NewDistanceMatrixAPI(apiKey string, languageCode string, unitSystem UnitSystem) *DistanceMatrixAPI {
	api := DistanceMatrixAPI{
		apiKey:       apiKey,
		languageCode: languageCode,
		unitSystem:   unitSystem,
	}

	return &api
}

func (api *DistanceMatrixAPI) GetDistances(origins []Coordinates, destinations []Coordinates, transportMode TransportMode) (*ApiResponse, error) {
	q := url.Values{}
	q.Add("key", api.apiKey)
	q.Add("mode", transportMode.String())
	q.Add("language", api.languageCode)
	q.Add("units", api.unitSystem.String())
	q.Add("origins", convertCoordinateSliceToString(origins))
	q.Add("destinations", convertCoordinateSliceToString(destinations))

	url := base_url + q.Encode()
	log.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var apiResponse ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, err
	}

	if err = validateResponse(origins, destinations, apiResponse); err != nil {
		return nil, err
	}

	return &apiResponse, nil
}

func validateResponse(origins []Coordinates, destinations []Coordinates, apiResponse ApiResponse) error {
	if apiResponse.Status != "OK" {
		errors.New(fmt.Sprintf("API returned error: %s", apiResponse.Status))
	}
	if len(apiResponse.Rows) != len(origins) {
		return errors.New("API returned less rows than origins requested")
	}
	for _, r := range apiResponse.Rows {
		if len(r.Elements) != len(destinations) {
			return errors.New("API returned less elements than destinations requested")
		}
		for ei, e := range r.Elements {
			if e.Status != "OK" {
				errors.New(fmt.Sprintf("API returned error in element(%v): %s", ei, apiResponse.Status))
			}
		}
	}

	return nil
}

func convertCoordinateSliceToString(coordinates []Coordinates) (result string) {
	seperator := "|"
	for _, c := range coordinates {
		result += c.String() + seperator
	}
	result = strings.TrimSuffix(result, seperator)

	return result
}

// type directions_response struct {
// 	Routes []struct {
// 		Legs []struct {
// 			Distance struct {
// 				Value uint64
// 			}
// 		}
// 	}
// }
