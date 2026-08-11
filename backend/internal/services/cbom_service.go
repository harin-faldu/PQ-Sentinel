package services

import (
	"crypto/rsa"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/quantumpq/qv-cbom/internal/github"
	"github.com/quantumpq/qv-cbom/internal/models"
	"github.com/quantumpq/qv-cbom/internal/repository"
	"github.com/quantumpq/qv-cbom/internal/risk"
	"github.com/quantumpq/qv-cbom/internal/scanner"
)

var purlInvalidCharsRegex = regexp.MustCompile(`[^a-z0-9\.\-\_]`)

func sanitizePURLString(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = filepath.Base(s)
	s = purlInvalidCharsRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "application"
	}
	return s
}

const scoreHistoryFile = "score-history.json"

type CBOMService struct {
	repo             repository.CBOMRepository
	scanner          *scanner.Scanner
	riskEngine       *risk.RiskEngine
	githubAppID      string
	githubPrivateKey *rsa.PrivateKey
}

func NewCBOMService(repo repository.CBOMRepository) *CBOMService {
	return &CBOMService{
		repo:       repo,
		scanner:    scanner.NewScanner(),
		riskEngine: risk.NewRiskEngine(),
	}
}

func (s *CBOMService) SetGitHubAppCredentials(appID string, privateKey *rsa.PrivateKey) {
	s.githubAppID = appID
	s.githubPrivateKey = privateKey
}

func (s *CBOMService) CreateTarget(target models.Target) (*models.Target, error) {
	if target.ID == "" {
		target.ID = fmt.Sprintf("target-%s", models.GenerateUUID()[:8])
	}
	if target.CreatedAt.IsZero() {
		target.CreatedAt = time.Now()
	}
	if err := s.repo.CreateTarget(target); err != nil {
		return nil, err
	}
	return s.repo.GetTarget(target.ID)
}

func (s *CBOMService) ListTargets() ([]models.Target, error) {
	return s.repo.ListTargets()
}

func (s *CBOMService) GetTarget(id string) (*models.Target, error) {
	return s.repo.GetTarget(id)
}

func (s *CBOMService) GetScansByTarget(targetID string) ([]models.ScanSummary, error) {
	return s.repo.GetScansByTarget(targetID)
}

