package handlers

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/VooDooM1234/abs-visualiser/go-api/config"
	"github.com/VooDooM1234/abs-visualiser/go-api/db"
)

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#maker-funcs-return-the-handler
// func handleSomething(config *config.Config, logger *Logger) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         // handler logic
//     })
// }s

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

// Plothandler endpoint /plot/{graphName}/{dataflow}
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
