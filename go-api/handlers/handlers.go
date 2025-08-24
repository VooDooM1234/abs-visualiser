package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/VooDooM1234/abs-visualiser/go-api/config"
	"github.com/VooDooM1234/abs-visualiser/go-api/db"
)

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#maker-funcs-return-the-handler
// func handleSomething(config *config.Config, logger *Logger) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         // handler logic
//     })
// }

func validateGraphName(name string) error {
	var allowedGraphs = map[string]bool{
		"line":    true,
		"bar":     true,
		"pie":     true,
		"scatter": true,
	}
	if _, ok := allowedGraphs[name]; !ok {
		return fmt.Errorf("invalid graph name: %s", name)
	}
	return nil
}

// Fix this to not run query every time and finish caching function
func validateDataflowName(id string, db db.Database) error {
	query := fmt.Sprintf("SELECT * FROM abs_static_dataflow WHERE id = '%s'", id)

	allowedDataflows, err := db.GetABSDataflow(query)
	if err != nil {
		return fmt.Errorf("error fetching dataflow names: %w", err)
	}

	for _, dataflow := range allowedDataflows {
		if dataflow.ID == strings.ToUpper(id) {
			return nil
		}
	}
	return fmt.Errorf("invalid dataflow name: %s", id)
}

func HealthHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func SidebarHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		path := config.HTMLTemplates + "sidebar.html"
		tmpl := template.Must(template.ParseFiles(path))

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Print("Sidebar loaded")
	})
}

func IndexHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		path := config.HTMLTemplates + "index.html"
		tmpl := template.Must(template.ParseFiles(path))

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Print("Home Page")
	})
}

func HomeHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		path := config.HTMLTemplates + "home.html"
		tmpl := template.Must(template.ParseFiles(path))

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Print("Home Page")
	})
}

// Dashboard Page Handler
func DashboardHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	path := cfg.HTMLTemplates + "dashboard.html"
	tmpl := template.Must(template.ParseFiles(path))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Printf("Error executing dashboard template: %v", err)
		}
		log.Println("Dashboard Page")
	})
}

func ABSDataflowHandler(config *config.Config, logger *log.Logger, db *db.Database) http.Handler {
	path := config.HTMLTemplates + "abs_dataflow.html"
	tmpl := template.Must(template.ParseFiles(path))
	query := "SELECT id, version, agency_id, is_external_reference, is_final, name FROM abs_static_dataflow"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := db.GetABSDataflow(query)
		if err != nil {
			logger.Printf("Error fetching ABS dataflow: %v", err)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Printf("Error executing template: %v", err)
			return
		}
		logger.Print("ABS Dataflow Page")
	})
}

