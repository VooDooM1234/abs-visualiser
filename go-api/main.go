package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/VooDooM1234/abs-visualiser/config"
	"github.com/VooDooM1234/abs-visualiser/db"
	"github.com/VooDooM1234/abs-visualiser/handlers"
)

func launchPythonMicroservice(config *config.Config) {
	cmd := exec.Command(
		config.PythonPath,
		"-m", "uvicorn",
		"python_ds.main:app",
		"--host", config.Host,
		"--port", config.PlotServicePort,
		"--app-dir", "..",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to launch Python microservice: %v", err)
	}

	log.Printf("Python microservice running at http://%s:%s\n", config.Host, config.PlotServicePort)
}

func main() {
	logger := log.Default()

	ctx := context.Background()
	config, err := config.Init()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return
	}

	databaseConnect, err := db.NewDatabase(ctx)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return
	}

	defer databaseConnect.Close()

	// ABSClient := fetch.NewFetchData("https", "data.api.abs.gov.au", 0)
	// ABSresponse, _ := ABSClient.ABSRestDataCSV(config.GetABSKey(), "ABS,CPI", "1.115486.10.50.Q", "")
	// log.Print("INSERTING ABS CPI data")

	// for _, res := range ABSresponse {
	// 	fmt.Println(res.TimePeriodID)
	// 	fmt.Println(res.ObsValue)
	// val, _ := strconv.ParseFloat(res.ObsValue, 64)
	// cpi := db.ABS_CPI{
	// 	TIME_PERIOD: res.TimePeriodID,
	// 	VALUE:       val, // You need to convert string to float64
	// }
	// err := databaseConnect.UpsertDataABSCPI(cpi)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// }

	http.Handle("/", handlers.HomePageHandler(config, logger))
	http.Handle("/plot/bar-abs-cpi/", handlers.PlotABSCPIHandler(config, logger))
	http.Handle("/plot/test/", handlers.PlotTestHandler(config, logger))
	http.Handle("/plot/test/json/", handlers.PlotTestJSONHandler(config, logger))

	log.Println("Starting server on 0.0.0.0:8080")
	launchPythonMicroservice(config)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}
