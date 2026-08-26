package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type Scanner struct {
	rulesConfig      *RulesConfig
	astParser        *ASTParser
	treeSitterParser *TreeSitterParser
	configParser     *ConfigParser
	certParser       *CertParser
	secretScanner    *SecretScanner
	tlsScanner       *TLSScanner
	containerScanner *ContainerScanner
	symbolGraph      *SymbolGraph
}

func NewScanner() *Scanner {
	rulesCfg, err := LoadRulesFromYAML("rules.yaml")
	if err != nil {
		rulesCfg, err = LoadRulesFromYAML("../rules.yaml")
	}
	if err != nil {
		rulesCfg, err = LoadRulesFromYAML("../../rules.yaml")
	}
	if err != nil || rulesCfg == nil || len(rulesCfg.SourceCodeRules) == 0 {
		rulesCfg = buildDefaultRulesConfig()
	}

	return &Scanner{
		rulesConfig:      rulesCfg,
		astParser:        NewASTParser(),
		treeSitterParser: NewTreeSitterParser(rulesCfg),
		configParser:     NewConfigParser(),
		certParser:       NewCertParser(),
		secretScanner:    NewSecretScanner(rulesCfg),
		tlsScanner:       NewTLSScanner(10 * time.Second),
		containerScanner: NewContainerScanner(),
		symbolGraph:      NewSymbolGraph(),
	}
}

func buildDefaultRulesConfig() *RulesConfig {
	rules := []YAMLRule{
		{
			ID:             "JS-AES-01",
			Name:           "AES-256-GCM Cipher",
			Pattern:        `createCipheriv\(\s*['"]aes-256-gcm['"]`,
			Languages:      []string{"javascript"},
			Algorithm:      "AES-256-GCM",
			KeySize:        256,
			RiskLevel:      "SAFE",
			RiskRationale:  "Symmetric AES-256 is post-quantum safe.",
			PQCStatus:      "QUANTUM_SAFE",
			PQCReplacement: "None Required",
			Confidence:     1.0,
		},
		{
			ID:             "JS-KEYGEN-01",
			Name:           "RSA Key Generation",
			Pattern:        `generateKeyPair(?:Sync)?\(\s*['"]rsa['"]`,
			Languages:      []string{"javascript"},
			Algorithm:      "RSA",
			KeySize:        2048,
			KeySizeCapture: `modulusLength:\s*(\d+)`,
			RiskLevel:      "HIGH",
			RiskRationale:  "RSA key generation is vulnerable to quantum Shor's algorithm.",
			PQCStatus:      "VULNERABLE",
			PQCReplacement: "ML-KEM-768",
			Confidence:     1.0,
		},
		{
			ID:             "JS-JWT-01",
			Name:           "JWT Signing RS256",
			Pattern:        `jsonwebtoken|jwt\.sign`,
			Languages:      []string{"javascript"},
			Algorithm:      "RS256",
			KeySize:        2048,
			RiskLevel:      "HIGH",
			RiskRationale:  "RSA JWT signing is quantum-vulnerable.",
			PQCStatus:      "VULNERABLE",
			PQCReplacement: "ML-DSA-65",
			Confidence:     1.0,
		},
	}

	for i := range rules {
		rules[i].CompiledRegex = regexp.MustCompile(rules[i].Pattern)
		if rules[i].KeySizeCapture != "" {
			rules[i].CompiledCapture = regexp.MustCompile(rules[i].KeySizeCapture)
		}
	}

	return &RulesConfig{
		Version:         "1.0-default",
		SourceCodeRules: rules,
	}
}

