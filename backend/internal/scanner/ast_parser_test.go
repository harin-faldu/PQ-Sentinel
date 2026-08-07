package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestASTParserCommentFiltering(t *testing.T) {
	s := NewScanner()

	tempDir, err := os.MkdirTemp("", "ast_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file with both active code and commented-out code
	testFile := filepath.Join(tempDir, "sample_ast.js")
	codeContent := `
	// const badKey = crypto.generateKeyPairSync('rsa');
	/* const oldJwt = jsonwebtoken.sign(payload, secret, { algorithm: 'RS256' }); */
	const activeHash = crypto.createHash('md5');
	`
	if err := os.WriteFile(testFile, []byte(codeContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	assets, _, _, err := s.ScanDirectory(tempDir, "ast-test-scan")
	if err != nil {
		t.Fatalf("ScanDirectory failed: %v", err)
	}

	// Active code should match (MD5), while commented lines must be ignored!
	for _, asset := range assets {
		if asset.Algorithm == "RSA" || asset.Algorithm == "RS256" {
			t.Errorf("AST Parser failed: Commented out code was wrongly flagged as asset: %s", asset.Name)
		}
	}

	if len(assets) != 1 || assets[0].Algorithm != "MD5" {
		t.Errorf("Expected 1 active MD5 asset, got %d assets", len(assets))
	}
}
