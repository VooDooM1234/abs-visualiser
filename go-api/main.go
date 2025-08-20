package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os/exec"

	"github.com/VooDooM1234/abs-visualiser/config"
	"github.com/VooDooM1234/abs-visualiser/db"
	"github.com/VooDooM1234/abs-visualiser/handlers"
)

func homePageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	tmpl := template.Must(template.ParseFiles("../templates/index.html"))

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	log.Print("Home Page")
}

func startPythonMicroservice(config *config.Config) {
	cmd := exec.Command(config.PythonPath, "-m", "uvicorn", config.PlotServiceScript, "--host", config.Host, "--port", config.PlotServicePort)
	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to start Python microservice: %v", err)
	}
	log.Printf("Python microservice started on %s:%s\n", config.Host, config.PlotServicePort)
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

	http.HandleFunc("/", homePageHandler)
	// http.Handle("/plot/bar-abs-cpi", handlers.NewPlotABSCPIHandler(logger))
	http.Handle("/plot/bar-abs-cpi/", handlers.NewPlotABSCPIHandler())
	http.Handle("/plot/test/", handlers.PlotTestHandler(logger))
	http.Handle("/plot/test/json/", handlers.PlotTestJSONHandler(logger))

	log.Println("Starting server on 0.0.0.0:8080")
	startPythonMicroservice(config)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))

}
