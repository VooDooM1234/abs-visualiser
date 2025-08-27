package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/VooDooM1234/abs-visualiser/go-api/config"
	"github.com/VooDooM1234/abs-visualiser/go-api/db"
)

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#maker-funcs-return-the-handler
// func handleSomething(config *config.Config, logger *log.Logger) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         // handler logic
//     })
// }

type DataflowABS struct {
	// Status       string `json:"status"`
	DataflowID   string `json:"dataflowid"`
	DataflowName string `json:"dataflowname"`
}

// seperate the fetching and handling
func RequestDataflowABS(config *config.Config, logger *log.Logger) http.Handler {
	path := config.HTMLTemplates + "dataflow_contents.html"
	tmpl := template.Must(template.ParseFiles(path))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := fmt.Sprintf("http://%s:%d/request-dataflow/ABS/", config.Host, config.PlotServicePort)
		logger.Printf("GET request to: %s", url)
		resp, err := http.Get(url)
		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			logger.Printf("Failed to GET ABS dataflows: %v", err)
			return
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Printf("Error reading response body: %v", err)
			http.Error(w, "Failed to read response", http.StatusInternalServerError)
			return
		}
		// fix genric error reponse handler
		// if err := utils.CheckFailureResponse(body, logger); err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	log.Printf("Failed to check generic response: %v", err)
		// 	return
		// }

		var result []DataflowABS
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Failed to parse JSON: %v", err)
			http.Error(w, "Invalid response format", http.StatusInternalServerError)
			return
		}

		// move the serviering to a handler func
		if err := tmpl.Execute(w, result); err != nil {
			http.Error(w, "Failed to render template", http.StatusInternalServerError)
			logger.Printf("Template execution error: %v", err)
			return
		}

	})
}

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

		dataflowid := r.URL.Query().Get("dataflowid")
		if dataflowid == "" {
			dataflowid = "CPI"
		}

		if err := tmpl.Execute(w, dataflowid); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			logger.Printf("Error executing dashboard template: %v", err)
		}
		log.Println("Dashboard Page")
	})
}

// Reverse proxy for Ploty Dash (Python mservice)
func ReverseProxyDashHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target, err := url.Parse(fmt.Sprintf("http://%s:%d", cfg.PlotServiceHost, cfg.PlotServicePort))
		if err != nil {
			logger.Printf("Dash Reverse Proxy parse error: %v", err)
			http.Error(w, "bad upstream", http.StatusInternalServerError)
			return
		}

		proxy := httputil.NewSingleHostReverseProxy(target)
		// allow iframe serving for dashboard
		proxy.ModifyResponse = func(res *http.Response) error {
			res.Header.Del("X-Frame-Options")
			res.Header.Del("Content-Security-Policy")
			return nil
		}

		proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
			logger.Printf("Proxy error: %v", err)
			http.Error(rw, "Upstream unavailable", http.StatusBadGateway)
		}

		proxy.ServeHTTP(w, r)
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
		microserviceurl := fmt.Sprintf("http://%s:%d/request-data/ABS/", config.PlotServiceHost, config.PlotServicePort)
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

type RefreshDashboardResponse struct {
	Status string `json:"status"`
	// DataflowID string `json:"dataflowid"`
	// Timestamp  string `json:"timestamp"`
}

func RefreshDashboardhandler(config *config.Config, logger *log.Logger, db *db.Database) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		dataflowid := "CPI"
		var payload = map[string]string{
			"dataflowid": dataflowid,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			logger.Printf("Failed to marshal payload: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger.Printf("Retrieving data for dataflow: %s...", dataflowid)

		url := fmt.Sprintf("http://%s:%s/refresh-dashboard/", config.Host, "54850")
		logger.Printf("POST request to: %s", url)
		// trying different method of posting rather then http.NewRequest and a DO method like most articles seem to say idk why they dont use this when is seems easier?
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		var result RefreshDashboardResponse
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

	})
}

func GetDashboardHandler(cfg *config.Config, logger *log.Logger) http.Handler {
	path := cfg.HTMLTemplates + "dashboard.html"
	tmpl := template.Must(template.ParseFiles(path))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			logger.Printf("Invalid request: %v", err)
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		dataflowID := r.FormValue("dataflowid")
		if dataflowID == "" {
			logger.Printf("Missing dataflowid in request")
			http.Error(w, "Missing dataflowid", http.StatusBadRequest)
			return
		}

		// Step 1: Notify Dash microservice
		// url := fmt.Sprintf("http://%s:%s/refresh-dashboard/", cfg.PlotServiceHost, cfg.PlotServicePort)
		body, _ := json.Marshal(map[string]string{"dataflowid": dataflowID})
		url := "http://localhost:8082/dashboard/refresh-dashboard/"
		logger.Printf("Sending Payload: %s", body)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil || resp.StatusCode != http.StatusOK {
			logger.Printf("Failed to refresh dashboard data: %v", err)
			http.Error(w, "Failed to refresh dashboard data", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Step 2: Render dashboard fragment
		data := map[string]string{"DataflowID": dataflowID}
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, "Failed to render dashboard", http.StatusInternalServerError)
			logger.Printf("Dashboard template error: %v", err)
			return
		}
	})
}

func PlotTestHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		url := fmt.Sprintf("http://%s:%d/plot/test", config.Host, config.PlotServicePort)
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

		url := fmt.Sprintf("http://%s:%d/plot/test/json", config.Host, config.PlotServicePort)
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
