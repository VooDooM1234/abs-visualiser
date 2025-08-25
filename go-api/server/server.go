package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/VooDooM1234/abs-visualiser/go-api/config"
	"github.com/VooDooM1234/abs-visualiser/go-api/db"
)

func NewServer(
	logger *log.Logger,
	cfg *config.Config,
	db *db.Database,
) http.Handler {
	mux := http.NewServeMux()

	AddRoutes(mux, logger, cfg, db)

	var handler http.Handler = mux
	// wrap middlewares here if you want

	return handler
}

func launchPythonMicroservice(config *config.Config) {
	cmd := exec.Command(
		config.PythonPath,
		"-m", "uvicorn",
		"plotapp.main:app",
		"--host", config.PlotServiceHost,
		"--port", config.PlotServicePort,
	)

	cmd.Env = append(os.Environ(),
		"PLOT_SERVICE_HOST="+config.PlotServiceHost,
		"PLOT_SERVICE_PORT="+config.PlotServicePort,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()
	if err != nil {
		log.Fatalf("Failed to launch Python microservice: %v", err)
	}

	log.Printf("Python microservice running at http://%s:%s\n", config.PlotServiceHost, config.PlotServicePort)
}

// STOP STEALIN MY PORTS
func KillPort(port string) error {
	psCommand := fmt.Sprintf(`$connections = Get-NetTCPConnection -LocalPort %s -ErrorAction SilentlyContinue;
		foreach ($c in $connections) {
			Stop-Process -Id $c.OwningProcess -Force -ErrorAction SilentlyContinue
		}`, port)

	cmd := exec.Command("powershell", "-Command", psCommand)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func Run(ctx context.Context, w io.Writer, args []string) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()
	logger := log.Default()
	config, err := config.Init()
	if err != nil {
		fmt.Println("Failed to load config:", err)
		return err
	}
	databaseConnect, err := db.NewDatabase(ctx)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		return err
	}
	defer databaseConnect.Close()

	srv := NewServer(
		logger,
		config,
		databaseConnect,
	)

	httpServer := &http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: srv,
	}

	// client := fetch.NewFetch("https", "data.api.abs.gov.au", 0)
	// err = client.ABSRestDataflowAll(databaseConnect)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	KillPort(config.PlotServicePort)
	KillPort(config.Port)
	launchPythonMicroservice(config)

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "error listening and serving: %s\n", err)
		}
	}()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(os.Stderr, "error shutting down http server: %s\n", err)
		}
	}()
	wg.Wait()
	return err
}
