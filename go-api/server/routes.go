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
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/", handlers.IndexHandler(cfg, logger))
	mux.Handle("/dashboard", handlers.DashboardHandler(cfg, logger))
	mux.Handle("/home", handlers.HomeHandler(cfg, logger))

	mux.Handle("/sidebar", handlers.SidebarHandler(cfg, logger))
	mux.Handle("/abs_dataflow/", handlers.ABSDataflowHandler(cfg, logger, db))
	//helper routes
	mux.Handle("/health", handlers.HealthHandler(cfg, logger))

	mux.Handle("/request-data/ABS/", handlers.RequestABSData(cfg, logger))
	//plotting routes
	mux.Handle("/request-dashboard/", handlers.RequestDashboardHandler(cfg, logger))
	mux.Handle("/refresh-dashboard/", handlers.RefreshDashboardhandler(cfg, logger, db))
	mux.Handle("/plot/", handlers.PlotHandler(cfg, logger, db))

	mux.Handle("/plot/test/", handlers.PlotTestHandler(cfg, logger))
	mux.Handle("/plot/test/json/", handlers.PlotTestJSONHandler(cfg, logger))

}
