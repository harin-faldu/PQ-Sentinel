package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type AWSKMSScanner struct {
	region string
}

func NewAWSKMSScanner(region string) *AWSKMSScanner {
	if region == "" {
		region = "us-east-1"
	}
	return &AWSKMSScanner{region: region}
}

var (
	tfAlgoSpecRegex = regexp.MustCompile(`customer_master_key_spec\s*=\s*["']([^"']+)["']`)
	cfnKMSKeyRegex  = regexp.MustCompile(`Type:\s*['"]?AWS::KMS::Key['"]?`)
)

func hasAWSCredentials() bool {
	if os.Getenv("AWS_ACCESS_KEY_ID") != "" || os.Getenv("AWS_PROFILE") != "" || os.Getenv("AWS_WEB_IDENTITY_TOKEN_FILE") != "" {
		return true
	}
	homeDir, err := os.UserHomeDir()
	if err == nil {
		credPath := filepath.Join(homeDir, ".aws", "credentials")
		if _, err := os.Stat(credPath); err == nil {
			return true
		}
	}
	return false
}

// ScanKMSKeys runs both IaC and live discovery, then performs drift detection.
// Falls back to IaC-only if no AWS credentials are available.
func (ks *AWSKMSScanner) ScanKMSKeys(targetDir string, scanID string) ([]models.CryptoAsset, error) {
	var assets []models.CryptoAsset

	// IaC scan always runs — cheap, and needed for drift detection either way.
	iacAssets, declaredKeys := ks.scanIaCFilesEnhanced(targetDir, scanID)

	// Instant check: if no AWS credentials are found on the system, skip live network call
	// to avoid 70s IMDS metadata timeout.
	if !hasAWSCredentials() {
		return iacAssets, nil
	}

	// Live scan only runs if credentials are actually available
	ctx := context.Background()
	liveFindings, err := ScanLiveAWSKMS(ctx)
	if err != nil {
		// No credentials or totally unreachable — fall back to IaC-only,
		// same behavior as the scanner had before this change.
		return iacAssets, nil
	}

	if len(liveFindings) == 0 {
		// AWS creds exist but no keys found — still return IaC results
		return iacAssets, nil
	}

	drift := DetectKMSDrift(liveFindings, declaredKeys)

	// Convert live findings to assets
	for _, f := range liveFindings {
		if f.AccessDenied && f.KeyID == "" {
			// Region-level access denial — record as info
			assets = append(assets, models.CryptoAsset{
				ID:            models.GenerateUUID(),
				ScanID:        scanID,
				Name:          fmt.Sprintf("AWS KMS Region: %s (Access Denied)", f.Region),
				AssetType:     models.AssetKeyStore,
				Location:      fmt.Sprintf("aws-kms://%s", f.Region),
				Algorithm:     "UNKNOWN",
				RiskLevel:     models.RiskMedium,
				RiskRationale: fmt.Sprintf("Cannot enumerate KMS keys in region %s: %s", f.Region, f.AccessDeniedMsg),
				PQCStatus:     models.PQCStatusVulnerable,
				Details:       "Live AWS KMS Discovery — region-level access denied. Check IAM permissions.",
			})
			continue
		}
		assets = append(assets, liveKMSFindingToAsset(f, scanID))
	}

	// IaC-only drift findings
	for _, d := range drift.IaCOnlyKeys {
		assets = append(assets, iacOnlyKeyToAsset(d, scanID))
	}

	// Add drift metadata to matching live keys
	for i := range assets {
		for _, alias := range drift.MatchedAliases {
			if strings.Contains(assets[i].Name, alias) {
				assets[i].Details += " [DRIFT: Matched in IaC]"
			}
		}
	}

	return assets, nil
}

