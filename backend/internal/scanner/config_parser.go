package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
	"gopkg.in/yaml.v3"
)

// ConfigParser detects cryptographic configuration in structured files (YAML, JSON, TOML, INI)
type ConfigParser struct{}

func NewConfigParser() *ConfigParser {
	return &ConfigParser{}
}

// weakAlgoMap maps known weak algorithm strings to risk metadata
var weakAlgoMap = map[string]struct {
	Algorithm      string
	RiskLevel      models.RiskLevel
	PQCStatus      models.PQCStatus
	PQCReplacement string
	RiskRationale  string
}{
	"des":        {"DES", models.RiskCritical, models.StatusVulnerable, "AES-256-GCM", "DES cipher configured — obsolete and easily cracked."},
	"3des":       {"3DES", models.RiskCritical, models.StatusVulnerable, "AES-256-GCM", "Triple DES configured — deprecated due to SWEET32 attack."},
	"rc4":        {"RC4", models.RiskCritical, models.StatusVulnerable, "AES-256-GCM or ChaCha20", "RC4 stream cipher configured — known statistical biases."},
	"md5":        {"MD5", models.RiskCritical, models.StatusVulnerable, "SHA-256 or SHA-3", "MD5 hash configured — cryptographically broken, collision attacks trivial."},
	"sha1":       {"SHA-1", models.RiskHigh, models.StatusVulnerable, "SHA-256 or SHA-3", "SHA-1 hash configured — deprecated by NIST, collision attacks demonstrated."},
	"sha-1":      {"SHA-1", models.RiskHigh, models.StatusVulnerable, "SHA-256 or SHA-3", "SHA-1 hash configured — deprecated by NIST."},
	"rsa":        {"RSA", models.RiskHigh, models.StatusVulnerable, "ML-KEM-768 / ML-DSA-65", "RSA algorithm configured — vulnerable to quantum Shor's algorithm."},
	"rs256":      {"RSA-SHA256 (RS256)", models.RiskHigh, models.StatusVulnerable, "ML-DSA-65", "RS256 JWT algorithm configured — RSA signature vulnerable to quantum attacks."},
	"rs384":      {"RSA-SHA384 (RS384)", models.RiskHigh, models.StatusVulnerable, "ML-DSA-65", "RS384 JWT algorithm configured — RSA signature vulnerable to quantum attacks."},
	"rs512":      {"RSA-SHA512 (RS512)", models.RiskHigh, models.StatusVulnerable, "ML-DSA-65", "RS512 JWT algorithm configured — RSA signature vulnerable to quantum attacks."},
	"es256":      {"ECDSA-P256 (ES256)", models.RiskHigh, models.StatusVulnerable, "ML-DSA-44 or SLH-DSA", "ES256 JWT algorithm configured — ECC vulnerable to quantum attacks."},
	"ecdsa":      {"ECDSA", models.RiskHigh, models.StatusVulnerable, "ML-DSA-65", "ECDSA algorithm configured — vulnerable to quantum Shor's algorithm."},
	"ecdh":       {"ECDH", models.RiskHigh, models.StatusVulnerable, "ML-KEM-768", "ECDH key exchange configured — vulnerable to quantum attacks."},
	"diffie-hellman": {"Diffie-Hellman", models.RiskHigh, models.StatusVulnerable, "ML-KEM-768", "DH key exchange configured — vulnerable to quantum attacks."},
	"dh":        {"Diffie-Hellman", models.RiskHigh, models.StatusVulnerable, "ML-KEM-768", "Diffie-Hellman key exchange configured — quantum-vulnerable."},
	"blowfish":  {"Blowfish", models.RiskHigh, models.StatusVulnerable, "AES-256-GCM", "Blowfish cipher configured — 64-bit block size makes it vulnerable."},
	"sslv3":     {"SSLv3", models.RiskCritical, models.StatusVulnerable, "TLS 1.3", "SSLv3 configured — POODLE attack, completely broken."},
	"tlsv1":     {"TLS 1.0", models.RiskCritical, models.StatusVulnerable, "TLS 1.3", "TLS 1.0 configured — deprecated, BEAST/POODLE vulnerable."},
	"tlsv1.0":   {"TLS 1.0", models.RiskCritical, models.StatusVulnerable, "TLS 1.3", "TLS 1.0 configured — deprecated."},
	"tls 1.0":   {"TLS 1.0", models.RiskCritical, models.StatusVulnerable, "TLS 1.3", "TLS 1.0 configured — deprecated."},
	"tlsv1.1":   {"TLS 1.1", models.RiskCritical, models.StatusVulnerable, "TLS 1.3", "TLS 1.1 configured — deprecated."},
	"tls 1.1":   {"TLS 1.1", models.RiskCritical, models.StatusVulnerable, "TLS 1.3", "TLS 1.1 configured — deprecated."},
}

// cryptoKeyPatterns are key names that indicate cryptographic configuration
var cryptoKeyPatterns = []string{
	"algorithm", "cipher", "encryption", "hash", "digest",
	"signing_algorithm", "jwt_algorithm", "encryption_method",
	"ssl_protocol", "tls_version", "protocol", "kex_algorithm",
	"key_exchange", "mac_algorithm",
}

// keySizeKeyPatterns are key names indicating key/bit sizes
var keySizeKeyPatterns = []string{
	"key_size", "keysize", "key_length", "keylength",
	"bits", "modulus_length", "moduluslength", "rsa_bits",
	"key_bits", "dh_bits",
}

