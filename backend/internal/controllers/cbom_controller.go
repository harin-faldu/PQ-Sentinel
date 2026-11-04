package controllers

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/models"
	"github.com/quantumpq/qv-cbom/internal/services"
)

const (
	MaxTotalUploadSize = 100 << 20 // 100 MB Max Total Project Upload
	MaxSingleFileSize  = 10 << 20  // 10 MB Max Per Single File
	MaxUploadFileCount = 5000      // 5,000 Files Max
)

var allowedExtensions = map[string]bool{
	".go": true, ".js": true, ".jsx": true, ".ts": true, ".tsx": true, ".mjs": true, ".cjs": true,
	".py": true, ".pyw": true, ".java": true, ".rs": true, ".c": true, ".cpp": true, ".cc": true,
	".cxx": true, ".h": true, ".hpp": true, ".cs": true, ".rb": true, ".php": true, ".sol": true,
	".kt": true, ".scala": true, ".swift": true, ".dart": true, ".lua": true,
	".json": true, ".yaml": true, ".yml": true, ".toml": true, ".xml": true, ".conf": true,
	".ini": true, ".properties": true, ".env.example": true, ".gitignore": true,
	".pem": true, ".crt": true, ".cer": true, ".key": true, ".csr": true, ".der": true,
	".p12": true, ".pfx": true, ".jks": true, ".pub": true, ".kdbx": true, ".asc": true,
}

var allowedFilenames = map[string]bool{
	"dockerfile": true, "containerfile": true, "makefile": true, "gemfile": true,
	"go.mod": true, "go.sum": true, "package.json": true, "package-lock.json": true,
	"yarn.lock": true, "pnpm-lock.yaml": true, "cargo.toml": true, "cargo.lock": true,
	"pom.xml": true, "build.gradle": true, "build.gradle.kts": true, "requirements.txt": true,
	"pipfile": true, "pipfile.lock": true, "pyproject.toml": true, "setup.py": true,
	"composer.json": true, "composer.lock": true, "cmakelists.txt": true,
}

var blockedExtensions = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true, ".msi": true,
	".iso": true, ".img": true, ".bat": true, ".cmd": true, ".ps1": true, ".vbs": true,
	".vbe": true, ".wsf": true, ".wsh": true, ".scr": true, ".com": true, ".pif": true,
	".hta": true, ".cpl": true, ".msc": true, ".jar": true, ".war": true, ".class": true,
	".pyc": true, ".pyo": true, ".pyd": true, ".elf": true, ".o": true, ".a": true, ".lib": true,
}

// sanitizeRelativePath cleanses path traversal attempts (e.g. ../, ..\)
func sanitizeRelativePath(rawPath string) string {
	cleaned := strings.ReplaceAll(rawPath, "\\", "/")
	for strings.Contains(cleaned, "../") {
		cleaned = strings.ReplaceAll(cleaned, "../", "")
	}
	cleaned = strings.TrimPrefix(cleaned, "/")
	return path.Clean(cleaned)
}

// isExecutableBinary checks for known binary executable magic byte signatures
func isExecutableBinary(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// Windows PE (MZ)
	if data[0] == 'M' && data[1] == 'Z' {
		return true
	}
	// Linux ELF (\x7fELF)
	if data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
		return true
	}
	// Mach-O (macOS binaries / Java class)
	if (data[0] == 0xca && data[1] == 0xfe && data[2] == 0xba && data[3] == 0xbe) ||
		(data[0] == 0xfe && data[1] == 0xed && data[2] == 0xfa && (data[3] == 0xce || data[3] == 0xcf)) ||
		(data[0] == 0xce && data[1] == 0xfa && data[2] == 0xed && data[3] == 0xfe) ||
		(data[0] == 0xcf && data[1] == 0xfa && data[2] == 0xed && data[3] == 0xfe) {
		return true
	}
	return false
}

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

	// Limit multipart memory buffer to MaxTotalUploadSize (100 MB)
	if err := r.ParseMultipartForm(MaxTotalUploadSize); err != nil {
		writeJSONError(w, "Upload rejected: "+err.Error(), http.StatusBadRequest)
		return
	}

	folderName := sanitizeRelativePath(r.FormValue("folder_name"))
	if folderName == "" || folderName == "." {
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
	var totalBytes int64 = 0

	// Extract, sanitize, and validate files from multipart form
	if r.MultipartForm != nil && r.MultipartForm.File != nil {
		for _, headers := range r.MultipartForm.File {
			for _, fh := range headers {
				// 1. Check max file count cap
				if len(files) >= MaxUploadFileCount {
					writeJSONError(w, "Upload exceeds the max 5,000 files limit.", http.StatusBadRequest)
					return
				}

				// 2. Check individual file size limit
				if fh.Size > MaxSingleFileSize {
					continue // Safely skip files > 10MB
				}

				rawName := fh.Filename
				relPath := sanitizeRelativePath(rawName)
				lowerPath := strings.ToLower(relPath)
				baseName := path.Base(lowerPath)
				ext := path.Ext(baseName)

				// 3. Reject dangerous executables
				if blockedExtensions[ext] {
					continue // Discard malicious/binary executables
				}

				// 4. Whitelist check
				if !allowedExtensions[ext] && !allowedFilenames[baseName] {
					continue // Discard non-source files
				}

				file, err := fh.Open()
				if err != nil {
					continue
				}
				content, err := io.ReadAll(file)
				file.Close()
				if err != nil {
					continue
				}

				// 5. Binary magic byte inspection (prevent renamed executables)
				if isExecutableBinary(content) {
					continue
				}

				totalBytes += int64(len(content))
				if totalBytes > MaxTotalUploadSize {
					writeJSONError(w, "Total upload size exceeds 100 MB limit.", http.StatusBadRequest)
					return
				}

				files = append(files, services.UploadedFile{
					RelativePath: relPath,
					Content:      content,
				})
			}
		}
	}

	if len(files) == 0 {
		writeJSONError(w, "No valid source code files received in upload (executable binaries and unsupported formats were filtered out).", http.StatusBadRequest)
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
		target.Location = sanitizeRelativePath(target.Location)
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

	pathStr := strings.TrimPrefix(r.URL.Path, "/api/v1/targets/")
	parts := strings.Split(pathStr, "/")
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

		req.TargetDirectory = sanitizeRelativePath(req.TargetDirectory)

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

	pathStr := strings.TrimPrefix(r.URL.Path, "/api/v1/scans/")
	parts := strings.Split(pathStr, "/")
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
		"service": "PQ-Sentinel Engine",
		"version": "2.0.0",
	})
}
