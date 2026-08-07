package controllers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/models"
	"github.com/quantumpq/qv-cbom/internal/services"
)

type CBOMController struct {
	service *services.CBOMService
}

func NewCBOMController(service *services.CBOMService) *CBOMController {
	return &CBOMController{service: service}
}

type ScanRequest struct {
	TargetID        string `json:"target_id"`
	TargetDirectory string `json:"target_directory"`
}

// HandleTargets handles GET /api/v1/targets and POST /api/v1/targets
func (c *CBOMController) HandleTargets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		targets, err := c.service.ListTargets()
		if err != nil {
			http.Error(w, `{"error":"Failed to retrieve targets"}`, http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(targets)

	case http.MethodPost:
		var target models.Target
		if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
			http.Error(w, `{"error":"Invalid request payload"}`, http.StatusBadRequest)
			return
		}
		created, err := c.service.CreateTarget(target)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)

	default:
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// HandleTargetDetail handles GET /api/v1/targets/{id} and /api/v1/targets/{id}/scans
func (c *CBOMController) HandleTargetDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, `{"error":"Target ID required"}`, http.StatusBadRequest)
		return
	}

	targetID := parts[0]

	if len(parts) > 1 && parts[1] == "scans" {
		scans, err := c.service.GetScansByTarget(targetID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(scans)
		return
	}

	target, err := c.service.GetTarget(targetID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(target)
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

		var detail *models.ScanDetail
		var err error

		if req.TargetID != "" {
			detail, err = c.service.ExecuteScanByTarget(req.TargetID)
		} else {
			detail, err = c.service.ExecuteScan(req.TargetDirectory)
		}

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
		"version": "2.0.0",
	})
}
