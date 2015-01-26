package main

import (
	. "github.com/jondunning/gogoogledm"
)

func main() {
	apiKey := "your-google-api-key" //obtain your key from Google Developers Console
	languageCode := "en-GB"         //codes available here https://developers.google.com/maps/faq#languagesupport
	unitSystem := ImperialUnit      //ImperialUnit or MetricUnit
	api := gogoogledm.NewDistanceMatrixAPI(apiKey, languageCode, unitSystem)

	origins := []Coordinates{
		Coordinates{
			Latitude:  55.853551,
			Longitude: -4.311093,
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
	}

	transportMode := Driving //Driving, Walking and Bicycling
	resp, err := api.GetDistances(origins, destinations, transportMode)
}
