package scanner

import (
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

// ScanKMSKeys attempts live discovery or falls back to static IaC scanning (.tf, .yaml)
func (ks *AWSKMSScanner) ScanKMSKeys(targetDir string, scanID string) ([]models.CryptoAsset, error) {
	// Static IaC file scanner (Terraform, CloudFormation, AWS Config)
	assets := ks.scanIaCFiles(targetDir, scanID)
	return assets, nil
}

var (
	tfKMSKeyRegex   = regexp.MustCompile(`resource\s+["']aws_kms_key["']\s+["']([^"']+)["']`)
	tfAlgoSpecRegex = regexp.MustCompile(`customer_master_key_spec\s*=\s*["']([^"']+)["']`)
	tfRotEnableRegex = regexp.MustCompile(`enable_key_rotation\s*=\s*(true|false)`)
	cfnKMSKeyRegex  = regexp.MustCompile(`Type:\s*['"]?AWS::KMS::Key['"]?`)
)

func (ks *AWSKMSScanner) scanIaCFiles(targetDir string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset

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

		// Match Terraform aws_kms_key declarations block-by-block
		blocks := strings.Split(content, "resource \"aws_kms_key\"")
		if len(blocks) > 1 {
			for _, block := range blocks[1:] {
				keyName := "unnamed"
				if nameMatch := regexp.MustCompile(`^\s*["']([^"']+)["']`).FindStringSubmatch(block); len(nameMatch) > 1 {
					keyName = nameMatch[1]
				} else {
					continue
				}

				algo := "SYMMETRIC_DEFAULT"
				keySize := 256
				riskLevel := models.RiskSafe
				pqcStatus := models.PQCStatusSafe
				pqcReplacement := "None Required (AES-256 KMS Key is PQC Safe)"
				rationale := "AWS KMS Symmetric Default Key (AES-256) provides 128-bit quantum security against Grover's algorithm."

				if specMatch := tfAlgoSpecRegex.FindStringSubmatch(block); len(specMatch) > 1 {
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

				assets = append(assets, models.CryptoAsset{
					ID:             models.GenerateUUID(),
					ScanID:         scanID,
					Name:           fmt.Sprintf("AWS KMS Key: %s", keyName),
					AssetType:      models.AssetKeyStore,
					Location:       relPath,
					LineNumber:     1,
					Algorithm:      algo,
					KeySize:        keySize,
					RiskLevel:      riskLevel,
					RiskRationale:  rationale,
					PQCStatus:      pqcStatus,
					PQCReplacement: pqcReplacement,
					Details:        fmt.Sprintf("Infrastructure-as-Code Terraform AWS KMS Key Resource: %s", keyName),
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

	return assets
}
