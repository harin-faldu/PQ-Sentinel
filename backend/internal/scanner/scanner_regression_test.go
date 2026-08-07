package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTreeSitterStringAndCommentFalsePositives(t *testing.T) {
	s := NewScanner()

	tempDir, err := os.MkdirTemp("", "ts_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create JS file with log strings and multi-line block comments
	testFile := filepath.Join(tempDir, "sample_false_positives.js")
	codeContent := `
	// Log message describing function name - must NOT be flagged!
	console.log("We previously used crypto.createHash('md5') for password hashing");
	
	/*
	 Multi-line comment block:
	 const key = crypto.generateKeyPairSync('rsa', { modulusLength: 1024 });
	 const cipher = crypto.createCipheriv('des-cbc', key, iv);
	*/

	const info = "crypto.createCipheriv('des', k, i)";
	`
	if err := os.WriteFile(testFile, []byte(codeContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	assets, _, _, err := s.ScanDirectory(tempDir, "ts-false-positive-test")
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Must be 0 assets! Log strings and comment blocks must never trigger call_expression matches!
	if len(assets) > 0 {
		for _, a := range assets {
			t.Errorf("Tree-Sitter AST failed: String literal or multi-line comment was wrongly flagged as asset: %s at %s:%d", a.Name, a.Location, a.LineNumber)
		}
	}
}

func TestTreeSitterMultilineCallArgumentExtraction(t *testing.T) {
	s := NewScanner()

	tempDir, err := os.MkdirTemp("", "ts_multiline_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Multi-line function call spanning 6 lines
	testFile := filepath.Join(tempDir, "multiline_call.js")
	codeContent := `
	const crypto = require('crypto');
	const key = crypto.generateKeyPairSync(
		'rsa',
		{
			modulusLength: 1024
		}
	);
	`
	if err := os.WriteFile(testFile, []byte(codeContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	assets, _, _, err := s.ScanDirectory(tempDir, "ts-multiline-test")
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	if len(assets) == 0 {
		t.Fatalf("Expected multi-line RSA call to be discovered, got 0 assets")
	}

	foundRSA1024 := false
	for _, a := range assets {
		if a.Algorithm == "RSA" && a.KeySize == 1024 {
			foundRSA1024 = true
			if a.Confidence <= 0 {
				t.Errorf("Expected Confidence > 0, got %.2f", a.Confidence)
			}
		}
	}

	if !foundRSA1024 {
		t.Errorf("Tree-sitter failed to extract multi-line modulusLength: 1024. Asset list: %+v", assets)
	}
}
