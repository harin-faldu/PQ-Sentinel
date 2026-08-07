package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/quantumpq/qv-cbom/internal/controllers"
	"github.com/quantumpq/qv-cbom/internal/middleware"
	"github.com/quantumpq/qv-cbom/internal/repository"
	"github.com/quantumpq/qv-cbom/internal/services"
	"github.com/quantumpq/qv-cbom/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting QuantumVault CBOM Engine on port %s...", cfg.Port)

	// Attempt MySQL connection; fallback to thread-safe Memory repository if MySQL is offline
	var repo repository.CBOMRepository
	mysqlRepo, err := repository.NewMySQLRepository(cfg.DBSource)
	if err != nil {
		log.Printf("Notice: MySQL connection not available (%v). Falling back to Memory Storage.", err)
		repo = repository.NewMemoryRepository()
	} else {
		log.Println("Successfully connected to MySQL database.")
		repo = mysqlRepo
	}

	cbomService := services.NewCBOMService(repo)
	cbomController := controllers.NewCBOMController(cbomService)

	mux := http.NewServeMux()

	// API Routing
	mux.HandleFunc("/api/v1/health", cbomController.HandleHealth)
	mux.HandleFunc("/api/v1/targets", cbomController.HandleTargets)
	mux.HandleFunc("/api/v1/targets/", cbomController.HandleTargetDetail)
	mux.HandleFunc("/api/v1/scans", cbomController.HandleScans)
	mux.HandleFunc("/api/v1/scans/", cbomController.HandleScanDetail)

	// Apply Middlewares: CORS -> Logger -> Recovery -> Handler
	handler := middleware.CORS(middleware.Logger(middleware.Recovery(mux)))

	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("QuantumVault CBOM API Server running at http://localhost%s", serverAddr)
	if err := http.ListenAndServe(serverAddr, handler); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
