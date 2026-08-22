package controllers

import (
	"encoding/json"
	"io"
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

func writeJSONError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

type ScanRequest struct {
	TargetID        string   `json:"target_id"`
	TargetDirectory string   `json:"target_directory"`
	TLSTargets      []string `json:"tls_targets,omitempty"`
}

// getSessionID extracts the client session ID from the X-Session-ID header.
// Falls back to "default" for direct API calls without a session header.
func getSessionID(r *http.Request) string {
	id := strings.TrimSpace(r.Header.Get("X-Session-ID"))
	if id == "" || len(id) > 128 {
		return "default"
	}
	return id
}

// svc returns a session-scoped CBOMService for this request.
func (c *CBOMController) svc(r *http.Request) *services.CBOMService {
	return c.service.ForSession(getSessionID(r))
}


// HandleUploadScan handles POST /api/v1/scans/upload (multipart/form-data)
func (c *CBOMController) HandleUploadScan(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit multipart memory buffer to 128 MB (excess stored in temp files)
	if err := r.ParseMultipartForm(128 << 20); err != nil {
		writeJSONError(w, "Failed to parse uploaded form data: "+err.Error(), http.StatusBadRequest)
		return
	}

	folderName := r.FormValue("folder_name")
	if folderName == "" {
		folderName = "uploaded_project"
	}

	tlsTargetsRaw := r.FormValue("tls_targets")
	var tlsTargets []string
	if tlsTargetsRaw != "" {
		for _, t := range strings.Split(tlsTargetsRaw, ",") {
			if trimmed := strings.TrimSpace(t); trimmed != "" {
				tlsTargets = append(tlsTargets, trimmed)
			}
		}
	}

	var files []services.UploadedFile

	// Extract files from multipart form
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, headers := range r.MultipartForm.File {
			for _, fh := range headers {
				file, err := fh.Open()
				if err != nil {
					continue
				}
				content, err := io.ReadAll(file)
				file.Close()
				if err != nil {
					continue
				}

				relPath := fh.Filename
				files = append(files, services.UploadedFile{
					RelativePath: relPath,
					Content:      content,
				})
			}
		}
	}

	if len(files) == 0 {
		writeJSONError(w, "No valid files received in upload", http.StatusBadRequest)
		return
	}

	detail, err := c.svc(r).ExecuteScanFromUploadedFiles(files, folderName, tlsTargets)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(detail)
}

// HandleTargets handles GET /api/v1/targets and POST /api/v1/targets
func (c *CBOMController) HandleTargets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		targets, err := c.svc(r).ListTargets()
		if err != nil {
			writeJSONError(w, "Failed to retrieve targets", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(targets)

	case http.MethodPost:
		var target models.Target
		if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
			writeJSONError(w, "Invalid request payload", http.StatusBadRequest)
			return
		}
		created, err := c.svc(r).CreateTarget(target)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(created)

	default:
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleTargetDetail handles GET /api/v1/targets/{id} and /api/v1/targets/{id}/scans
func (c *CBOMController) HandleTargetDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSONError(w, "Target ID required", http.StatusBadRequest)
		return
	}

	targetID := parts[0]

	if len(parts) > 1 {
		switch parts[1] {
		case "scans":
			scans, err := c.svc(r).GetScansByTarget(targetID)
			if err != nil {
				writeJSONError(w, err.Error(), http.StatusNotFound)
				return
			}
			json.NewEncoder(w).Encode(scans)
			return

		case "scan":
			if r.Method != http.MethodPost {
				writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}
			// Parse optional TLS targets from request body
			var scanReq ScanRequest
			if r.Body != nil {
				_ = json.NewDecoder(r.Body).Decode(&scanReq)
			}
			detail, err := c.svc(r).ExecuteScanByTargetWithTLS(targetID, scanReq.TLSTargets)
			if err != nil {
				writeJSONError(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(detail)
			return
		}
	}

	target, err := c.svc(r).GetTarget(targetID)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(target)
}

// HandleScans handles GET /api/v1/scans and POST /api/v1/scans
func (c *CBOMController) HandleScans(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		scans, err := c.svc(r).ListScans()
		if err != nil {
			writeJSONError(w, "Failed to retrieve scans", http.StatusInternalServerError)
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
			detail, err = c.svc(r).ExecuteScanByTarget(req.TargetID)
		} else {
			detail, err = c.svc(r).ExecuteScan(req.TargetDirectory)
		}

		if err != nil {
			writeJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(detail)

	default:
		writeJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleScanDetail handles GET /api/v1/scans/{id} and subroutes
func (c *CBOMController) HandleScanDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	path := strings.TrimPrefix(r.URL.Path, "/api/v1/scans/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeJSONError(w, "Scan ID required", http.StatusBadRequest)
		return
	}

	scanID := parts[0]

	// Handle /api/v1/scans/{id}/cyclonedx
	if len(parts) > 1 && parts[1] == "cyclonedx" {
		format := r.URL.Query().Get("format")
		if format == "xml" {
		xmlStr, err := c.svc(r).GenerateCycloneDXXML(scanID)
			if err != nil {
				writeJSONError(w, err.Error(), http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			w.Header().Set("Content-Disposition", `attachment; filename="cbom-cyclonedx-`+scanID+`.xml"`)
			w.Write([]byte(xmlStr))
			return
		}

		bom, err := c.svc(r).GenerateCycloneDX(scanID)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Disposition", `attachment; filename="cbom-cyclonedx-`+scanID+`.json"`)
		json.NewEncoder(w).Encode(bom)
		return
	}

	// Handle /api/v1/scans/{id}/migration
	if len(parts) > 1 && parts[1] == "migration" {
		plan, err := c.svc(r).GetMigrationPlan(scanID)
		if err != nil {
			writeJSONError(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(plan)
		return
	}

	// Default GET /api/v1/scans/{id}
	detail, err := c.svc(r).GetScanDetail(scanID)
	if err != nil {
		writeJSONError(w, err.Error(), http.StatusNotFound)
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