// ScanDirectory walks the target directory and parses code, certs, configs, and secrets
func (s *Scanner) ScanDirectory(targetDir string, scanID string) ([]models.CryptoAsset, int, int, error) {
	targetDir = strings.TrimSpace(targetDir)
	cleanTarget := strings.ReplaceAll(targetDir, "\\", "/")

	// 1. Check if directory exists directly (e.g. running on Windows host or relative path)
	if stat, err := os.Stat(targetDir); err == nil && stat.IsDir() {
		// Valid directly
	} else if stat, err := os.Stat(cleanTarget); err == nil && stat.IsDir() {
		targetDir = cleanTarget
	} else {
		resolved := false

		// 2. Dynamic Host Drive Translation for Docker (e.g. C:/Users/harin/Desktop -> /host_c/Users/harin/Desktop or /Users/harin/Desktop)
		if len(cleanTarget) >= 2 && (cleanTarget[1] == ':' || strings.HasPrefix(strings.ToLower(cleanTarget), "c:/")) {
			colonIdx := strings.Index(cleanTarget, ":")
			var relFromDrive string
			if colonIdx != -1 && len(cleanTarget) > colonIdx+1 {
				relFromDrive = strings.TrimPrefix(cleanTarget[colonIdx+1:], "/")
			} else {
				relFromDrive = cleanTarget
			}

			cCandidates := []string{
				filepath.Join("/host_c", relFromDrive),
				filepath.Join("/c", relFromDrive),
				filepath.Join("/", relFromDrive),
			}
			for _, candidate := range cCandidates {
				if statCandidate, errCandidate := os.Stat(candidate); errCandidate == nil && statCandidate.IsDir() {
					targetDir = candidate
					resolved = true
					break
				}
			}
		}

		// 3. Fallback to baseName relative resolution (e.g. sample_target)
		if !resolved {
			baseName := filepath.Base(cleanTarget)
			if baseName != "" && baseName != "." && baseName != "/" {
				candidates := []string{
					baseName,
					filepath.Join(".", baseName),
					filepath.Join("..", baseName),
					filepath.Join("../..", baseName),
					filepath.Join("../../..", baseName),
					filepath.Join("/app", baseName),
					filepath.Join("/", baseName),
				}
				for _, candidate := range candidates {
					if statCandidate, errCandidate := os.Stat(candidate); errCandidate == nil && statCandidate.IsDir() {
						targetDir = candidate
						resolved = true
						break
					}
				}
			}
		}

		if !resolved {
			return nil, 0, 0, fmt.Errorf("directory not found or inaccessible: '%s'", targetDir)
		}
	}

	var assets []models.CryptoAsset
	totalFiles := 0
	scannedFiles := 0

	// 1. Source Code, Certificate, Config Scan
	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			name := strings.ToLower(info.Name())
			switch name {
			// === VCS ===
			case ".git", ".svn", ".hg":
				return filepath.SkipDir

			// === JavaScript / TypeScript / Node.js ===
			case "node_modules", ".next", ".nuxt", ".turbo", ".parcel-cache", "bower_components":
				return filepath.SkipDir

			// === Go ===
			case "vendor": // Go modules vendor
				return filepath.SkipDir

			// === Java / Kotlin / Scala / Android ===
			case ".gradle", ".m2", ".mvn", ".idea":
				return filepath.SkipDir
			case "target": // Maven/Gradle build output
				return filepath.SkipDir

			// === .NET / C# ===
			case "packages", "obj":
				return filepath.SkipDir

			// === Python ===
			case "__pycache__", "venv", ".venv", ".tox", ".mypy_cache",
				".pytest_cache", ".eggs", "*.egg-info":
				return filepath.SkipDir

			// === Rust ===
			case ".cargo":
				return filepath.SkipDir

			// === Ruby ===
			case ".bundle":
				return filepath.SkipDir

			// === PHP ===
			// "vendor" already handled above

			// === Generic build / dist / output ===
			case "bin", "dist", "build", "out", "output", "release", "debug",
				"artifacts", ".cache", "tmp", "temp", "logs",
				"coverage", ".nyc_output", "htmlcov":
				return filepath.SkipDir

			// === IDE / Editor ===
			case ".vs", ".vscode", ".settings", ".classpath", ".project":
				return filepath.SkipDir

			// === Docker / Terraform state ===
			case ".terraform":
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))

		// Certificate file handling (multi-PEM via cert_parser)
		if ext == ".crt" || ext == ".pem" || ext == ".cer" {
			totalFiles++
			scannedFiles++
			certAssets := s.certParser.ParseCertFile(path, targetDir, scanID)
			for i := range certAssets {
				certAssets[i].BusinessScope, certAssets[i].ScopeWeight = classifyBusinessScope(certAssets[i].Location)
			}
			assets = append(assets, certAssets...)
			return nil
		}

		// PKCS#12 / PFX handling
		if ext == ".p12" || ext == ".pfx" {
			totalFiles++
			scannedFiles++
			p12Assets := s.certParser.ParsePKCS12File(path, targetDir, scanID)
			for i := range p12Assets {
				p12Assets[i].BusinessScope, p12Assets[i].ScopeWeight = classifyBusinessScope(p12Assets[i].Location)
			}
			assets = append(assets, p12Assets...)
			return nil
		}

		// JKS Java Keystore handling
		if ext == ".jks" {
			totalFiles++
			scannedFiles++
			jksAssets := s.certParser.ParseJKSFile(path, targetDir, scanID)
			for i := range jksAssets {
				jksAssets[i].BusinessScope, jksAssets[i].ScopeWeight = classifyBusinessScope(jksAssets[i].Location)
			}
			assets = append(assets, jksAssets...)
			return nil
		}

		if isScannableExtension(ext, info.Name()) {
			totalFiles++
			scannedFiles++

			// Structured config file parsing (YAML/JSON/TOML)
			if ext == ".yaml" || ext == ".yml" || ext == ".json" || ext == ".toml" {
				relPath := path
				if rel, err := filepath.Rel(targetDir, path); err == nil {
					relPath = rel
				}
				configAssets := s.configParser.ParseConfigFile(path, relPath, scanID)
				for i := range configAssets {
					configAssets[i].BusinessScope, configAssets[i].ScopeWeight = classifyBusinessScope(configAssets[i].Location)
				}
				assets = append(assets, configAssets...)
			}

			// Standard source code rule scanning
			codeAssets := s.parseCodeFile(path, targetDir, scanID)
			assets = append(assets, codeAssets...)
		}

		return nil
	})

	// 2. Embedded Gitleaks Secret Scanner & Git Commit History Scanner
	secretAssets, _ := s.secretScanner.ScanSecrets(targetDir)
	for i := range secretAssets {
		secretAssets[i].ScanID = scanID
		secretAssets[i].BusinessScope, secretAssets[i].ScopeWeight = classifyBusinessScope(secretAssets[i].Location)
		assets = append(assets, secretAssets[i])
	}

	// 3. Container / Image Dockerfile & Syft Package Scanner
	containerAssets, _ := s.containerScanner.ScanContainerImage(targetDir, scanID)
	for i := range containerAssets {
		containerAssets[i].BusinessScope, containerAssets[i].ScopeWeight = classifyBusinessScope(containerAssets[i].Location)
		assets = append(assets, containerAssets[i])
	}

	return assets, totalFiles, scannedFiles, err
}

