package main

import (
	"fmt"
	"os"

	"github.com/VooDooM1234/abs-visualiser/config"
	"github.com/VooDooM1234/abs-visualiser/fetch"
	"github.com/gocarina/gocsv"
)

func main() {
	// config.Init()

	ABSClient := fetch.NewFetchData("https", "data.api.abs.gov.au", 0)
	ABSresponse, _ := ABSClient.ABSRestDataCSV(config.GetABSKey(), "ABS,CPI", "1.115486.10.50.Q", "")

	for _, res := range ABSresponse {
		fmt.Println("Period: ", res.TimePeriodID)
		fmt.Println("Value: ", res.ObsValue)

	}

	file, err := os.Create("../.testdata/CPI.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	gocsv.MarshalFile(&ABSresponse, file)

	// http.HandleFunc('/')

}
