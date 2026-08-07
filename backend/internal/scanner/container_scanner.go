package scanner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type ContainerScanner struct{}

func NewContainerScanner() *ContainerScanner {
	return &ContainerScanner{}
}

var safeImageRegex = regexp.MustCompile(`^[a-zA-Z0-9_\.\/\:\-\\\s]+$`)

// ScanContainerImage scans a container image via Syft CLI or Dockerfile fallback
func (cs *ContainerScanner) ScanContainerImage(imageName string, scanID string) ([]models.CryptoAsset, error) {
	// Shell injection prevention guard: enforce valid container image name pattern
	if !safeImageRegex.MatchString(imageName) {
		return nil, fmt.Errorf("invalid container image name: unsafe characters detected")
	}

	// Check if syft binary exists on PATH
	syftPath, err := exec.LookPath("syft")
	if err == nil {
		assets, err := cs.scanWithSyft(syftPath, imageName, scanID)
		if err == nil && len(assets) > 0 {
			return assets, nil
		}
	}

	// Fallback to static Dockerfile & Compose analyzer
	return cs.scanDockerfileFallback(imageName, scanID), nil
}

type syftOutput struct {
	Artifacts []struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Type     string `json:"type"`
		Language string `json:"language"`
	} `json:"artifacts"`
}

func (cs *ContainerScanner) scanWithSyft(syftPath, imageName, scanID string) ([]models.CryptoAsset, error) {
	cmd := exec.Command(syftPath, imageName, "-o", "json")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("syft scan execution failed: %w", err)
	}

	return cs.ParseSyftJSON(out.Bytes(), imageName, scanID)
}

func (cs *ContainerScanner) ParseSyftJSON(data []byte, imageName, scanID string) ([]models.CryptoAsset, error) {
	var syftRes syftOutput
	if err := json.Unmarshal(data, &syftRes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal syft json: %w", err)
	}

	var assets []models.CryptoAsset
	cryptoPkgs := map[string]string{
		"openssl":   "OpenSSL Cryptography Library",
		"libssl":    "OpenSSL LibSSL Shared Library",
		"gnutls":    "GnuTLS Cryptography Library",
		"mbedtls":   "mbedTLS Embedded Crypto Library",
		"libgcrypt": "Libgcrypt Crypto Library",
		"crypto":    "Go Crypto Package",
	}

	for _, art := range syftRes.Artifacts {
		artNameLower := strings.ToLower(art.Name)
		for key, label := range cryptoPkgs {
			if strings.Contains(artNameLower, key) {
				riskLevel := models.RiskSafe
				pqcStatus := models.PQCStatusSafe
				rationale := fmt.Sprintf("Container package %s (%s) installed.", art.Name, art.Version)
				pqcReplacement := "None Required (Modern Package)"

				if strings.Contains(art.Version, "1.1.1") || strings.Contains(art.Version, "1.0.") {
					riskLevel = models.RiskCritical
					pqcStatus = models.PQCStatusVulnerable
					rationale = fmt.Sprintf("EOL package %s (%s) lacks PQC algorithm provider support.", art.Name, art.Version)
					pqcReplacement = "Upgrade to OpenSSL 3.2+ / oqs-provider"
				}

				assets = append(assets, models.CryptoAsset{
					ID:             models.GenerateUUID(),
					ScanID:         scanID,
					Name:           fmt.Sprintf("Container Package: %s", label),
					AssetType:      models.AssetLibrary,
					Location:       fmt.Sprintf("Image: %s", imageName),
					LineNumber:     1,
					Algorithm:      art.Name,
					KeySize:        0,
					RiskLevel:      riskLevel,
					RiskRationale:  rationale,
					PQCStatus:      pqcStatus,
					PQCReplacement: pqcReplacement,
					Details:        fmt.Sprintf("Syft Discovered Package: %s@%s (Type: %s)", art.Name, art.Version, art.Type),
				})
				break
			}
		}
	}

	return assets, nil
}

func (cs *ContainerScanner) scanDockerfileFallback(targetDirOrImage string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset

	// Check if target directory contains a Dockerfile
	_ = filepath.Walk(targetDirOrImage, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		baseName := strings.ToLower(info.Name())
		if baseName == "dockerfile" || strings.HasPrefix(baseName, "dockerfile.") {
			contentBytes, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			content := string(contentBytes)
			relPath := path
			if rel, err := filepath.Rel(targetDirOrImage, path); err == nil {
				relPath = rel
			}

			lines := strings.Split(content, "\n")
			for i, line := range lines {
				trimmed := strings.TrimSpace(line)
				if strings.HasPrefix(trimmed, "FROM ") {
					baseImage := strings.TrimPrefix(trimmed, "FROM ")
					assets = append(assets, models.CryptoAsset{
						ID:             models.GenerateUUID(),
						ScanID:         scanID,
						Name:           fmt.Sprintf("Dockerfile Base Image (%s)", baseImage),
						AssetType:      models.AssetLibrary,
						Location:       relPath,
						LineNumber:     i + 1,
						Algorithm:      "BaseImage",
						KeySize:        0,
						RiskLevel:      models.RiskSafe,
						RiskRationale:  fmt.Sprintf("Dockerfile uses base image: %s", baseImage),
						PQCStatus:      models.PQCStatusSafe,
						PQCReplacement: "Ensure base image is updated with PQC-ready libraries",
						Details:        trimmed,
					})
				} else if strings.Contains(strings.ToLower(trimmed), "openssl") || strings.Contains(strings.ToLower(trimmed), "libssl") {
					assets = append(assets, models.CryptoAsset{
						ID:             models.GenerateUUID(),
						ScanID:         scanID,
						Name:           "Dockerfile Crypto Package Installation",
						AssetType:      models.AssetLibrary,
						Location:       relPath,
						LineNumber:     i + 1,
						Algorithm:      "OpenSSL",
						KeySize:        0,
						RiskLevel:      models.RiskMedium,
						RiskRationale:  "Package installation command references OpenSSL / SSL libraries.",
						PQCStatus:      models.PQCStatusTransitional,
						PQCReplacement: "Ensure OpenSSL 3.2+ and oqs-provider are installed",
						Details:        trimmed,
					})
				}
			}
		}
		return nil
	})

	return assets
}
