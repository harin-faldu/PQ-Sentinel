package scanner

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
	"github.com/spf13/viper"
	"github.com/zricethezav/gitleaks/v8/config"
	"github.com/zricethezav/gitleaks/v8/detect"
	"github.com/zricethezav/gitleaks/v8/sources"
)

type SecretScanner struct {
	rules    *RulesConfig
	detector *detect.Detector
}

func NewSecretScanner(rules *RulesConfig) *SecretScanner {
	v := viper.New()
	v.SetConfigType("toml")

	var cfg config.Config
	if err := v.ReadConfig(strings.NewReader(config.DefaultConfig)); err == nil {
		var vc config.ViperConfig
		if err := v.Unmarshal(&vc); err == nil {
			if translated, err := vc.Translate(); err == nil {
				cfg = translated
			}
		}
	}

	if cfg.Rules == nil {
		cfg.Rules = make(map[string]config.Rule)
	}

	// Add custom secret rules from external rules.yaml into Gitleaks engine
	if rules != nil {
		for _, sr := range rules.SecretRules {
			if sr.CompiledRegex != nil {
				cfg.Rules[sr.ID] = config.Rule{
					Description: sr.Name,
					RuleID:      sr.ID,
					Regex:       sr.CompiledRegex,
				}
			}
		}
	}

	detector := detect.NewDetector(cfg)
	return &SecretScanner{
		rules:    rules,
		detector: detector,
	}
}

// RedactSecret masks sensitive characters in a secret match string for security
func RedactSecret(secret string) string {
	secret = strings.TrimSpace(secret)
	if len(secret) <= 6 {
		return "******"
	}
	prefixLen := 4
	if len(secret) < 10 {
		prefixLen = 2
	}
	return secret[:prefixLen] + strings.Repeat("*", len(secret)-prefixLen)
}

