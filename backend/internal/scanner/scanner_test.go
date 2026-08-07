package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanner(t *testing.T) {
	s := NewScanner()

	// Create temp test directory
	tempDir, err := os.MkdirTemp("", "cbom_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create dummy test file with crypto patterns
	testFile := filepath.Join(tempDir, "sample.js")
	codeContent := `
	const crypto = require('crypto');
	const key = crypto.generateKeyPairSync('rsa', { modulusLength: 2048 });
	const jwt = jsonwebtoken.sign(payload, secret, { algorithm: 'RS256' });
	`
	if err := os.WriteFile(testFile, []byte(codeContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	assets, totalFiles, scannedFiles, err := s.ScanDirectory(tempDir, "test-scan-id")
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if totalFiles < 1 || scannedFiles < 1 {
		t.Errorf("Expected scanned files >= 1, got total=%d scanned=%d", totalFiles, scannedFiles)
	}

	if len(assets) == 0 {
		t.Errorf("Expected discovered crypto assets > 0, got 0")
	}
}
