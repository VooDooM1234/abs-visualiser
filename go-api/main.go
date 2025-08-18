package main

import (
	"fmt"

	"github.com/VooDooM1234/abs-visualiser/config"
	"github.com/VooDooM1234/abs-visualiser/fetch"
)

func main() {
	// config.Init()

	ABSClient := fetch.NewFetchData("https", "https://data.api.abs.gov.au/rest/data/", 0)
	ABSresponse, _ := ABSClient.ABSRestDataCSV(config.GetABSKey(), "/ABS%2CCPI/1.115486.10.50.Q?startPeriod=2023&endPeriod=2024&format=csvfilewithlabels&detail=dataonly")

	for _, res := range ABSresponse {
		fmt.Println("Period: ", res.TimePeriodID)
		fmt.Println("Value: ", res.ObsValue)

	}

	// http.HandleFunc('/')

}
