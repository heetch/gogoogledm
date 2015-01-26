package gogoogledm

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	base_url     = "https://maps.googleapis.com/maps/api/distancematrix/json?"
	maxUrlLength = 2000
)

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

func NewDistanceMatrixAPI(apiKey string, accountType AccountType, languageCode string, unitSystem UnitSystem) *DistanceMatrixAPI {
	api := DistanceMatrixAPI{
		apiKey:       apiKey,
		languageCode: languageCode,
		unitSystem:   unitSystem,
		timeToWait:   10 * time.Second,
	}

	// Users of the free API:
	// 100 elements per query.
	// 100 elements per 10 seconds.
	// 2,500 elements per 24 hour period.

	// Google Maps API for Work customers:
	// 625 elements per query.
	// 1,000 elements per 10 seconds.
	// 100,000 elements per 24 hour period.
	switch accountType {
	case FreeAccount:
		api.maxElementsPerQuery = 100
	case GoogleForWorkAccount:
		api.maxElementsPerQuery = 625
	default:
		panic("Unknown accountType")
	}

	return &api
}

func (api *DistanceMatrixAPI) GetDistances(origins []Coordinates, destinations []Coordinates, transportMode TransportMode) (*[]ApiResponse, error) {
	baseUrlValues := url.Values{}
	baseUrlValues.Add("key", api.apiKey)
	baseUrlValues.Add("language", api.languageCode)
	baseUrlValues.Add("units", api.unitSystem.String())
	baseUrlValues.Add("mode", transportMode.String())

	apiRequestCount := numberOfApiCallsRequired(origins, destinations, api.maxElementsPerQuery, baseUrlValues)
	apiCalls := api.apiCallSplitter(origins, destinations, apiRequestCount)

	var apiResponses []ApiResponse
	for _, apiCall := range apiCalls {
		uv := baseUrlValues
		uv.Add("origins", coordinatesSliceToString(apiCall.Origins))
		uv.Add("destinations", coordinatesSliceToString(apiCall.Destinations))
		url := base_url + baseUrlValues.Encode()

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

		apiResponses = append(apiResponses, apiResponse)
	}

	return &apiResponses, nil
}

func (api *DistanceMatrixAPI) apiCallSplitter(origins []Coordinates, destinations []Coordinates, apiRequestCount int) (apiCalls []ApiCall) {
	if apiRequestCount == 1 {
		apiCalls = append(apiCalls, ApiCall{origins, destinations})
		return apiCalls
	}

	destinationsSize := len(destinations)
	originsSize := len(origins)

	if destinationsSize > originsSize {
		//Split destinations
		maxBlockSize := math.Floor(float64(destinationsSize) / float64(apiRequestCount))
		blocks := splitSliceIntoBlocks(destinations, int(maxBlockSize))

		for _, b := range blocks {
			apiCalls = append(apiCalls, ApiCall{
				Origins:      origins,
				Destinations: b,
			})
		}
	} else {
		//Split origins
		maxBlockSize := math.Floor(float64(originsSize) / float64(apiRequestCount))
		blocks := splitSliceIntoBlocks(origins, int(maxBlockSize))

		for _, o := range blocks {
			apiCalls = append(apiCalls, ApiCall{
				Origins:      o,
				Destinations: destinations,
			})
		}
	}

	return apiCalls
}

func numberOfApiCallsRequired(origins []Coordinates, destinations []Coordinates, maxElementsPerCall int, baseUrlValues url.Values) int {
	//Number of calls required by origin/destination combination
	elementCount := float64(len(origins) * len(destinations))
	apiCallsRequired := math.Ceil(elementCount / float64(maxElementsPerCall))
	log.Printf("apiCallsRequired: %v", apiCallsRequired)

	//Number of calls required due to url length limitation
	baseUrlValues.Add("origins", coordinatesSliceToString(origins))
	baseUrlValues.Add("destinations", coordinatesSliceToString(destinations))
	url := base_url + baseUrlValues.Encode()
	urlLength := len(url)
	log.Printf("urlLength: %v", urlLength)

	apiCallsRequiredByUrl := math.Ceil(float64(urlLength) / float64(maxUrlLength))

	return int(math.Max(apiCallsRequired, apiCallsRequiredByUrl))
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
	}

	return nil
}

func coordinatesSliceToString(coordinates []Coordinates) (result string) {
	seperator := "|"
	for _, c := range coordinates {
		result += c.String() + seperator
	}
	result = strings.TrimSuffix(result, seperator)

	return result
}

func splitSliceIntoBlocks(slice []Coordinates, maxBlockSize int) [][]Coordinates {
	sliceSize := len(slice)
	numberOfBlocks := int(math.Ceil(float64(sliceSize) / float64(maxBlockSize)))
	blocks := make([][]Coordinates, numberOfBlocks)

	i := 0
	for remaining := sliceSize; remaining > 0; remaining -= maxBlockSize {
		start := i * maxBlockSize
		if remaining < maxBlockSize {
			maxBlockSize = remaining
		}
		blocks[i] = make([]Coordinates, maxBlockSize)
		blocks[i] = slice[start : start+maxBlockSize]
		i++
	}

	return blocks
}