// ScanSecrets performs secret detection using embedded Gitleaks Go engine + Entropy Analysis + Git History Scanner
func (ss *SecretScanner) ScanSecrets(targetDir string) ([]models.CryptoAsset, error) {
	var assets []models.CryptoAsset

	// 1. Scan Git Commit History (.git) if repository exists (with full patch diff -p)
	gitDir := filepath.Join(targetDir, ".git")
	if _, err := os.Stat(gitDir); err == nil && ss.detector != nil {
		gitCmd, err := sources.NewGitLogCmd(targetDir, "--all -p")
		if err != nil {
			gitCmd, err = sources.NewGitLogCmd(targetDir, "--all")
		}
		if err == nil {
			// Pass non-nil *detect.RemoteInfo to prevent nil-pointer dereference in Gitleaks createScmLink
			remoteInfo := &detect.RemoteInfo{}
			findings, err := ss.detector.DetectGit(gitCmd, remoteInfo)
			if err == nil {
				for _, finding := range findings {
					redactedMatch := RedactSecret(finding.Match)
					locationInfo := finding.File
					if finding.Commit != "" {
						commitHash := finding.Commit
						if len(commitHash) >= 8 {
							commitHash = commitHash[:8]
						}
						locationInfo = fmt.Sprintf("%s (Git Commit: %s)", finding.File, commitHash)
					}

					assets = append(assets, models.CryptoAsset{
						ID:             models.GenerateUUID(),
						Name:           "Git History Secret: " + finding.Description,
						Type:           models.AssetKeyMaterial,
						AssetType:      models.AssetKeyMaterial,
						Algorithm:      finding.RuleID,
						KeySize:        len(finding.Secret) * 8,
						Location:       locationInfo,
						LineNumber:     finding.StartLine,
						ContextSnippet: "Historical Commit Secret: " + redactedMatch,
						RiskLevel:      models.RiskCritical,
						PQCStatus:      models.StatusVulnerable,
						PQCReplacement: "Rotate Secret & Purge Git Commit History (git filter-repo / BFG)",
						RiskRationale:  "Exposed secret found in historical Git commit log. Still vulnerable even if deleted from latest code!",
						DiscoveredAt:   time.Now(),
					})
				}
			}
		}
	}

	// 2. Scan Filesystem & Dotenv (.env, .env.local, .env.production)
	err := filepathWalk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Ignore framework dependency, build output, cache, and IDE directories
		// Use path separator awareness to avoid false positives (e.g. "rebuild" matching "build")
		lowerPath := strings.ToLower(path)
		sep := string(filepath.Separator)
		ignoreDirs := []string{
			// VCS
			".git" + sep, ".svn" + sep, ".hg" + sep,
			// JS/TS/Node
			"node_modules" + sep, ".next" + sep, ".nuxt" + sep, ".turbo" + sep,
			".parcel-cache" + sep, "bower_components" + sep,
			// Go
			"vendor" + sep,
			// Java/Kotlin/Scala/Android
			".gradle" + sep, ".m2" + sep, ".mvn" + sep, ".idea" + sep,
			sep + "target" + sep,
			// .NET/C#
			"packages" + sep, sep + "obj" + sep,
			// Python
			"__pycache__" + sep, "venv" + sep, ".venv" + sep, ".tox" + sep,
			".mypy_cache" + sep, ".pytest_cache" + sep, ".eggs" + sep,
			// Rust
			".cargo" + sep,
			// Ruby
			".bundle" + sep,
			// Generic build/dist/output
			sep + "bin" + sep, sep + "dist" + sep, sep + "build" + sep,
			sep + "out" + sep, sep + "output" + sep, sep + "release" + sep,
			sep + "debug" + sep, ".cache" + sep, sep + "tmp" + sep,
			sep + "coverage" + sep, ".nyc_output" + sep, "htmlcov" + sep,
			sep + "logs" + sep,
			// IDE
			".vs" + sep, ".vscode" + sep, ".settings" + sep,
			// IaC state
			".terraform" + sep,
		}
		skipFile := false
		for _, dir := range ignoreDirs {
			if strings.Contains(lowerPath, dir) {
				skipFile = true
				break
			}
		}
		if skipFile {
			return nil
		}

		relPath, _ := filepath.Rel(targetDir, path)
		baseName := strings.ToLower(info.Name())
		ext := strings.ToLower(filepath.Ext(path))

		// Ignore dependency checksum lockfiles (go.sum, package-lock.json, yarn.lock, etc.)
		if baseName == "go.sum" || baseName == "package-lock.json" ||
			baseName == "yarn.lock" || baseName == "pnpm-lock.yaml" ||
			baseName == "cargo.lock" || baseName == "composer.lock" ||
			baseName == "poetry.lock" || baseName == "pipfile.lock" {
			return nil
		}

		// Check .env files explicitly
		isDotEnv := strings.HasPrefix(baseName, ".env")

		// 1. Check dangerous file extensions (.pem, .key, .jks, .pfx, .p12)
		if ext == ".pem" || ext == ".key" || ext == ".jks" || ext == ".pfx" || ext == ".p12" {
			assets = append(assets, models.CryptoAsset{
				ID:             models.GenerateUUID(),
				Name:           "Key File: " + filepath.Base(path),
				Type:           models.AssetKeyMaterial,
				AssetType:      models.AssetKeyMaterial,
				Algorithm:      strings.ToUpper(strings.TrimPrefix(ext, ".")),
				KeySize:        int(info.Size() * 8),
				Location:       relPath,
				LineNumber:     1,
				ContextSnippet: "Unencrypted key material stored on filesystem: " + info.Name(),
				RiskLevel:      models.RiskCritical,
				PQCStatus:      models.StatusVulnerable,
				PQCReplacement: "Migrate to HSM / Key Vault",
				RiskRationale:  "Raw key file committed to project directory.",
				DiscoveredAt:   time.Now(),
			})
			return nil
		}

		if isBinaryExt(ext) && !isDotEnv {
			return nil
		}

		// Read file content
		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// 2. Embedded Gitleaks Scan on File Bytes (In-Memory with Secret Redaction)
		if ss.detector != nil {
			fragment := detect.Fragment{
				FilePath: relPath,
				Raw:      string(contentBytes),
			}
			for _, finding := range ss.detector.Detect(fragment) {
				redactedMatch := RedactSecret(finding.Match)
				assets = append(assets, models.CryptoAsset{
					ID:             models.GenerateUUID(),
					Name:           "Secret: " + finding.Description,
					Type:           models.AssetKeyMaterial,
					AssetType:      models.AssetKeyMaterial,
					Algorithm:      finding.RuleID,
					KeySize:        len(finding.Secret) * 8,
					Location:       relPath,
					LineNumber:     finding.StartLine,
					ContextSnippet: "Redacted Match: " + redactedMatch,
					RiskLevel:      models.RiskCritical,
					PQCStatus:      models.StatusVulnerable,
					PQCReplacement: "Rotate Secret & Store in a managed KMS",
					RiskRationale:  "Hardcoded credential exposed in source code / .env file (Gitleaks scan).",
					DiscoveredAt:   time.Now(),
				})
			}
		}

		// 3. Custom YAML Secret Rules & Shannon Entropy Analysis line-by-line
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNo := 0
		for scanner.Scan() {
			lineNo++
			line := scanner.Text()

			// Check YAML secret rules
			if ss.rules != nil {
				for _, rule := range ss.rules.SecretRules {
					if rule.CompiledRegex != nil && rule.CompiledRegex.MatchString(line) {
						redactedLine := RedactSecret(strings.TrimSpace(line))
						assets = append(assets, models.CryptoAsset{
							ID:             models.GenerateUUID(),
							Name:           "Secret: " + rule.Name,
							Type:           models.AssetKeyMaterial,
							AssetType:      models.AssetKeyMaterial,
							Algorithm:      rule.ID,
							KeySize:        256,
							Location:       relPath,
							LineNumber:     lineNo,
							ContextSnippet: "Redacted Secret Line: " + redactedLine,
							RiskLevel:      models.RiskLevel(rule.RiskLevel),
							PQCStatus:      models.StatusVulnerable,
							PQCReplacement: "Rotate credential & use a secrets manager",
							RiskRationale:  "Hardcoded secret detected matching pattern " + rule.Name,
							DiscoveredAt:   time.Now(),
						})
					}
				}
			}

			// High Shannon Entropy String check (> 4.6 bits/char on secret token >= 32 chars)
			// Extract pure secret tokens (alphanumeric with _, -, =) to avoid code syntax false positives
			candidateRegex := regexp.MustCompile(`[a-zA-Z0-9_\-=]{32,}`)
			for _, token := range candidateRegex.FindAllString(line, -1) {
				// Skip URLs, git remote configs, common imports/keywords, XML/ResX schemas, base64 data, or PublicKeyToken identifiers
				if strings.HasPrefix(token, "http") || strings.HasPrefix(token, "https") ||
					strings.Contains(token, "import") || strings.Contains(line, "url =") ||
					strings.Contains(line, "fetch =") || strings.HasSuffix(relPath, "config") ||
					ext == ".resx" || ext == ".resources" ||
					strings.Contains(line, "PublicKeyToken") || strings.Contains(line, "mimetype=") ||
					strings.Contains(line, "base64") || strings.Contains(line, "ResXResource") {
					continue
				}

				if calculateEntropy(token) > 4.6 {
					redactedToken := RedactSecret(token)
					assets = append(assets, models.CryptoAsset{
						ID:             models.GenerateUUID(),
						Name:           "High-Entropy Secret String",
						Type:           models.AssetKeyMaterial,
						AssetType:      models.AssetKeyMaterial,
						Algorithm:      "HighEntropyToken",
						KeySize:        len(token) * 8,
						Location:       relPath,
						LineNumber:     lineNo,
						ContextSnippet: "High entropy token: " + redactedToken,
						RiskLevel:      models.RiskHigh,
						PQCStatus:      models.StatusVulnerable,
						PQCReplacement: "Environment variable / Vault secret injection",
						RiskRationale:  "High Shannon entropy string detected (potential API key or secret token).",
						Confidence:     0.7,
						DiscoveredAt:   time.Now(),
					})
				}
			}
		}

		return nil
	})

	return assets, err
}

func filepathWalk(targetDir string, fn filepath.WalkFunc) error {
	return filepath.Walk(targetDir, fn)
}

func calculateEntropy(s string) float64 {
	if len(s) == 0 {
		return 0
	}
	freq := make(map[rune]float64)
	for _, char := range s {
		freq[char]++
	}

	var entropy float64
	length := float64(len(s))
	for _, count := range freq {
		p := count / length
		entropy -= p * math.Log2(p)
	}
	return entropy
}

func isBinaryExt(ext string) bool {
	binaries := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
		".exe": true, ".dll": true, ".so": true, ".dylib": true, ".zip": true,
		".tar": true, ".gz": true, ".pdf": true, ".woff": true, ".ttf": true,
		".pyc": true, ".pyo": true, ".class": true, ".o": true, ".obj": true,
		".sqlite": true, ".db": true, ".wasm": true, ".bin": true,
		".resx": true, ".resources": true, ".nupkg": true, ".p7s": true,
		".pack": true, ".idx": true, ".rev": true, ".sample": true,
	}
	return binaries[ext]
}
