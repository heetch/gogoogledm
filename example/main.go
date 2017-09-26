package main

import (
	"fmt"
	"time"

	"golang.org/x/net/context"

	. "github.com/heetch/gogoogledm"
)

func main() {
	apiKey := ""               //obtain your key from Google Developers Console
	accountType := FreeAccount //FreeAccount or GoogleForWorkAccount
	languageCode := "en-GB"    //codes available here https://developers.google.com/maps/faq#languagesupport
	unitSystem := ImperialUnit //ImperialUnit or MetricUnit
	api := NewDistanceMatrixAPI(apiKey, accountType, languageCode, unitSystem)

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
	ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
	defer cancel()
	resp, err := api.GetDistances(ctx, origins, destinations, transportMode)
	if err != nil {
		if ctx.Err() != nil {
			// Request has timeouts
		}
		panic(err)
	}

	for _, r := range resp.Rows {
		for _, e := range r.Elements {
			fmt.Printf("Status=%s, Distance=%s, Duration=%s", e.Status, e.Distance.Text, e.Duration.Text)
		}
	}
}
