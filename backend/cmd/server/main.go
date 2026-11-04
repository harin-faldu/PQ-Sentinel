package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/quantumpq/qv-cbom/internal/controllers"
	"github.com/quantumpq/qv-cbom/internal/middleware"
	"github.com/quantumpq/qv-cbom/internal/repository"
	"github.com/quantumpq/qv-cbom/internal/services"
	"github.com/quantumpq/qv-cbom/pkg/config"
)

func main() {
	cfg := config.LoadConfig()

	log.Printf("Starting PQ-Sentinel Engine on port %s (Zero-DB / In-Memory Mode)...", cfg.Port)

	// Per-client Session Store (ephemeral, isolated memory buckets per browser session)
	sessionStore := repository.NewSessionStore()
	sessionStore.StartCleanup(2 * time.Hour) // Auto-evict inactive sessions after 2h
	cbomService := services.NewCBOMService(sessionStore)
	cbomController := controllers.NewCBOMController(cbomService)

	mux := http.NewServeMux()

	// API Routing
	mux.HandleFunc("/api/v1/health", cbomController.HandleHealth)
	mux.HandleFunc("/api/v1/targets", cbomController.HandleTargets)
	mux.HandleFunc("/api/v1/targets/", cbomController.HandleTargetDetail)
	mux.HandleFunc("/api/v1/scans/upload", cbomController.HandleUploadScan)
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

	log.Printf("PQ-Sentinel API Server running at http://localhost:%s", port)
	if err := http.Serve(listener, handler); err != nil {
		log.Fatalf("Server shutdown error: %v", err)
	}
}