func (s *CBOMService) ExecuteScanByTarget(targetID string) (*models.ScanDetail, error) {
	target, err := s.repo.GetTarget(targetID)
	if err != nil {
		return nil, fmt.Errorf("unknown target: %w", err)
	}

	var headSHA string

	// Early Skip Check for GitHub Repositories (Unchanged remote HEAD commit)
	if target.Type == models.TargetTypeGitHubRepo {
		appJWT, err := github.GenerateAppJWT(s.githubAppID, s.githubPrivateKey)
		if err == nil {
			tokenResp, err := github.GetInstallationToken(appJWT, target.InstallationID)
			if err == nil && tokenResp != nil {
				branch := target.DefaultBranch
				if branch == "" {
					branch = "main"
				}
				headSHA, err = github.GetRemoteHeadSHA(tokenResp.Token, target.FullName, branch)
				if err == nil && headSHA != "" && target.LastScannedSHA != "" && headSHA == target.LastScannedSHA {
					existingDetail, err := s.repo.GetLatestScanByTargetID(target.ID)
					if err == nil && existingDetail != nil {
						existingDetail.Summary.Notice = fmt.Sprintf("Repository commit unchanged (SHA: %s). Scan skipped.", headSHA)
						return existingDetail, nil
					}
				}
			}
		}
	}

	scanID := fmt.Sprintf("scan-%d", time.Now().Unix())
	startTime := time.Now()

	var assets []models.CryptoAsset
	var totalFiles, scannedFiles int

	switch target.Type {
	case models.TargetTypeRepo:
		// MODE 1: LOCAL DIRECTORY CODEBASE SCANNING
		assets, totalFiles, scannedFiles, err = s.scanner.ScanDirectory(target.Location, scanID)
		if err != nil {
			return nil, fmt.Errorf("directory scan failed: %w", err)
		}

	case models.TargetTypeGitHubRepo:
		// MODE 2: REMOTE GITHUB REPOSITORY SCANNING VIA GITHUB APP
		if s.githubAppID == "" || s.githubPrivateKey == nil {
			return nil, fmt.Errorf("GitHub App credentials not configured on server. Please check GITHUB_APP_ID and GITHUB_PRIVATE_KEY_PATH in .env")
		}

		// Auto-recover installation_id if target was created before GitHub App authorization
		if target.InstallationID <= 0 {
			if allTargets, err := s.repo.ListTargets(); err == nil {
				for _, t := range allTargets {
					if t.InstallationID > 0 {
						target.InstallationID = t.InstallationID
						break
					}
				}
			}
		}

		if target.InstallationID <= 0 {
			return nil, fmt.Errorf("missing or invalid GitHub Installation ID for target '%s'. Please click 'Install App on GitHub' to authorize access", target.Name)
		}

		appJWT, err := github.GenerateAppJWT(s.githubAppID, s.githubPrivateKey)
		if err != nil {
			return nil, fmt.Errorf("github app JWT generation failed: %w", err)
		}
		tokenResp, err := github.GetInstallationToken(appJWT, target.InstallationID)
		if err != nil {
			return nil, fmt.Errorf("github installation token exchange failed (Installation ID %d may be revoked/expired): %w", target.InstallationID, err)
		}

		branch := target.DefaultBranch
		if branch == "" {
			branch = "main"
		}

		tempDir, err := os.MkdirTemp("", "gh-scan-*")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp workspace: %w", err)
		}
		defer os.RemoveAll(tempDir)

		if err := github.CloneRepository(tokenResp.Token, target.FullName, branch, tempDir); err != nil {
			return nil, fmt.Errorf("github clone failed: %w", err)
		}

		assets, totalFiles, scannedFiles, err = s.scanner.ScanDirectory(tempDir, scanID)
		if err != nil {
			return nil, fmt.Errorf("github directory scan failed: %w", err)
		}

	case models.TargetTypeHost:
		assets = s.scanner.ScanTLSEndpoints([]string{target.Location}, scanID)
		totalFiles = 1
		scannedFiles = 1

	default:
		// Fallback to directory scan
		assets, totalFiles, scannedFiles, err = s.scanner.ScanDirectory(target.Location, scanID)
		if err != nil {
			return nil, fmt.Errorf("directory scan failed: %w", err)
		}
	}

	if target.Type == models.TargetTypeGitHubRepo && headSHA != "" {
		_ = s.repo.UpdateTargetLastScannedSHA(target.ID, headSHA)
	}

	return s.saveScanResults(target, scanID, assets, totalFiles, scannedFiles, startTime)
}

// Extracted shared helper method to compute PQC risk score, populate summary metrics, persist scan summary, migration plan & score history
func (s *CBOMService) saveScanResults(target *models.Target, scanID string, assets []models.CryptoAsset, totalFiles, scannedFiles int, startTime time.Time) (*models.ScanDetail, error) {
	pqcScore, critical, high, medium, safe := s.riskEngine.CalculatePQCScore(assets)
	summary := models.ScanSummary{
		ID:              scanID,
		TargetID:        target.ID,
		TargetDirectory: target.Location,
		TotalFiles:      totalFiles,
		ScannedFiles:    scannedFiles,
		TotalAssets:     len(assets),
		CriticalCount:   critical,
		HighCount:       high,
		MediumCount:     medium,
		SafeCount:       safe,
		PQCScore:        pqcScore,
		Status:          "COMPLETED",
		DurationMs:      time.Since(startTime).Milliseconds(),
		CreatedAt:       time.Now(),
	}

	if err := s.repo.SaveScan(summary, assets); err != nil {
		return nil, fmt.Errorf("failed to save scan to database: %w", err)
	}

	plan := s.riskEngine.GenerateMigrationPlan(scanID, assets)
	_ = s.repo.SaveMigrationPlan(plan)
	s.trackScoreHistory(summary)

	return &models.ScanDetail{
		Summary: summary,
		Assets:  assets,
	}, nil
}

