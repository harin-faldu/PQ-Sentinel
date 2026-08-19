package main

import (
	"log"
	"net"
	"net/http"

	"github.com/quantumpq/qv-cbom/internal/controllers"
	"github.com/quantumpq/qv-cbom/internal/middleware"
	"github.com/quantumpq/qv-cbom/internal/repository"
	"github.com/quantumpq/qv-cbom/internal/services"
	"github.com/quantumpq/qv-cbom/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting QuantumVault CBOM Engine on port %s (Zero-DB / In-Memory Mode)...", cfg.Port)

	// In-Memory Repository (Ephemeral single-scan storage, no database required)
	repo := repository.NewMemoryRepository()
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

	port := cfg.Port
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil && port == "8080" {
		log.Printf("Notice: Port 8080 is occupied by another process. Fallback binding to port 8081...")
		port = "8081"
		listener, err = net.Listen("tcp", ":"+port)
	}
	if err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}

	log.Printf("QuantumVault CBOM API Server running at http://localhost:%s", port)
	if err := http.Serve(listener, handler); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
