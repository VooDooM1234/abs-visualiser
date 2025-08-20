package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/VooDooM1234/abs-visualiser/config"
)

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#maker-funcs-return-the-handler
// func handleSomething(config *config.Config, logger *Logger) http.Handler {
//     return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//         // handler logic
//     })
// }

func decode[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}

func encode[T any](w http.ResponseWriter, status int, v T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		return fmt.Errorf("encode json: %w", err)
	}
	return nil
}

func HomePageHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		tmpl := template.Must(template.ParseFiles("../templates/index.html"))

		if err := tmpl.Execute(w, nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		log.Print("Home Page")
	})
}

func PlotABSCPIHandler(config *config.Config, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		url := fmt.Sprintf("http://%s:%s/plot/bar-abs-cpi", config.Host, config.PlotServicePort)
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
