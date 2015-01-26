package gogoogledm

import (
	"encoding/json"
	"errors"
	"fmt"
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
		api.maxElementsPerRequest = 100
	case GoogleForWorkAccount:
		api.maxElementsPerRequest = 625
	default:
		panic("Unknown accountType")
	}

	return &api
}

func (api *DistanceMatrixAPI) buildBaseUrlParams() url.Values {
	params := url.Values{}
	params.Add("key", api.apiKey)
	params.Add("language", api.languageCode)
	params.Add("units", api.unitSystem.String())

	return params
}

func (api *DistanceMatrixAPI) GetDistances(origins []Coordinates, destinations []Coordinates, transportMode TransportMode) (*ApiResponse, error) {
	apiRequestCount := api.numberOfApiCallsRequired(origins, destinations, transportMode)
	groupedCoordinates := api.groupCoordinates(origins, destinations, apiRequestCount)

	var joinedResponse ApiResponse
	remaining := api.maxElementsPerRequest
	for _, group := range groupedCoordinates {
		need := (len(group.Origins) * len(group.Destinations))
		if remaining < need {
			time.Sleep(api.timeToWait)
			remaining = api.maxElementsPerRequest
		}

		resp, err := api.sendRequest(group.Origins, group.Destinations, transportMode)
		if err != nil {
			return nil, err
		}

		joinedResponse.Status = resp.Status
		joinedResponse.OriginAddresses = append(joinedResponse.OriginAddresses, resp.OriginAddresses...)
		joinedResponse.DestinationAddresses = append(joinedResponse.DestinationAddresses, resp.DestinationAddresses...)
		joinedResponse.Rows = append(joinedResponse.Rows, resp.Rows...)

		remaining -= need
	}

	return &joinedResponse, nil
}

func (api *DistanceMatrixAPI) sendRequest(origins []Coordinates, destinations []Coordinates, transportMode TransportMode) (*ApiResponse, error) {
	urlValues := api.buildBaseUrlParams()
	urlValues.Add("mode", transportMode.String())
	urlValues.Add("origins", coordinatesSliceToString(origins))
	urlValues.Add("destinations", coordinatesSliceToString(destinations))
	url := base_url + urlValues.Encode()

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

func (api *DistanceMatrixAPI) groupCoordinates(origins []Coordinates, destinations []Coordinates, maxGroupSize int) (apiCalls []ApiCall) {
	if maxGroupSize == 1 {
		apiCalls = append(apiCalls, ApiCall{origins, destinations})
		return apiCalls
	}

	destinationsSize := len(destinations)
	originsSize := len(origins)

	if destinationsSize > originsSize {
		//Split destinations
		maxBlockSize := math.Floor(float64(destinationsSize) / float64(maxGroupSize))
		blocks := splitSliceIntoBlocks(destinations, int(maxBlockSize))

		for _, b := range blocks {
			apiCalls = append(apiCalls, ApiCall{
				Origins:      origins,
				Destinations: b,
			})
		}
	} else {
		//Split origins
		maxBlockSize := math.Floor(float64(originsSize) / float64(maxGroupSize))
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

func (api *DistanceMatrixAPI) numberOfApiCallsRequired(origins []Coordinates, destinations []Coordinates, transportMode TransportMode) int {
	urlValues := api.buildBaseUrlParams()
	urlValues.Add("mode", transportMode.String())

	//Number of calls required by origin/destination combination
	elementCount := float64(len(origins) * len(destinations))
	apiCallsRequired := math.Ceil(elementCount / float64(api.maxElementsPerRequest))

	//Number of calls required due to url length limitation
	urlValues.Add("origins", coordinatesSliceToString(origins))
	urlValues.Add("destinations", coordinatesSliceToString(destinations))
	url := base_url + urlValues.Encode()
	urlLength := len(url)
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