// Get request to python data science to request ABS data
// for raw data expertimentation - will be made redundent by a direct call for a dashboard request.
func RequestABSData(config *config.Config, logger *log.Logger) http.Handler {
	type ABSresponse struct {
		MEASURE     string  `json:"MEASURE"`
		INDEX       string  `json:"INDEX"`
		TSEST       string  `json:"TSEST"`
		REGION      string  `json:"REGION"`
		FREQ        string  `json:"FREQ"`
		TIME_PERIOD string  `json:"TIME_PERIOD"`
		Value       float64 `json:"VALUE"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		dataflowid := r.URL.Query().Get("dataflowid")
		if dataflowid == "" {
			http.Error(w, "Missing dataflowid parameter", http.StatusBadRequest)
			return
		}

		var payload = map[string]string{
			"dataflowid": dataflowid,
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			logger.Printf("Failed to marshal payload: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		logger.Printf("Retrieving data for dataflow: %s...", dataflowid)
		microserviceurl := fmt.Sprintf("http://%s:%s/request-data/ABS/", config.PlotServiceHost, config.PlotServicePort)
		logger.Printf("POST request to: %s", microserviceurl)
		req, err := http.NewRequest("POST", microserviceurl, bytes.NewBuffer(jsonPayload))
		if err != nil {
			logger.Printf("Failed to create POST request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 120 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("Python service unavailable: %v", err)
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("Error reading response body: %v", err)
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}

		var result []ABSresponse
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			http.Error(w, "Invalid response format", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.Printf("Failed to write response: %v", err)
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
		logger.Printf("Successfully retrieved data for dataflow: %s", dataflowid)
	})
}

// Plothandler endpoint /plot/{graphName}/{dataflow}
// change to use querty param nor endpoint
func PlotHandler(config *config.Config, logger *log.Logger, db *db.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		urlParts := strings.Split(r.URL.Path, "/")
		pathMap := make(map[string]string)
		if len(urlParts) > 1 {
			pathMap["base"] = urlParts[1]
		}
		if len(urlParts) > 2 {
			pathMap["graphName"] = urlParts[2]
		}
		if len(urlParts) > 3 {
			pathMap["dataflow"] = urlParts[3]
		}

		if err := validateGraphName(pathMap["graphName"]); err != nil {
			http.Error(w, fmt.Sprintf("Invalid graph name: %s", pathMap["graphName"]), http.StatusBadRequest)
			logger.Printf("Invalid graph name: %s", pathMap["graphName"])
			return
		}
		dataflow := strings.ToUpper(pathMap["dataflow"])
		if err := validateDataflowName(dataflow, *db); err != nil {
			http.Error(w, fmt.Sprintf("Invalid dataflow name: %s", dataflow), http.StatusBadRequest)
			logger.Printf("Invalid dataflow name: %s", dataflow)
			logger.Printf("Dataflow validation failed: %v", err)
			return
		}

		url := fmt.Sprintf("http://%s:%s/plot/%s/%s", config.Host, config.PlotServicePort, pathMap["graphName"], pathMap["dataflow"])
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		htmlBytes, _ := io.ReadAll(resp.Body)
		htmlString := string(htmlBytes)
		fmt.Fprint(w, htmlString)
	})
}

// FIX THIS FUNC AND ABSTRACT OUT ITEMS
// Retrieves POST request for dashboard data button then calls python plot microservice
// to generate the dashboard for the dataflow (ABS name for data ids)
// Endpoint: /dashboard-retrieve/
// Python microservice returns HTML for the dashboard
func RetrieveDashboardHandler(config *config.Config, logger *log.Logger) http.Handler {

	type Observation struct {
		TimePeriod string  `json:"period"`
		Value      float64 `json:"value"`
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			logger.Printf("Failed to parse form: %v", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		dataflowId := r.PostFormValue("dataflowId")
		datakey := "......"

		absurl := fmt.Sprintf(
			"https://data.api.abs.gov.au/rest/data/%s/%s?detail=dataonly",
			dataflowId, datakey,
		)
		logger.Printf("Fetching ABS data: %s", absurl)

		client := &http.Client{Timeout: 10 * time.Second}

		req, err := http.NewRequest("GET", absurl, nil)
		if err != nil {
			logger.Printf("Failed to create request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		req.Header.Set("Accept", "application/vnd.sdmx.data+json")

		resp, err := client.Do(req)
		if err != nil {
			logger.Printf("Failed to call ABS API: %v", err)
			http.Error(w, "ABS service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("Failed to read ABS response: %v", err)
			http.Error(w, "Failed to read ABS response", http.StatusInternalServerError)
			return
		}

		if resp.StatusCode >= 400 {
			logger.Printf("ABS API returned status %d", resp.StatusCode)
			var apiError struct {
				Errors []struct {
					ID     string `json:"id"`
					Detail string `json:"detail"`
					Code   string `json:"code"`
					Source struct {
						Parameter string `json:"parameter"`
					} `json:"source"`
				} `json:"errors"`
			}
			if json.Valid(body) {
				_ = json.Unmarshal(body, &apiError)
				if len(apiError.Errors) > 0 {
					msg := "ABS API Errors:\n"
					for _, e := range apiError.Errors {
						msg += fmt.Sprintf("%s - %s (Code: %s, Parameter: %s)\n", e.ID, e.Detail, e.Code, e.Source.Parameter)
					}
					http.Error(w, msg, resp.StatusCode)
					return
				}
			} else {
				http.Error(w, fmt.Sprintf("ABS API returned status %d: %s", resp.StatusCode, string(body)), resp.StatusCode)
				return
			}
		}

		// Use interface{} to handle numbers or strings in observations
		var raw struct {
			Data struct {
				DataSets []struct {
					Series map[string]struct {
						Observations map[string][]interface{} `json:"observations"`
					} `json:"series"`
				} `json:"dataSets"`
			} `json:"data"`
		}

		if err := json.Unmarshal(body, &raw); err != nil {
			logger.Printf("Failed to parse ABS JSON: %v", err)
			http.Error(w, "Failed to parse ABS JSON", http.StatusInternalServerError)
			return
		}

		var results []Observation
		for _, dataset := range raw.Data.DataSets {
			for _, series := range dataset.Series {
				for period, values := range series.Observations {
					if len(values) == 0 {
						continue
					}

					var val float64
					switch v := values[0].(type) {
					case string:
						val, _ = strconv.ParseFloat(v, 64)
					case float64:
						val = v
					case json.Number:
						val, _ = v.Float64()
					default:
						continue
					}

					results = append(results, Observation{
						TimePeriod: period,
						Value:      val,
					})
				}
			}
		}

		// for i, obs := range results {
		// 	fmt.Printf("Observation %d: Period=%s, Value=%.2f\n", i+1, obs.Period, obs.Value)
		// }

		// Forward to microservice
		logger.Printf("Retrieving dashboard for dataflow: %s", dataflowId)
		microserviceurl := fmt.Sprintf("http://%s:%s/dashboard/api/%s/", config.Host, config.PlotServicePort, dataflowId)

		payload, err := json.Marshal(results)
		if err != nil {
			logger.Printf("Failed to marshal observations: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		req, err = http.NewRequest("POST", microserviceurl, bytes.NewBuffer(payload))
		if err != nil {
			logger.Printf("Failed to create POST request: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		req.Header.Set("Content-Type", "application/json")

		client = &http.Client{Timeout: 10 * time.Second}
		resp, err = client.Do(req)
		if err != nil {
			logger.Printf("Python service unavailable: %v", err)
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		htmlBytes, _ := io.ReadAll(resp.Body)
		fmt.Fprint(w, string(htmlBytes))
	})
}

// Python Microservice handler for dashboard plots
// Endpoint: /plot/dashboard/{dataflow}
func PlotDashboardHandler(config *config.Config, logger *log.Logger, db *db.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		urlParts := strings.Split(r.URL.Path, "/")
		pathMap := make(map[string]string)
		if len(urlParts) > 1 {
			pathMap["base"] = urlParts[1]
		}
		if len(urlParts) > 3 {
			pathMap["dataflow"] = urlParts[2]
		}

		dataflow := strings.ToUpper(pathMap["dataflow"])
		if err := validateDataflowName(dataflow, *db); err != nil {
			http.Error(w, fmt.Sprintf("Invalid dataflow name: %s", dataflow), http.StatusBadRequest)
			logger.Printf("Invalid dataflow name: %s", dataflow)
			logger.Printf("Dataflow validation failed: %v", err)
			return
		}

		url := fmt.Sprintf("http://%s:%s/plot/dashboard/%s", config.Host, config.PlotServicePort, pathMap["dataflow"])
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		defer resp.Body.Close()
		htmlBytes, _ := io.ReadAll(resp.Body)
		htmlString := string(htmlBytes)
		fmt.Fprint(w, htmlString)
	})
}

func PlotTestHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		url := fmt.Sprintf("http://%s:%s/plot/test", config.Host, config.PlotServicePort)
		resp, err := http.Get(url)

		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}

		defer resp.Body.Close()
		htmlBytes, _ := io.ReadAll(resp.Body)
		htmlString := string(htmlBytes)
		fmt.Fprint(w, htmlString)
	})
}

func PlotTestJSONHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		url := fmt.Sprintf("http://%s:%s/plot/test/json", config.Host, config.PlotServicePort)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			logger.Printf("Error fetching JSON from Python service: %v", err)
			return
		}
		defer resp.Body.Close()

		if _, err := io.Copy(w, resp.Body); err != nil {
			http.Error(w, "Failed to stream response", http.StatusInternalServerError)
			logger.Printf("Error copying response body: %v", err)
		}
	})
}