func liveKMSFindingToAsset(f LiveKMSFinding, scanID string) models.CryptoAsset {
	algo := "SYMMETRIC_DEFAULT"
	keySize := 256
	riskLevel := models.RiskSafe
	pqcStatus := models.PQCStatusSafe
	pqcReplacement := "None Required (AES-256 KMS Key is PQC Safe)"
	rationale := "Live AWS KMS Symmetric Default Key (AES-256) provides 128-bit quantum security."

	if f.KeySpec != "" {
		algo = f.KeySpec
	}

	if strings.Contains(algo, "RSA") {
		keySize = 2048
		if strings.Contains(algo, "4096") {
			keySize = 4096
		}
		riskLevel = models.RiskHigh
		pqcStatus = models.PQCStatusVulnerable
		pqcReplacement = "AWS KMS Post-Quantum Hybrid Key or ML-KEM-768"
		rationale = fmt.Sprintf("Live AWS KMS RSA Key (%s) vulnerable to Shor's algorithm.", algo)
	} else if strings.Contains(algo, "ECC") || strings.Contains(algo, "ECDSA") {
		keySize = 256
		riskLevel = models.RiskHigh
		pqcStatus = models.PQCStatusVulnerable
		pqcReplacement = "AWS KMS Hybrid PQC Signature"
		rationale = fmt.Sprintf("Live AWS KMS Elliptic Curve Key (%s) vulnerable to Shor's algorithm.", algo)
	}

	name := fmt.Sprintf("AWS KMS Live Key: %s", f.KeyID[:8])
	if f.Alias != "" {
		name = fmt.Sprintf("AWS KMS Live Key: %s", f.Alias)
	}

	details := fmt.Sprintf("Region: %s | KeySpec: %s | Usage: %s | Enabled: %v", f.Region, f.KeySpec, f.KeyUsage, f.Enabled)
	if f.RotationUnknown {
		details += " | Rotation: UNKNOWN (access denied)"
	} else if f.RotationEnabled {
		details += " | Rotation: ENABLED"
	} else {
		details += " | Rotation: DISABLED"
	}

	return models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           name,
		AssetType:      models.AssetKeyStore,
		Location:       fmt.Sprintf("aws-kms://%s/%s", f.Region, f.KeyID),
		Algorithm:      algo,
		KeySize:        keySize,
		RiskLevel:      riskLevel,
		RiskRationale:  rationale,
		PQCStatus:      pqcStatus,
		PQCReplacement: pqcReplacement,
		Details:        details,
	}
}

func iacOnlyKeyToAsset(d IaCDeclaredKey, scanID string) models.CryptoAsset {
	name := fmt.Sprintf("AWS KMS IaC-Only Key: %s", d.LocalName)
	if d.Alias != "" {
		name = fmt.Sprintf("AWS KMS IaC-Only Key: %s (%s)", d.LocalName, d.Alias)
	}
	details := fmt.Sprintf("Drift: Declared in Terraform but NOT found in live AWS account. LocalName=%s", d.LocalName)
	if d.RotationStated {
		details += fmt.Sprintf(" | Rotation=%v", d.RotationEnabled)
	}

	return models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           name,
		AssetType:      models.AssetKeyStore,
		Location:       fmt.Sprintf("terraform/aws_kms_key.%s", d.LocalName),
		Algorithm:      "SYMMETRIC_DEFAULT",
		KeySize:        256,
		RiskLevel:      models.RiskMedium,
		RiskRationale:  "IaC-only drift: KMS key declared in Terraform but not deployed. May indicate pending apply or stale config.",
		PQCStatus:      models.PQCStatusSafe,
		PQCReplacement: "None Required (AES-256 is PQC Safe)",
		Details:        details,
	}
}