func (s *CBOMService) ExecuteScan(targetDir string) (*models.ScanDetail, error) {
	return s.ExecuteScanWithTLS(targetDir, nil)
}

func (s *CBOMService) ExecuteScanWithTLS(targetDir string, tlsTargets []string) (*models.ScanDetail, error) {
	startTime := time.Now()
	scanID := fmt.Sprintf("scan-%d", time.Now().Unix())

	if targetDir == "" {
		targetDir = "."
	}

	assets, totalFiles, scannedFiles, err := s.scanner.ScanDirectory(targetDir, scanID)
	if err != nil {
		return nil, fmt.Errorf("directory scan failed: %w", err)
	}

	// Auto-discover and probe TLS endpoints if requested
	if len(tlsTargets) > 0 {
		var activeTargets []string
		for _, target := range tlsTargets {
			discovered := s.scanner.AutoDiscoverTLSEndpoints(target)
			activeTargets = append(activeTargets, discovered...)
		}
		tlsAssets := s.scanner.ScanTLSEndpoints(activeTargets, scanID)
		assets = append(assets, tlsAssets...)
	}

	pqcScore, critical, high, medium, safe := s.riskEngine.CalculatePQCScore(assets)

	summary := models.ScanSummary{
		ID:              scanID,
		TargetDirectory: targetDir,
		TotalFiles:      totalFiles,
		ScannedFiles:    scannedFiles,
		TotalAssets:     len(assets),
		CriticalCount:   critical,
		HighCount:       high,
		MediumCount:     medium,
		SafeCount:       safe,
		PQCScore:        pqcScore,
		Status:          "COMPLETED",
		DurationMs:      time.Since(startTime).Milliseconds(),
		CreatedAt:       time.Now(),
	}

	if err := s.repo.SaveScan(summary, assets); err != nil {
		return nil, fmt.Errorf("failed to save scan to database: %w", err)
	}

	plan := s.riskEngine.GenerateMigrationPlan(scanID, assets)
	_ = s.repo.SaveMigrationPlan(plan)

	// Track historical score
	s.trackScoreHistory(summary)

	return &models.ScanDetail{
		Summary: summary,
		Assets:  assets,
	}, nil
}

// trackScoreHistory appends the current scan score to score-history.json
func (s *CBOMService) trackScoreHistory(summary models.ScanSummary) {
	entry := models.ScoreHistoryEntry{
		ScanID:    summary.ID,
		Target:    summary.TargetDirectory,
		PQCScore:  summary.PQCScore,
		Critical:  summary.CriticalCount,
		High:      summary.HighCount,
		Medium:    summary.MediumCount,
		Safe:      summary.SafeCount,
		Timestamp: time.Now(),
	}

	// Load existing history
	history := loadScoreHistory()
	history = append(history, entry)

	// Write updated history
	data, err := json.MarshalIndent(history, "", "  ")
	if err == nil {
		_ = os.WriteFile(scoreHistoryFile, data, 0644)
	}
}

// GetScoreHistory returns all past score entries for trend display
func (s *CBOMService) GetScoreHistory() []models.ScoreHistoryEntry {
	return loadScoreHistory()
}

// loadScoreHistory reads the score history from disk
func loadScoreHistory() []models.ScoreHistoryEntry {
	var history []models.ScoreHistoryEntry
	data, err := os.ReadFile(scoreHistoryFile)
	if err != nil {
		return history
	}
	_ = json.Unmarshal(data, &history)
	return history
}

func (s *CBOMService) ListScans() ([]models.ScanSummary, error) {
	return s.repo.GetScans()
}