// ScanContainerImage delegates container image scanning
func (s *Scanner) ScanContainerImage(imageName string, scanID string) ([]models.CryptoAsset, error) {
	return s.containerScanner.ScanContainerImage(imageName, scanID)
}

// AutoDiscoverTLSEndpoints delegates TLS port discovery to tlsScanner
func (s *Scanner) AutoDiscoverTLSEndpoints(targetHost string) []string {
	return s.tlsScanner.AutoDiscoverTLSEndpoints(targetHost)
}

// ScanTLSEndpoints connects to live host:port targets and performs active TLS probing
func (s *Scanner) ScanTLSEndpoints(targets []string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset
	for _, t := range targets {
		if strings.TrimSpace(t) == "" {
			continue
		}
		tlsAssets, err := s.tlsScanner.ScanTLSEndpoint(t)
		if err == nil {
			for i := range tlsAssets {
				tlsAssets[i].ScanID = scanID
				assets = append(assets, tlsAssets[i])
			}
		}
	}
	return assets
}

// parseCodeFile scans a source code file using Tree-Sitter AST syntax engine with line fallback
func (s *Scanner) parseCodeFile(path string, targetDir string, scanID string) []models.CryptoAsset {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	relPath := path
	if rel, err := filepath.Rel(targetDir, path); err == nil {
		relPath = rel
	}

	// Parse symbol exports and import call-sites across project files
	if s.symbolGraph != nil {
		s.symbolGraph.ParseImportsAndExports(string(content), relPath)
	}

	var assets []models.CryptoAsset

	// 1. Try Tree-Sitter AST Syntax Tree Parsing (JS, TS, Py, Go, Java)
	if s.treeSitterParser != nil {
		if tsAssets, ok := s.treeSitterParser.ParseCodeWithTreeSitter(content, relPath, scanID); ok {
			for i := range tsAssets {
				tsAssets[i].BusinessScope, tsAssets[i].ScopeWeight = classifyBusinessScope(relPath)
				if s.symbolGraph != nil {
					s.symbolGraph.AttachCallSiteContext(tsAssets[i].Name, &tsAssets[i])
				}
			}
			return tsAssets
		}
	}

	// 2. Fallback to line-based AST parser for remaining languages
	assets = s.parseCodeFileLineFallback(path, targetDir, scanID, string(content))
	for i := range assets {
		if s.symbolGraph != nil {
			s.symbolGraph.AttachCallSiteContext(assets[i].Name, &assets[i])
		}
	}
	return assets
}

