package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/services"
)

type CBOMController struct {
	service *services.CBOMService
}

func NewCBOMController(service *services.CBOMService) *CBOMController {
	return &CBOMController{service: service}
}

type ScanRequest struct {
	TargetDirectory string `json:"target_directory"`
}

// HandleScans handles GET /api/v1/scans and POST /api/v1/scans
func (c *CBOMController) HandleScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		scans, err := c.service.ListScans()
		if err != nil {
			http.Error(w, `{"error":"Failed to retrieve scans"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(scans)

	case http.MethodPost:
		var req ScanRequest
		if r.Body != nil {
			_ = json.NewDecoder(r.Body).Decode(&req)
		}

		detail, err := c.service.ExecuteScan(req.TargetDirectory)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(detail)

	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleScanDetail handles GET /api/v1/scans/{id} and subroutes
func (c *CBOMController) HandleScanDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/scans/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"Scan ID required"}`, http.StatusBadRequest)
		return
	}

	scanID := parts[0]

	// Handle /api/v1/scans/{id}/cyclonedx
	if len(parts) > 1 && parts[1] == "cyclonedx" {
		bom, err := c.service.GenerateCycloneDX(scanID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Disposition", `attachment; filename="cbom-cyclonedx-`+scanID+`.json"`)
		json.NewEncoder(w).Encode(bom)
		return
	}

	// Handle /api/v1/scans/{id}/migration
	if len(parts) > 1 && parts[1] == "migration" {
		plan, err := c.service.GetMigrationPlan(scanID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	}

	// Default GET /api/v1/scans/{id}
	detail, err := c.service.GetScanDetail(scanID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(detail)
}

// HandleHealth check endpoint
func (c *CBOMController) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "QuantumVault CBOM Engine",
		"version": "1.0.0",
	})
}
