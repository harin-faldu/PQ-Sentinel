package services

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
	"github.com/quantumpq/qv-cbom/internal/repository"
	"github.com/quantumpq/qv-cbom/internal/risk"
	"github.com/quantumpq/qv-cbom/internal/scanner"
)

const scoreHistoryFile = "score-history.json"

type CBOMService struct {
	repo       repository.CBOMRepository
	scanner    *scanner.Scanner
	riskEngine *risk.RiskEngine
}

func NewCBOMService(repo repository.CBOMRepository) *CBOMService {
	return &CBOMService{
		repo:       repo,
		scanner:    scanner.NewScanner(),
		riskEngine: risk.NewRiskEngine(),
	}
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

func (s *CBOMService) GenerateCycloneDX(scanID string) (*models.CycloneDXCBOM, error) {
	detail, err := s.repo.GetScanByID(scanID)
	if err != nil {
		return nil, err
	}

	appRef := fmt.Sprintf("pkg:application/%s@1.0.0", detail.Summary.TargetDirectory)
	var components []models.CycloneDXComponent
	var cryptoRefs []string

	for _, a := range detail.Assets {
		purl := fmt.Sprintf("pkg:crypto/%s@%d", a.Algorithm, a.KeySize)
		cryptoRefs = append(cryptoRefs, purl)

		comp := models.CycloneDXComponent{
			Type: "cryptographic-asset",
			Name: a.Name,
			PURL: purl,
			CryptoProperties: &models.CycloneDXCryptoProps{
				AssetType: string(a.Type),
				AlgorithmProperties: &models.CycloneDXAlgoProps{
					Primitive:                a.Algorithm,
					ParameterSetID:           fmt.Sprintf("%d-bit", a.KeySize),
					NISTQuantumSecurityLevel: 1,
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
			Timestamp: time.Now().Format(time.RFC3339),
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
