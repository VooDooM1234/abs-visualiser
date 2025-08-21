package server

import (
	"log"
	"net/http"

	"github.com/VooDooM1234/abs-visualiser/config"
	"github.com/VooDooM1234/abs-visualiser/db"
	"github.com/VooDooM1234/abs-visualiser/handlers"
)

func AddRoutes(
	mux *http.ServeMux,
	logger *log.Logger,
	cfg *config.Config,
	db *db.Database,
) {
	mux.Handle("/", handlers.HomePageHandler(cfg, logger))
	//helper routes
	mux.Handle("/health", handlers.HealthHander(cfg, logger))
	//plotting routes
	mux.Handle("/plot/bar-abs-cpi/", handlers.PlotABSCPIHandler(cfg, logger))
	mux.Handle("/plot/test/", handlers.PlotTestHandler(cfg, logger))
	mux.Handle("/plot/test/json/", handlers.PlotTestJSONHandler(cfg, logger))

}
