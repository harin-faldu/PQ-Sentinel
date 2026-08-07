package services

import (
	"strings"
	"testing"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
	"github.com/quantumpq/qv-cbom/internal/repository"
)

func TestCycloneDXSpecCompliance(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := NewCBOMService(repo)

	// Seed test scan with crypto assets of different algorithms
	scanID := "test-cyclonedx-scan"
	summary := models.ScanSummary{
		ID:              scanID,
		TargetDirectory: "sample_target",
		TotalFiles:      10,
		ScannedFiles:    10,
		TotalAssets:     3,
		PQCScore:        50,
		Status:          "COMPLETED",
		CreatedAt:       time.Now(),
	}

	assets := []models.CryptoAsset{
		{
			ID:        "asset-rsa",
			ScanID:    scanID,
			Name:      "RSA Key Generation",
			AssetType: models.AssetAlgorithm,
			Algorithm: "RSA",
			KeySize:   2048,
		},
		{
			ID:        "asset-pqc",
			ScanID:    scanID,
			Name:      "Kyber Encapsulation",
			AssetType: models.AssetAlgorithm,
			Algorithm: "ML-KEM-768",
			KeySize:   768,
		},
		{
			ID:        "asset-aes",
			ScanID:    scanID,
			Name:      "AES Cipher",
			AssetType: models.AssetAlgorithm,
			Algorithm: "AES-256-GCM",
			KeySize:   256,
		},
	}

	if err := repo.SaveScan(summary, assets); err != nil {
		t.Fatalf("Failed to save test scan: %v", err)
	}

	// Generate CycloneDX 1.6 JSON CBOM
	bom, err := service.GenerateCycloneDX(scanID)
	if err != nil {
		t.Fatalf("GenerateCycloneDX failed: %v", err)
	}

	// 1. Spec Version Assertion
	if bom.SpecVersion != "1.6" {
		t.Errorf("Expected SpecVersion 1.6, got %s", bom.SpecVersion)
	}

	// 2. Component PURL & NIST Security Level Assertions
	if len(bom.Components) != 3 {
		t.Fatalf("Expected 3 components, got %d", len(bom.Components))
	}

	for _, comp := range bom.Components {
		// Valid PURL format assertion: pkg:generic/algo@keysize
		if !strings.HasPrefix(comp.PURL, "pkg:generic/") {
			t.Errorf("Invalid PURL format per spec: %s (must start with pkg:generic/)", comp.PURL)
		}

		if comp.EvaluatedAt == "" {
			t.Errorf("Component %s is missing EvaluatedAt timestamp", comp.Name)
		}

		// Dynamic NIST Level verification
		if comp.CryptoProperties != nil && comp.CryptoProperties.AlgorithmProperties != nil {
			primitive := comp.CryptoProperties.AlgorithmProperties.Primitive
			level := comp.CryptoProperties.AlgorithmProperties.NISTQuantumSecurityLevel

			switch primitive {
			case "ML-KEM-768":
				if level != 3 {
					t.Errorf("Expected NIST Level 3 for ML-KEM-768, got %d", level)
				}
			case "AES-256-GCM":
				if level != 5 {
					t.Errorf("Expected NIST Level 5 for AES-256-GCM, got %d", level)
				}
			case "RSA":
				if level != 0 {
					t.Errorf("Expected NIST Level 0 for RSA (vulnerable), got %d", level)
				}
			}
		}
	}

	// 3. XML Export Verification
	xmlStr, err := service.GenerateCycloneDXXML(scanID)
	if err != nil {
		t.Fatalf("GenerateCycloneDXXML failed: %v", err)
	}
	if !strings.Contains(xmlStr, "<?xml version=\"1.0\"") || !strings.Contains(xmlStr, "<CycloneDXCBOM>") {
		t.Errorf("Invalid CycloneDX XML export string format")
	}
}

func TestTargetEntityManagement(t *testing.T) {
	repo := repository.NewMemoryRepository()
	service := NewCBOMService(repo)

	// 1. Create Target
	target := models.Target{
		Name:     "Payment API Microservice",
		Type:     models.TargetTypeRepo,
		Location: "sample_target",
	}

	created, err := service.CreateTarget(target)
	if err != nil {
		t.Fatalf("CreateTarget failed: %v", err)
	}

	if created.ID == "" {
		t.Fatalf("Expected target ID to be generated, got empty string")
	}

	// 2. List Targets
	targets, err := service.ListTargets()
	if err != nil {
		t.Fatalf("ListTargets failed: %v", err)
	}
	if len(targets) < 1 {
		t.Fatalf("Expected at least 1 target, got %d", len(targets))
	}

	// 3. Get Target
	fetched, err := service.GetTarget(created.ID)
	if err != nil {
		t.Fatalf("GetTarget failed: %v", err)
	}
	if fetched.Name != target.Name {
		t.Errorf("Expected target name '%s', got '%s'", target.Name, fetched.Name)
	}

	// 4. Execute Scan by Target ID
	detail, err := service.ExecuteScanByTarget(created.ID)
	if err != nil {
		t.Fatalf("ExecuteScanByTarget failed: %v", err)
	}

	if detail.Summary.TargetID != created.ID {
		t.Errorf("Expected scan summary TargetID '%s', got '%s'", created.ID, detail.Summary.TargetID)
	}

	// 5. Get Scans by Target
	scans, err := service.GetScansByTarget(created.ID)
	if err != nil {
		t.Fatalf("GetScansByTarget failed: %v", err)
	}
	if len(scans) == 0 {
		t.Errorf("Expected scans for target %s, got 0", created.ID)
	}
}
