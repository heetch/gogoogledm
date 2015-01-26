package gogoogledm

import (
	"log"
	"reflect"
	"testing"
)

func TestGetDistances(t *testing.T) {
	apiKey := ""
	api := NewDistanceMatrixAPI(apiKey, FreeAccount, "en-GB", ImperialUnit)

	origins := []Coordinates{
		Coordinates{
			Latitude:  55.853551,
			Longitude: -4.311093,
		},
		Coordinates{
			Latitude:  53.608092,
			Longitude: -2.1469184,
		},
	}

	destinations := []Coordinates{
		Coordinates{
			Latitude:  53.4720286,
			Longitude: -2.3308237,
		},
		Coordinates{
			Latitude:  51.556021,
			Longitude: -0.279519,
		},
		Coordinates{
			Latitude:  51.556023,
			Longitude: -0.279522,
		},
	}

	resp, err := api.GetDistances(origins, destinations, Driving)
	if err != nil {
		t.Error("Error getting distances")
	}

	if len(resp.Rows) != len(origins) {
		t.Error("Origin rows not the same as the count sent")
	}
	for _, v := range resp.Rows {
		if len(v.Elements) != len(destinations) {
			t.Error("Origin rows not the same as the count sent")
		}
	}
}

func TestGetDistancesWithOver100Elements(t *testing.T) {
	apiKey := ""
	api := NewDistanceMatrixAPI(apiKey, FreeAccount, "en-GB", ImperialUnit)

	origins := []Coordinates{
		Coordinates{
			Latitude:  55.85,
			Longitude: -4.31,
		},
		Coordinates{
			Latitude:  56.85,
			Longitude: -5.31,
		},
	}

	var destinations []Coordinates
	for i := 0; i < 103; i++ {
		destination := Coordinates{
			Latitude:  53.47,
			Longitude: -2.33,
		}
		destinations = append(destinations, destination)
	}

	_, err := api.GetDistances(origins, destinations, Driving)
	if err != nil {
		t.Error(err.Error())
	}

	// if len(resp.Rows) != len(origins) {
	// 	t.Error("Row count does not match origin count")
	// }
	// for _, r := range resp.Rows {
	// 	if len(r.Elements) != len(destinations) {
	// 		t.Error("Element count does not match destination count")
	// 	}
	// }
}

func TestCoordinatesSliceToString(t *testing.T) {
	coordinates := []Coordinates{
		Coordinates{
			Latitude:  53.4720286,
			Longitude: -2.3308237,
		},
		Coordinates{
			Latitude:  51.556021,
			Longitude: -0.279519,
		},
	}

	result := coordinatesSliceToString(coordinates)
	if result != "53.4720286,-2.3308237|51.556021,-0.279519" {
		t.Error("Coordinates didnt match expected string")
	}
}

func TestNumberOfApiCallsRequired(t *testing.T) {
	apiKey := ""
	api := NewDistanceMatrixAPI(apiKey, FreeAccount, "en-GB", ImperialUnit)

	origins := []Coordinates{
		Coordinates{
			Latitude:  55.853551,
			Longitude: -4.311093,
		},
	}

	var destinations []Coordinates
	for i := 0; i < 101; i++ {
		destination := Coordinates{
			Latitude:  53.47,
			Longitude: -2.33,
		}
		destinations = append(destinations, destination)
	}

	count := api.numberOfApiCallsRequired(origins, destinations, Driving)
	log.Println(count)
	if count != 2 {
		t.Error("Number of API requests does not equal expected value")
	}

	origins = []Coordinates{
		Coordinates{
			Latitude:  55.853551,
			Longitude: -4.311093,
		},
	}

	destinations = nil
	for i := 0; i < 100; i++ {
		destination := Coordinates{
			Latitude:  53.47,
			Longitude: -2.33,
		}
		destinations = append(destinations, destination)
	}

	count = api.numberOfApiCallsRequired(origins, destinations, Driving)
	log.Println(count)
	if count != 2 {
		t.Error("Number of API requests does not equal expected value")
	}
}

func TestSplitSliceIntoBlocks(t *testing.T) {
	var coordinates []Coordinates
	for i := 0; i < 3; i++ {
		c := Coordinates{
			Latitude:  float64(i),
			Longitude: float64(i + 1),
		}
		coordinates = append(coordinates, c)
	}

	blocks := splitSliceIntoBlocks(coordinates, 1)

	a1 := coordinates[:1]
	a2 := coordinates[1:2]
	a3 := coordinates[2:3]

	if !reflect.DeepEqual(blocks[0], a1) {
		t.Error("Block is not as expected")
	}
	if !reflect.DeepEqual(blocks[1], a2) {
		t.Error("Block is not as expected")
	}
	if !reflect.DeepEqual(blocks[2], a3) {
		t.Error("Block is not as expected")
	}

	coordinates = nil
	for i := 0; i < 7; i++ {
		c := Coordinates{
			Latitude:  float64(i),
			Longitude: float64(i + 1),
		}
		coordinates = append(coordinates, c)
	}

	blocks = splitSliceIntoBlocks(coordinates, 2)

	b1 := coordinates[:2]
	b2 := coordinates[2:4]
	b3 := coordinates[4:6]
	b4 := coordinates[6:7]

	if !reflect.DeepEqual(blocks[0], b1) {
		t.Error("Block is not as expected")
	}
	if !reflect.DeepEqual(blocks[1], b2) {
		t.Error("Block is not as expected")
	}
	if !reflect.DeepEqual(blocks[2], b3) {
		t.Error("Block is not as expected")
	}
	if !reflect.DeepEqual(blocks[3], b4) {
		t.Error("Block is not as expected")
	}
}

func TestDDD(t *testing.T) {
	// Pick the largest
	// Divide by the number of calls required and and add remainder to last api call
	//
	//
	// 9 origins, 13 destinations = 117 elements (100 max per call) = 2 api calls
	// 1 origins, 101 destinations = 101 elements (100 max per call) = 2 api calls
	// 101 origins, 1 destinations = 101 elements (100 max per call) = 2 api calls
	// 11 origins, 11 destinations, 121 elements (100 max per call) = 2 api calls
	// 1 origin, 2 destinations, 2 elements (1 max per call) = 2 api calls
	// 3 origin, 1 destinations, 3 elements (1 max per call) = 3 api calls

	// origins := float64(9)
	// destinations := float64(13)
	// //elements := origins * destinations //117
	// maxElementsPerApiCall := float64(100)
	// //apiCallsRequired := 2

	// //Work out the max calls that will fit into the first API call
	// maxOrigins := math.Floor(float64(maxElementsPerApiCall / origins))
	// log.Println(maxOrigins)
	// maxDistinations := math.Floor(float64(maxElementsPerApiCall / destinations))
	// log.Println(maxDistinations)

	// log.Println(maxOrigins * origins)
	// log.Println(maxDistinations * destinations)
}
