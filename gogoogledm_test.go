package gogoogledm

import (
	"log"
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
	}

	var destinations []Coordinates
	for i := 0; i < 101; i++ {
		destination := Coordinates{
			Latitude:  53.47,
			Longitude: -2.33,
		}
		destinations = append(destinations, destination)
	}
	log.Println("...")
	log.Println(len(destinations))

	resp, err := api.GetDistances(origins, destinations, Driving)
	log.Println(resp)
	log.Println(err)
	if err != nil {

		t.Error("Error getting distances")
	}
}

func TestConvertCoordinateSliceToString(t *testing.T) {
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

	result := convertCoordinateSliceToString(coordinates)
	if result != "53.4720286,-2.3308237|51.556021,-0.279519" {
		t.Error("Coordinates didnt match expected string")
	}

}