// ParseConfigFile scans a structured config file for cryptographic settings
func (cp *ConfigParser) ParseConfigFile(path string, relPath string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset

	content, err := os.ReadFile(path)
	if err != nil {
		return assets
	}

	ext := strings.ToLower(filepath.Ext(path))

	var data interface{}
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(content, &data); err != nil {
			return assets
		}
	case ".json":
		if err := json.Unmarshal(content, &data); err != nil {
			return assets
		}
	default:
		return assets
	}

	// Recursively walk the parsed config tree
	cp.walkConfigTree(data, "", relPath, scanID, &assets)
	return assets
}

// walkConfigTree recursively walks a parsed config tree looking for crypto-relevant keys
func (cp *ConfigParser) walkConfigTree(node interface{}, keyPath string, relPath string, scanID string, assets *[]models.CryptoAsset) {
	switch v := node.(type) {
	case map[string]interface{}:
		for key, val := range v {
			fullKey := key
			if keyPath != "" {
				fullKey = keyPath + "." + key
			}
			cp.walkConfigTree(val, fullKey, relPath, scanID, assets)
		}
	case map[interface{}]interface{}:
		// YAML sometimes produces this type
		for key, val := range v {
			keyStr := fmt.Sprintf("%v", key)
			fullKey := keyStr
			if keyPath != "" {
				fullKey = keyPath + "." + keyStr
			}
			cp.walkConfigTree(val, fullKey, relPath, scanID, assets)
		}
	case []interface{}:
		for i, item := range v {
			cp.walkConfigTree(item, fmt.Sprintf("%s[%d]", keyPath, i), relPath, scanID, assets)
		}
	case string:
		cp.checkConfigValue(keyPath, v, relPath, scanID, assets)
	case float64:
		cp.checkConfigKeySize(keyPath, int(v), relPath, scanID, assets)
	case int:
		cp.checkConfigKeySize(keyPath, v, relPath, scanID, assets)
	}
}

// checkConfigValue checks if a string config value represents a weak crypto algorithm
func (cp *ConfigParser) checkConfigValue(keyPath string, value string, relPath string, scanID string, assets *[]models.CryptoAsset) {
	lowerKey := strings.ToLower(keyPath)
	lowerVal := strings.ToLower(strings.TrimSpace(value))

	// Only check keys that look crypto-related
	isCryptoKey := false
	for _, pattern := range cryptoKeyPatterns {
		if strings.Contains(lowerKey, pattern) {
			isCryptoKey = true
			break
		}
	}

	if !isCryptoKey {
		return
	}

	// Check value against known weak algorithms
	if info, found := weakAlgoMap[lowerVal]; found {
		*assets = append(*assets, models.CryptoAsset{
			ID:             models.GenerateUUID(),
			ScanID:         scanID,
			Name:           fmt.Sprintf("Config: %s = \"%s\"", keyPath, value),
			AssetType:      models.AssetAlgorithm,
			Type:           models.AssetAlgorithm,
			Location:       relPath,
			LineNumber:     0,
			Algorithm:      info.Algorithm,
			KeySize:        0,
			RiskLevel:      info.RiskLevel,
			RiskRationale:  info.RiskRationale,
			PQCStatus:      info.PQCStatus,
			PQCReplacement: info.PQCReplacement,
			Confidence:     0.9, // Config key-value match
			ContextSnippet: fmt.Sprintf("%s: %s", keyPath, value),
			DiscoveredAt:   time.Now(),
		})
	}

	// Check RSA key size parameters (handle as string)
	if (strings.Contains(lowerKey, "keysize") || strings.Contains(lowerKey, "key_length") || strings.Contains(lowerKey, "bits")) {
		if valInt, err := strconv.Atoi(lowerVal); err == nil && valInt > 0 {
			cp.checkConfigKeySize(keyPath, valInt, relPath, scanID, assets)
		}
	}
}

// checkConfigKeySize checks if a numeric config value represents a weak key size
func (cp *ConfigParser) checkConfigKeySize(keyPath string, value int, relPath string, scanID string, assets *[]models.CryptoAsset) {
	lowerKey := strings.ToLower(keyPath)

	isKeySizeKey := false
	for _, pattern := range keySizeKeyPatterns {
		if strings.Contains(lowerKey, pattern) {
			isKeySizeKey = true
			break
		}
	}

	if !isKeySizeKey || value <= 0 {
		return
	}

	// Flag weak key sizes
	var risk models.RiskLevel
	var rationale string

	if value <= 512 {
		risk = models.RiskCritical
		rationale = fmt.Sprintf("Critically weak key size (%d bits) configured at %s. Trivially breakable.", value, keyPath)
	} else if value <= 1024 {
		risk = models.RiskCritical
		rationale = fmt.Sprintf("Weak key size (%d bits) configured at %s. Factorable by classical computers.", value, keyPath)
	} else if value < 2048 {
		risk = models.RiskHigh
		rationale = fmt.Sprintf("Below-standard key size (%d bits) configured at %s. Below NIST minimum recommendation.", value, keyPath)
	} else {
		return // 2048+ is acceptable (though still quantum-vulnerable for RSA)
	}

	*assets = append(*assets, models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           fmt.Sprintf("Config: Weak Key Size (%d bits)", value),
		AssetType:      models.AssetAlgorithm,
		Type:           models.AssetAlgorithm,
		Location:       relPath,
		LineNumber:     0,
		Algorithm:      fmt.Sprintf("KeySize-%d", value),
		KeySize:        value,
		RiskLevel:      risk,
		RiskRationale:  rationale,
		PQCStatus:      models.StatusVulnerable,
		PQCReplacement: "ML-KEM-768 (2048+ equivalent PQC)",
		Confidence:     0.9,
		ContextSnippet: fmt.Sprintf("%s: %d", keyPath, value),
		DiscoveredAt:   time.Now(),
	})
}
