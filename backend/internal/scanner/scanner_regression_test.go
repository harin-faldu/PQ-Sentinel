package scanner

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/quantumpq/qv-cbom/internal/models"
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

// assertLineHasAsset fails the test if no asset with the given algorithm exists at the given line
func assertLineHasAsset(t *testing.T, allAssets []models.CryptoAsset, filename string, line int, algo string) {
	t.Helper()
	for _, a := range allAssets {
		if a.LineNumber == line && a.Algorithm == algo {
			return
		}
	}
	t.Errorf("Missing asset in %s: expected %s at line %d, but no engine detected it", filename, algo, line)
}

func TestEngineParityAndZeroRegression(t *testing.T) {
	s := NewScanner()

	tempDir, err := os.MkdirTemp("", "parity_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	type testFile struct {
		filename       string
		content        string
		lineAssertions map[int]string // lineNumber -> expectedAlgorithm (must match rule's algorithm field exactly)
	}

	testFiles := []testFile{
		{
			filename: "test_edge_cases.js",
			content: `const crypto = require('crypto');
// Line 2: Standard dot-notation call
const h1 = crypto.createHash('md5');
// Line 4: Chained require() call
const h2 = require('crypto').createHash('sha1');
// Line 6: Bracket notation
const h3 = crypto["createHash"]("md5");
// Line 8: RSA keygen (generateKeyPairSync)
const k1 = crypto.generateKeyPairSync('rsa', { modulusLength: 2048 });
`,
			lineAssertions: map[int]string{
				3: "MD5",   // standard dot-notation → JS-CRYPTO-01
				5: "SHA-1", // chained require() call → JS-CRYPTO-02
				7: "MD5",   // bracket notation → JS-CRYPTO-01
				9: "RSA",   // generateKeyPairSync → JS-CRYPTO-03 (with Sync? fix)
			},
		},
		{
			filename: "test_webcrypto.ts",
			content: `// Line 1: comment
// Line 2: WebCrypto subtle.digest
window.crypto.subtle.digest('SHA-256', data);
`,
			lineAssertions: map[int]string{
				3: "SHA-256", // subtle.digest → JS-CRYPTO-08
			},
		},
		{
			filename: "test_python.py",
			content: `import hashlib
# Line 2: Python hashlib MD5
h1 = hashlib.md5()
# Line 4: Python hashlib SHA1
h2 = hashlib.sha1()
`,
			lineAssertions: map[int]string{
				3: "MD5",   // hashlib.md5 → PY-CRYPTO-01
				5: "SHA-1", // hashlib.sha1 → PY-CRYPTO-02
			},
		},
		{
			filename: "test_go.go",
			content: `package main
import "crypto/rsa"
func main() {
	// Line 4: Go RSA keygen
	key, _ := rsa.GenerateKey(nil, 2048)
}
`,
			lineAssertions: map[int]string{
				5: "RSA", // rsa.GenerateKey → GO-CRYPTO-01
			},
		},
		{
			filename: "Demo.java",
			content: `import java.security.MessageDigest;
import javax.crypto.Cipher;
public class Demo {
	public void test() throws Exception {
		// Line 5 & 6: Java crypto calls
		MessageDigest md = MessageDigest.getInstance("MD5");
		Cipher c = Cipher.getInstance("RSA");
	}
}
`,
			lineAssertions: map[int]string{
				6: "MD5/SHA-1", // JV-CRYPTO-02 algorithm field is "MD5/SHA-1"
				7: "RSA",       // JV-CRYPTO-01
			},
		},
	}

	for _, tf := range testFiles {
		path := filepath.Join(tempDir, tf.filename)
		if err := os.WriteFile(path, []byte(tf.content), 0644); err != nil {
			t.Fatalf("Failed to write test file %s: %v", tf.filename, err)
		}

		sourceBytes := []byte(tf.content)

		// 1. Engine 1: Force Line-Based Regex Engine Fallback
		regexAssets := s.parseCodeFileLineFallback(path, tempDir, "regex-scan", tf.content)

		// 2. Engine 2: Force Tree-Sitter AST Engine directly
		astAssets, treeSitterAvailable := s.treeSitterParser.ParseCodeWithTreeSitter(sourceBytes, tf.filename, "ast-scan")

		if treeSitterAvailable {
			// 3. Count Assertion: Tree-Sitter must find at least as many valid assets as regex engine
			if len(astAssets) < len(regexAssets) {
				t.Errorf("Parity Failure in %s: Tree-Sitter found %d assets, regex engine found %d assets",
					tf.filename, len(astAssets), len(regexAssets))
			}

			// 4. Real 1-to-1 Line-Level Diffing (regex vs AST)
			for _, rAsset := range regexAssets {
				foundInAST := false
				for _, aAsset := range astAssets {
					if rAsset.LineNumber == aAsset.LineNumber && rAsset.Algorithm == aAsset.Algorithm {
						foundInAST = true
						break
					}
				}
				if !foundInAST {
					t.Errorf("Parity Delta in %s: Regex found %s at line %d, Tree-Sitter MISSED it!",
						tf.filename, rAsset.Algorithm, rAsset.LineNumber)
				}
			}
		} else {
			t.Logf("Tree-Sitter not available for %s (non-CGO build), testing regex engine only", tf.filename)
		}

		// 5. Per-Line Edge-Case Assertions (prevents silent partial failures)
		// Combine both engines' outputs — at least one must detect each edge case
		combinedAssets := append(regexAssets, astAssets...)
		for line, algo := range tf.lineAssertions {
			assertLineHasAsset(t, combinedAssets, tf.filename, line, algo)
		}
	}
}
