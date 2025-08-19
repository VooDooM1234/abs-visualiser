package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/VooDooM1234/abs-visualiser/config"
	"github.com/VooDooM1234/abs-visualiser/db"
	"github.com/VooDooM1234/abs-visualiser/fetch"
)

func main() {
	ctx := context.Background()
	config.Init()

	databaseConnect, err := db.NewDatabase(ctx)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}

	defer databaseConnect.Close()

	ABSClient := fetch.NewFetchData("https", "data.api.abs.gov.au", 0)
	ABSresponse, _ := ABSClient.ABSRestDataCSV(config.GetABSKey(), "ABS,CPI", "1.115486.10.50.Q", "")
	log.Print("INSERTING ABS CPI data")

	for _, res := range ABSresponse {
		val, _ := strconv.ParseFloat(res.ObsValue, 64)
		cpi := db.ABS_CPI{
			TIME_PERIOD: res.TimePeriodID,
			VALUE:       val, // You need to convert string to float64
		}
		err := databaseConnect.UpsertDataABSCPI(cpi)
		if err != nil {
			fmt.Println(err)
			return
		}

	}
	// ...existing code...

	// http.HandleFunc('/')

}