func (s *CBOMService) GetScanDetail(id string) (*models.ScanDetail, error) {
	return s.repo.GetScanByID(id)
}

func (s *CBOMService) GetMigrationPlan(scanID string) (*models.MigrationPlan, error) {
	return s.repo.GetMigrationPlanByScanID(scanID)
}

var nistLevelByAlgorithm = map[string]int{
	"ML-KEM-768":  3,
	"ML-KEM-1024": 5,
	"ML-DSA-65":   3,
	"ML-DSA-87":   5,
	"SLH-DSA-128": 1,
	"AES-256-GCM": 5,
	"AES-128-GCM": 1,
	"SHA-256":     0,
	"SHA-384":     0,
	"MD5":         0,
	"SHA-1":       0,
	"RSA":         0,
	"ECDSA":       0,
}

func (s *CBOMService) GenerateCycloneDX(scanID string) (*models.CycloneDXCBOM, error) {
	detail, err := s.repo.GetScanByID(scanID)
	if err != nil {
		return nil, err
	}

	cleanAppTarget := sanitizePURLString(detail.Summary.TargetDirectory)
	appRef := fmt.Sprintf("pkg:application/%s@1.0.0", cleanAppTarget)
	var components []models.CycloneDXComponent
	var cryptoRefs []string
	nowISO := time.Now().Format(time.RFC3339)

	for _, a := range detail.Assets {
		keySizeStr := fmt.Sprintf("%d", a.KeySize)
		if a.KeySize == 0 {
			keySizeStr = "unspecified"
		}
		// Valid generic PURL format per spec: pkg:generic/algo@keysize with slash sanitization
		cleanAlgoPURL := sanitizePURLString(a.Algorithm)
		purl := fmt.Sprintf("pkg:generic/%s@%s", cleanAlgoPURL, keySizeStr)
		cryptoRefs = append(cryptoRefs, purl)

		nistLevel := 0
		if level, ok := nistLevelByAlgorithm[a.Algorithm]; ok {
			nistLevel = level
		}

		comp := models.CycloneDXComponent{
			Type:        "cryptographic-asset",
			Name:        a.Name,
			PURL:        purl,
			EvaluatedAt: nowISO,
			CryptoProperties: &models.CycloneDXCryptoProps{
				AssetType: string(a.Type),
				AlgorithmProperties: &models.CycloneDXAlgoProps{
					Primitive:                a.Algorithm,
					ParameterSetID:           fmt.Sprintf("%d-bit", a.KeySize),
					NISTQuantumSecurityLevel: nistLevel,
				},
			},
		}
		components = append(components, comp)
	}

	dependencies := []models.CycloneDXDependency{
		{
			Ref:       appRef,
			DependsOn: cryptoRefs,
		},
	}

	bom := &models.CycloneDXCBOM{
		BOMFormat:    "CycloneDX",
		SpecVersion:  "1.6",
		SerialNumber: fmt.Sprintf("urn:uuid:qv-cbom-%s", scanID),
		Version:      1,
		Metadata: models.CycloneDXMetadata{
			Timestamp: nowISO,
			Tools: []models.CycloneDXTool{
				{Vendor: "QuantumVault", Name: "QV-CBOM Scanner", Version: "2.0.0"},
			},
			Component: models.CycloneDXComponent{
				Type: "application",
				Name: detail.Summary.TargetDirectory,
				PURL: appRef,
			},
		},
		Components:   components,
		Dependencies: dependencies,
	}

	return bom, nil
}

// GenerateCycloneDXXML generates standard CycloneDX 1.6 XML format CBOM
func (s *CBOMService) GenerateCycloneDXXML(scanID string) (string, error) {
	bom, err := s.GenerateCycloneDX(scanID)
	if err != nil {
		return "", err
	}

	xmlData, err := xml.MarshalIndent(bom, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal CycloneDX XML: %w", err)
	}

	xmlHeader := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n"
	return xmlHeader + string(xmlData), nil
}
