package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAWSKMSScannerIaCFallback(t *testing.T) {
	kms := NewAWSKMSScanner("us-east-1")

	tempDir, err := os.MkdirTemp("", "kms_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test Terraform file with AWS KMS key resources
	tfContent := `
	resource "aws_kms_key" "payment_key" {
		description             = "KMS Key for Payment Encryption"
		deletion_window_in_days = 10
		enable_key_rotation     = true
		customer_master_key_spec = "RSA_2048"
	}

	resource "aws_kms_key" "default_key" {
		description = "Default AES Key"
	}
	`

	tfFile := filepath.Join(tempDir, "main.tf")
	if err := os.WriteFile(tfFile, []byte(tfContent), 0644); err != nil {
		t.Fatalf("Failed to write main.tf: %v", err)
	}

	assets, err := kms.ScanKMSKeys(tempDir, "test-kms-scan")
	if err != nil {
		t.Fatalf("ScanKMSKeys failed: %v", err)
	}

	if len(assets) == 0 {
		t.Fatalf("Expected KMS assets to be discovered from main.tf, got 0")
	}

	foundRSA := false
	foundDefault := false

	for _, a := range assets {
		if a.Algorithm == "RSA_2048" {
			foundRSA = true
		}
		if a.Algorithm == "SYMMETRIC_DEFAULT" {
			foundDefault = true
		}
	}

	if !foundRSA {
		t.Errorf("KMS Scanner failed to discover RSA_2048 key spec in Terraform resource")
	}
	if !foundDefault {
		t.Errorf("KMS Scanner failed to discover default SYMMETRIC_DEFAULT key in Terraform resource")
	}
}