func (s *Scanner) parseCodeFileLineFallback(path string, targetDir string, scanID string, content string) []models.CryptoAsset {
	var assets []models.CryptoAsset
	relPath := path
	if rel, err := filepath.Rel(targetDir, path); err == nil {
		relPath = rel
	}

	lines := strings.Split(content, "\n")
	s.astParser.ResetState()
	ext := strings.ToLower(filepath.Ext(path))

	for i, line := range lines {
		lineNumber := i + 1
		astNode := s.astParser.ParseLineToNode(line, lineNumber)

		start := i - 5
		if start < 0 {
			start = 0
		}
		end := i + 6
		if end > len(lines) {
			end = len(lines)
		}
		contextLines := lines[start:end]

		if s.rulesConfig != nil {
			for _, rule := range s.rulesConfig.SourceCodeRules {
				if matchLanguage(ext, rule.Languages) {
					if asset := s.astParser.MatchYAMLRuleAgainstAST(astNode, rule, scanID, relPath, contextLines); asset != nil {
						asset.BusinessScope, asset.ScopeWeight = classifyBusinessScope(relPath)
						// Fallback matches get 0.8 confidence
						if asset.Confidence == 1.0 {
							asset.Confidence = 0.8
						}
						assets = append(assets, *asset)
					}
				}
			}

			for _, rule := range s.rulesConfig.TLSConfigRules {
				if matchLanguage(ext, rule.Languages) {
					if asset := s.astParser.MatchYAMLRuleAgainstAST(astNode, rule, scanID, relPath, contextLines); asset != nil {
						asset.BusinessScope, asset.ScopeWeight = classifyBusinessScope(relPath)
						if asset.Confidence == 1.0 {
							asset.Confidence = 0.8
						}
						assets = append(assets, *asset)
					}
				}
			}
		}
	}

	return assets
}

// classifyBusinessScope auto-detects business scope from file path for risk weighting
func classifyBusinessScope(relPath string) (string, float64) {
	lower := strings.ToLower(relPath)

	// Test/fixture patterns → low weight
	testPatterns := []string{
		"test", "spec", "mock", "fixture", "example", "sample",
		"demo", "__test__", "_test.go", "_test.py", "_test.js",
		"testdata", "tests/", "test/",
	}
	for _, p := range testPatterns {
		if strings.Contains(lower, p) {
			return "TEST", 0.3
		}
	}

	// Staging/dev patterns → medium weight
	stagingPatterns := []string{"staging", "dev", "development", "sandbox", "debug"}
	for _, p := range stagingPatterns {
		if strings.Contains(lower, p) {
			return "STAGING", 0.7
		}
	}

	// Everything else is production → high weight
	return "PRODUCTION", 1.5
}

func isScannableExtension(ext string, filename string) bool {
	if strings.HasPrefix(strings.ToLower(filename), ".env") {
		return true
	}
	scannable := map[string]bool{
		// JavaScript / TypeScript
		".js": true, ".mjs": true, ".cjs": true, ".ts": true, ".tsx": true,
		// Python
		".py": true,
		// Java / Kotlin
		".java": true, ".kt": true, ".kts": true,
		// Go
		".go": true,
		// C / C++
		".c": true, ".cpp": true, ".cc": true, ".cxx": true, ".h": true, ".hpp": true,
		// C#
		".cs": true,
		// Rust
		".rs": true,
		// Ruby
		".rb": true,
		// PHP
		".php": true,
		// Infrastructure / Config
		".tf": true, ".hcl": true, ".toml": true, ".ini": true,
		".conf": true, ".cfg": true, ".htaccess": true,
		// Structured Data
		".json": true, ".yaml": true, ".yml": true,
	// Environment
		".env": true,
	}
	return scannable[ext]
}

// matchLanguage checks if a rule's specified languages match the target file extension
func matchLanguage(ext string, ruleLangs []string) bool {
	if len(ruleLangs) == 0 {
		return true // Rule applies to all languages
	}

	langMap := map[string][]string{
		".js":        {"javascript", "typescript", "js", "ts"},
		".mjs":       {"javascript", "js"},
		".cjs":       {"javascript", "js"},
		".ts":        {"typescript", "javascript", "ts"},
		".tsx":       {"typescript", "javascript", "tsx"},
		".py":        {"python", "py"},
		".java":      {"java"},
		".kt":        {"kotlin", "kt"},
		".kts":       {"kotlin", "kt"},
		".go":        {"go"},
		".c":         {"c"},
		".cpp":       {"cpp", "c"},
		".cc":        {"cpp", "c"},
		".cxx":       {"cpp", "c"},
		".h":         {"c", "cpp"},
		".hpp":       {"cpp"},
		".cs":        {"csharp", "cs"},
		".rs":        {"rust", "rs"},
		".rb":        {"ruby", "rb"},
		".php":       {"php"},
		".tf":        {"terraform", "tf"},
		".hcl":       {"terraform", "hcl"},
		".conf":      {"config"},
		".cfg":       {"config"},
		".htaccess":  {"config"},
		".yaml":      {"config", "yaml"},
		".yml":       {"config", "yaml"},
		".json":      {"config", "json"},
		".toml":      {"config", "toml"},
		".ini":       {"config", "ini"},
	}

	supported, ok := langMap[ext]
	if !ok {
		return true
	}

	for _, rl := range ruleLangs {
		rlLower := strings.ToLower(rl)
		for _, s := range supported {
			if rlLower == s {
				return true
			}
		}
	}
	return false
}

