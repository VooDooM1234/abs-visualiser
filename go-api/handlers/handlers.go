package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
)

// https://grafana.com/blog/2024/02/09/how-i-write-http-services-in-go-after-13-years/#maker-funcs-return-the-handler
// func handleSomething(logger *Logger) http.Handler {
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

// func NewPlotABSCPIHandler(logger *log.Logger) http.Handler {
// 	tmpl := template.Must(template.ParseFiles("../templates/div_frags/bar_plot_abs_cpi.html"))

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		w.Header().Set("Content-Type", "text/html")

// 		// You could also pass dynamic data to the template here
// 		if err := tmpl.Execute(w, nil); err != nil {
// 			http.Error(w, err.Error(), http.StatusInternalServerError)
// 			logger.Println("Error rendering template:", err)
// 			return
// 		}

// 		logger.Println("Served ABS CPI plot page")
// 	})
// }

type PlotlyChart struct {
	X    []string  `json:"x"`
	Y    []float64 `json:"y"`
	Type string    `json:"type"`
}

type PlotlyTemplateData struct {
	X    template.JS
	Y    template.JS
	Type string
}

func NewPlotABSCPIHandler() http.Handler {
	tmpl := template.Must(template.ParseFiles("../templates/div_frags/bar_plot_abs_cpi.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		// Example data, replace with your actual data source
		x := []string{"2024-Q1", "2024-Q2", "2024-Q3"}
		y := []float64{123.45, 124.56, 125.67}
		xJSON, _ := json.Marshal(x)
		yJSON, _ := json.Marshal(y)

		data := PlotlyTemplateData{
			X:    template.JS(xJSON),
			Y:    template.JS(yJSON),
			Type: "bar",
		}

		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			// logger.Println("Error rendering template:", err)
			return
		}

		// logger.Println("Served ABS CPI plot page")
	})
}

func PlotTestHandler(logger *log.Logger) http.Handler {
	// tmpl := template.Must(template.ParseFiles("../templates/div_frags/bar_plot_abs_cpi.html"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")

		resp, err := http.Get("http://localhost:8081/plot/test")

		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}

		defer resp.Body.Close()
		htmlBytes, _ := io.ReadAll(resp.Body)
		htmlString := string(htmlBytes)
		fmt.Fprint(w, htmlString)

		// logger.Println("Served ABS CPI plot page")
	})
}

func PlotTestJSONHandler(logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		resp, err := http.Get("http://localhost:8081/plot/test/json")
		if err != nil {
			http.Error(w, "Python service unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		io.Copy(w, resp.Body)
	})
}
