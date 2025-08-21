package server

import (
	"log"
	"net/http"

	"github.com/VooDooM1234/abs-visualiser/go-api/config"
	"github.com/VooDooM1234/abs-visualiser/go-api/db"
	"github.com/VooDooM1234/abs-visualiser/go-api/handlers"
)

func AddRoutes(
	mux *http.ServeMux,
	logger *log.Logger,
	cfg *config.Config,
	db *db.Database,
) {
	mux.Handle("/", handlers.HomePageHandler(cfg, logger))
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static"))))
	mux.Handle("/sidebar", handlers.SidebarHandler(cfg, logger))
	mux.Handle("/abs_dataflow/", handlers.ABSDataflowHandler(cfg, logger, db))
	//helper routes
	mux.Handle("/health", handlers.HealthHander(cfg, logger))
	//plotting routes
	mux.Handle("/plot/bar-abs-cpi/", handlers.PlotABSCPIHandler(cfg, logger))
	mux.Handle("/plot/", handlers.PlotHandler(cfg, logger, db))

	mux.Handle("/plot/test/", handlers.PlotTestHandler(cfg, logger))
	mux.Handle("/plot/test/json/", handlers.PlotTestJSONHandler(cfg, logger))

}
