# Google Distance Matrix API Library for Go

This library allows you obtain the distance and travel times between multiple origins and destinations via various travel methods. 
The library handles Google API rate limiting and max encoded url length for you so you can just implement and enjoy :D

Limitations found here https://developers.google.com/maps/documentation/distancematrix/#Limits

Full details of the Google Matrix API is here https://developers.google.com/maps/documentation/distancematrix

## Installation

    go get github.com/jondunning/gogoogledm

## Usage

    package main

    import (
        "fmt"
        . "github.com/jondunning/gogoogledm"
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
        resp, err := api.GetDistances(origins, destinations, transportMode)
        if err != nil {
            panic(err)
        }

        for _, r := range resp.Rows {
            for _, e := range r.Elements {
                fmt.Printf("Status=%s, Distance=%s, Duration=%s", e.Status, e.Distance.Text, e.Duration.Text)
            }
        }
    }

## Limitations

1. The library only implements origins and destinations in a coordinate format
2. Currently only supports Driving, Walking and Bicycling travel modes

## Contributing

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D 

## Testing

Various tests included, just run;

    go test

## License

Copyright (c) 2015 Jon Dunning. See the LICENSE file for license rights and limitations (MIT).
