package scanner

import (
	"testing"
)

// TestParseTerraformKMSBlocks verifies the enhanced IaC parser correctly
// correlates aws_kms_key and aws_kms_alias blocks through target_key_id references
func TestParseTerraformKMSBlocks(t *testing.T) {
	tfContent := `
resource "aws_kms_key" "payment_key" {
  description             = "KMS key for encrypting payment data"
  enable_key_rotation     = true
  deletion_window_in_days = 30
  customer_master_key_spec = "SYMMETRIC_DEFAULT"
}

resource "aws_kms_alias" "payment_alias" {
  name          = "alias/payment-key"
  target_key_id = aws_kms_key.payment_key.id
}

resource "aws_kms_key" "signing_key" {
  description             = "Signing key for document integrity"
  enable_key_rotation     = false
  customer_master_key_spec = "RSA_2048"
}

resource "aws_kms_alias" "signing_alias" {
  name          = "alias/signing-key"
  target_key_id = aws_kms_key.signing_key.arn
}

resource "aws_kms_key" "orphan_key" {
  description         = "Key with no alias"
  enable_key_rotation = true
}
`

	declared := parseTerraformKMSBlocks(tfContent)

	if len(declared) != 3 {
		t.Fatalf("Expected 3 declared keys, got %d", len(declared))
	}

	byName := make(map[string]IaCDeclaredKey)
	for _, d := range declared {
		byName[d.LocalName] = d
	}

	// Test payment_key: rotation=true, alias resolved
	pk, ok := byName["payment_key"]
	if !ok {
		t.Fatal("payment_key not found in parsed results")
	}
	if !pk.RotationStated || !pk.RotationEnabled {
		t.Errorf("payment_key: expected rotation=true, got stated=%v enabled=%v", pk.RotationStated, pk.RotationEnabled)
	}
	if pk.Alias != "alias/payment-key" {
		t.Errorf("payment_key: expected alias='alias/payment-key', got '%s'", pk.Alias)
	}

	// Test signing_key: rotation=false, alias resolved via .arn reference
	sk, ok := byName["signing_key"]
	if !ok {
		t.Fatal("signing_key not found in parsed results")
	}
	if !sk.RotationStated || sk.RotationEnabled {
		t.Errorf("signing_key: expected rotation=false, got stated=%v enabled=%v", sk.RotationStated, sk.RotationEnabled)
	}
	if sk.Alias != "alias/signing-key" {
		t.Errorf("signing_key: expected alias='alias/signing-key', got '%s'", sk.Alias)
	}

	// Test orphan_key: has rotation but no alias
	orphan, ok := byName["orphan_key"]
	if !ok {
		t.Fatal("orphan_key not found in parsed results")
	}
	if !orphan.RotationStated || !orphan.RotationEnabled {
		t.Errorf("orphan_key: expected rotation=true, got stated=%v enabled=%v", orphan.RotationStated, orphan.RotationEnabled)
	}
	if orphan.Alias != "" {
		t.Errorf("orphan_key: expected no alias, got '%s'", orphan.Alias)
	}
}

// TestDetectKMSDrift validates the drift detection engine sorts matched,
// console-only, and IaC-only keys correctly
func TestDetectKMSDrift(t *testing.T) {
	liveFindings := []LiveKMSFinding{
		{Region: "us-east-1", KeyID: "key-111", Alias: "alias/payment-key", Enabled: true, KeySpec: "SYMMETRIC_DEFAULT"},
		{Region: "us-east-1", KeyID: "key-222", Alias: "alias/console-only", Enabled: true, KeySpec: "SYMMETRIC_DEFAULT"},
		{Region: "eu-west-1", KeyID: "key-333", AccessDenied: true, AccessDeniedMsg: "DescribeKey: access denied"},
	}

	declaredKeys := []IaCDeclaredKey{
		{LocalName: "payment_key", Alias: "alias/payment-key", RotationEnabled: true, RotationStated: true},
		{LocalName: "future_key", Alias: "alias/future-key", RotationEnabled: false, RotationStated: true},
	}

	drift := DetectKMSDrift(liveFindings, declaredKeys)

	// Matched aliases: only payment-key should match
	if len(drift.MatchedAliases) != 1 || drift.MatchedAliases[0] != "alias/payment-key" {
		t.Errorf("Expected 1 matched alias 'alias/payment-key', got %v", drift.MatchedAliases)
	}

	// Console-only: key-222 (alias/console-only) exists in live but not in IaC
	if len(drift.ConsoleOnlyKeys) != 1 || drift.ConsoleOnlyKeys[0].Alias != "alias/console-only" {
		t.Errorf("Expected 1 console-only key with alias 'alias/console-only', got %v", drift.ConsoleOnlyKeys)
	}

	// IaC-only: future_key (alias/future-key) declared in IaC but not in live
	if len(drift.IaCOnlyKeys) != 1 || drift.IaCOnlyKeys[0].LocalName != "future_key" {
		t.Errorf("Expected 1 IaC-only key 'future_key', got %v", drift.IaCOnlyKeys)
	}
}