// scanIaCFilesEnhanced scans IaC files and returns both CryptoAssets and parsed IaCDeclaredKeys for drift detection
func (ks *AWSKMSScanner) scanIaCFilesEnhanced(targetDir string, scanID string) ([]models.CryptoAsset, []IaCDeclaredKey) {
	var assets []models.CryptoAsset
	var allDeclaredKeys []IaCDeclaredKey

	_ = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".tf" && ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := string(contentBytes)
		relPath := path
		if rel, err := filepath.Rel(targetDir, path); err == nil {
			relPath = rel
		}

		// Enhanced Terraform parsing with alias resolution
		if ext == ".tf" {
			declared := parseTerraformKMSBlocks(content)
			allDeclaredKeys = append(allDeclaredKeys, declared...)

			for _, dk := range declared {
				algo := "SYMMETRIC_DEFAULT"
				keySize := 256
				riskLevel := models.RiskSafe
				pqcStatus := models.PQCStatusSafe
				pqcReplacement := "None Required (AES-256 KMS Key is PQC Safe)"
				rationale := "AWS KMS Symmetric Default Key (AES-256) provides 128-bit quantum security against Grover's algorithm."

				// Check if block content has algo spec overrides
				keyBlocks := splitIntoBlocks(content, tfKMSKeyBlockRegex)
				if blockBody, ok := keyBlocks[dk.LocalName]; ok {
					if specMatch := tfAlgoSpecRegex.FindStringSubmatch(blockBody); len(specMatch) > 1 {
						spec := specMatch[1]
						if strings.Contains(spec, "RSA_2048") || strings.Contains(spec, "RSA_1024") {
							algo = "RSA_2048"
							keySize = 2048
							riskLevel = models.RiskHigh
							pqcStatus = models.PQCStatusVulnerable
							pqcReplacement = "AWS KMS Post-Quantum Hybrid Key or ML-KEM-768"
							rationale = "AWS KMS RSA Key Spec is vulnerable to quantum Shor's algorithm."
						} else if strings.Contains(spec, "ECC_NIST_P256") || strings.Contains(spec, "ECDSA") {
							algo = "ECC_NIST_P256"
							keySize = 256
							riskLevel = models.RiskHigh
							pqcStatus = models.PQCStatusVulnerable
							pqcReplacement = "AWS KMS Hybrid PQC Signature"
							rationale = "AWS KMS Elliptic Curve Key Spec is vulnerable to quantum Shor's algorithm."
						}
					}
				}

				name := fmt.Sprintf("AWS KMS Key: %s", dk.LocalName)
				if dk.Alias != "" {
					name = fmt.Sprintf("AWS KMS Key: %s (%s)", dk.LocalName, dk.Alias)
				}

				assets = append(assets, models.CryptoAsset{
					ID:             models.GenerateUUID(),
					ScanID:         scanID,
					Name:           name,
					AssetType:      models.AssetKeyStore,
					Location:       relPath,
					LineNumber:     1,
					Algorithm:      algo,
					KeySize:        keySize,
					RiskLevel:      riskLevel,
					RiskRationale:  rationale,
					PQCStatus:      pqcStatus,
					PQCReplacement: pqcReplacement,
					Details:        fmt.Sprintf("Infrastructure-as-Code Terraform AWS KMS Key Resource: %s", dk.LocalName),
				})
			}
		}

		// Match CloudFormation AWS::KMS::Key
		if cfnKMSKeyRegex.MatchString(content) {
			assets = append(assets, models.CryptoAsset{
				ID:             models.GenerateUUID(),
				ScanID:         scanID,
				Name:           "CloudFormation AWS KMS Key",
				AssetType:      models.AssetKeyStore,
				Location:       relPath,
				LineNumber:     1,
				Algorithm:      "SYMMETRIC_DEFAULT",
				KeySize:        256,
				RiskLevel:      models.RiskSafe,
				RiskRationale:  "AWS KMS Key declared in CloudFormation template.",
				PQCStatus:      models.PQCStatusSafe,
				PQCReplacement: "None Required (AES-256 is PQC Safe)",
				Details:        "CloudFormation template resource: AWS::KMS::Key",
			})
		}

		return nil
	})

	return assets, allDeclaredKeys
}
