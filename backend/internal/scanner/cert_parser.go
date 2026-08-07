package scanner

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pavlo-v-chernykh/keystore-go/v4"
	"github.com/quantumpq/qv-cbom/internal/models"
	"software.sslmate.com/src/go-pkcs12"
)

var commonKeystorePasswords = []string{"changeit", "", "password", "123456", "secret"}

// CertParser provides enhanced certificate and keystore analysis: multi-PEM, PKCS#12, JKS, expiry, self-signed, CA detection
type CertParser struct{}

func NewCertParser() *CertParser {
	return &CertParser{}
}

// ParseCertFile parses all PEM blocks in a file (multi-PEM / fullchain.pem support)
func (cp *CertParser) ParseCertFile(path string, targetDir string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset
	content, err := os.ReadFile(path)
	if err != nil {
		return assets
	}

	relPath := path
	if rel, err := filepath.Rel(targetDir, path); err == nil {
		relPath = rel
	}

	rest := content
	blockIndex := 0
	for {
		block, remaining := pem.Decode(rest)
		if block == nil {
			break
		}
		rest = remaining
		blockIndex++

		if strings.Contains(block.Type, "PRIVATE KEY") {
			assets = append(assets, cp.createPrivateKeyAsset(block, relPath, blockIndex, scanID))
			continue
		}

		if block.Type != "CERTIFICATE" && block.Type != "X509 CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		asset := cp.AnalyzeCertificate(cert, relPath, blockIndex, scanID)
		assets = append(assets, asset)
	}

	return assets
}

// ParsePKCS12File decodes .p12 / .pfx keystores, detects weak default passwords, and extracts certs/keys
func (cp *CertParser) ParsePKCS12File(path string, targetDir string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset
	content, err := os.ReadFile(path)
	if err != nil {
		return assets
	}

	relPath := path
	if rel, err := filepath.Rel(targetDir, path); err == nil {
		relPath = rel
	}

	for _, pwd := range commonKeystorePasswords {
		key, cert, caCerts, err := pkcs12.DecodeChain(content, pwd)
		if err == nil {
			// SECURITY SIGNAL: Default Weak Password Finding
			assets = append(assets, models.CryptoAsset{
				ID:             models.GenerateUUID(),
				ScanID:         scanID,
				Name:           fmt.Sprintf("Weak PKCS#12 Keystore Credential: %s", filepath.Base(path)),
				Type:           models.AssetKeyStore,
				AssetType:      models.AssetKeyStore,
				Location:       relPath,
				LineNumber:     1,
				Algorithm:      "PKCS12-WEAK-PASSWORD",
				KeySize:        0, // Fixed: 0 instead of byte length * 8
				RiskLevel:      models.RiskCritical,
				RiskRationale:  fmt.Sprintf("PKCS#12 keystore is protected by weak default password '%s' — easily crackable by attackers.", pwd),
				PQCStatus:      models.StatusVulnerable,
				PQCReplacement: "Enforce strong passphrase & store keys in QuantumVault KMS",
				Confidence:     1.0,
				ContextSnippet: fmt.Sprintf("Decrypted with default password: '%s'", pwd),
				DiscoveredAt:   time.Now(),
			})

			if cert != nil {
				assets = append(assets, cp.AnalyzeCertificate(cert, relPath, 1, scanID))
			}
			if key != nil {
				if keyAsset := cp.analyzePrivateKeyObject(key, relPath, 1, scanID); keyAsset != nil {
					assets = append(assets, *keyAsset)
				}
			}
			for i, ca := range caCerts {
				assets = append(assets, cp.AnalyzeCertificate(ca, relPath, i+2, scanID))
			}
			return assets
		}
	}

	// Password-protected with non-default password
	assets = append(assets, cp.encryptedKeystoreAsset(path, relPath, "PKCS12", scanID))
	return assets
}

// ParseJKSFile decodes Java Keystores (.jks), handles per-entry passwords, and measures key sizes
func (cp *CertParser) ParseJKSFile(path string, targetDir string, scanID string) []models.CryptoAsset {
	var assets []models.CryptoAsset
	f, err := os.Open(path)
	if err != nil {
		return assets
	}
	defer f.Close()

	relPath := path
	if rel, err := filepath.Rel(targetDir, path); err == nil {
		relPath = rel
	}

	ks := keystore.New()
	for _, pwd := range commonKeystorePasswords {
		f.Seek(0, 0)
		if err := ks.Load(f, []byte(pwd)); err == nil {
			// SECURITY SIGNAL: Default Store Password Finding
			assets = append(assets, models.CryptoAsset{
				ID:             models.GenerateUUID(),
				ScanID:         scanID,
				Name:           fmt.Sprintf("Weak JKS Keystore Credential: %s", filepath.Base(path)),
				Type:           models.AssetKeyStore,
				AssetType:      models.AssetKeyStore,
				Location:       relPath,
				LineNumber:     1,
				Algorithm:      "JKS-WEAK-PASSWORD",
				KeySize:        0, // Fixed: 0 instead of byte length * 8
				RiskLevel:      models.RiskCritical,
				RiskRationale:  fmt.Sprintf("Java Keystore (JKS) is protected by weak default password '%s'.", pwd),
				PQCStatus:      models.StatusVulnerable,
				PQCReplacement: "Migrate to PKCS#12 or QuantumVault KMS",
				Confidence:     1.0,
				ContextSnippet: fmt.Sprintf("JKS Store Password: '%s'", pwd),
				DiscoveredAt:   time.Now(),
			})

			for _, alias := range ks.Aliases() {
				// 1. Trusted Certificate Entries
				if entry, err := ks.GetTrustedCertificateEntry(alias); err == nil {
					cert, err := x509.ParseCertificate(entry.Certificate.Content)
					if err == nil {
						assets = append(assets, cp.AnalyzeCertificate(cert, relPath, 0, scanID))
					}
				}

				// 2. Private Key Entries (Try store pwd first, then fallback per-entry passwords)
				for _, entryPwd := range append([]string{pwd}, commonKeystorePasswords...) {
					if entry, err := ks.GetPrivateKeyEntry(alias, []byte(entryPwd)); err == nil {
						if parsedKey, err := parsePrivateKeyBytes(entry.PrivateKey); err == nil {
							if keyAsset := cp.analyzePrivateKeyObject(parsedKey, relPath, 0, scanID); keyAsset != nil {
								assets = append(assets, *keyAsset)
							}
						}
						break
					}
				}
			}
			return assets
		}
	}

	assets = append(assets, cp.encryptedKeystoreAsset(path, relPath, "JKS", scanID))
	return assets
}

// AnalyzeCertificate performs deep analysis: algorithm, key size, expiry, self-signed, CA status
func (cp *CertParser) AnalyzeCertificate(cert *x509.Certificate, relPath string, index int, scanID string) models.CryptoAsset {
	algo := cert.SignatureAlgorithm.String()
	keySize := getCertPublicKeySize(cert)
	pqcStatus := models.StatusVulnerable
	pqcReplacement := "Hybrid ML-DSA-65 / ECDSA X.509 Certificate"
	riskLevel := models.RiskHigh

	riskRationale := fmt.Sprintf("X.509 Certificate #%d (%s) uses %s (%d-bit) — vulnerable to quantum Shor's algorithm.",
		index, cert.Subject.CommonName, algo, keySize)

	if strings.Contains(algo, "SHA1") || strings.Contains(algo, "MD5") || strings.Contains(algo, "MD2") {
		riskLevel = models.RiskCritical
		riskRationale = fmt.Sprintf("X.509 Certificate uses cryptographically broken signature algorithm %s.", algo)
	}

	if strings.Contains(algo, "ECDSA") {
		riskLevel = models.RiskMedium
		keySize = getCertPublicKeySize(cert)
		riskRationale = fmt.Sprintf("X.509 ECDSA certificate (%d-bit) requires post-quantum hybrid migration.", keySize)
	}

	now := time.Now()
	expiryNote := ""
	if cert.NotAfter.Before(now) {
		riskLevel = models.RiskCritical
		expiryNote = " ⚠ EXPIRED CERTIFICATE!"
		riskRationale += " Certificate has EXPIRED — immediate renewal required."
	} else if cert.NotAfter.Before(now.Add(90 * 24 * time.Hour)) {
		if riskLevel != models.RiskCritical {
			riskLevel = models.RiskHigh
		}
		daysLeft := int(cert.NotAfter.Sub(now).Hours() / 24)
		expiryNote = fmt.Sprintf(" ⚠ EXPIRING in %d days!", daysLeft)
		riskRationale += fmt.Sprintf(" Certificate expires in %d days — renewal recommended.", daysLeft)
	}

	selfSignedNote := ""
	if cert.Issuer.String() == cert.Subject.String() {
		selfSignedNote = " [SELF-SIGNED]"
		if riskLevel == models.RiskMedium || riskLevel == models.RiskLow {
			riskLevel = models.RiskHigh
		}
		riskRationale += " Self-signed certificate — no third-party trust chain validation."
	}

	caNote := ""
	if cert.IsCA {
		caNote = " [CA]"
		riskRationale += " This is a CA certificate — compromise would affect all issued certificates."
	}

	certName := fmt.Sprintf("X.509 Certificate (%s)%s%s%s",
		cert.Subject.CommonName, selfSignedNote, caNote, expiryNote)
	if cert.Subject.CommonName == "" {
		certName = fmt.Sprintf("X.509 Certificate #%d%s%s%s",
			index, selfSignedNote, caNote, expiryNote)
	}

	contextSnippet := fmt.Sprintf("Subject: %s | Issuer: %s | Expires: %s | SAN: %s",
		cert.Subject.CommonName,
		cert.Issuer.CommonName,
		cert.NotAfter.Format("2006-01-02"),
		strings.Join(cert.DNSNames, ", "))

	expiryTime := cert.NotAfter

	return models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           certName,
		Type:           models.AssetCertificate,
		Location:       relPath,
		LineNumber:     index,
		Algorithm:      algo,
		KeySize:        keySize,
		RiskLevel:      riskLevel,
		RiskRationale:  riskRationale,
		PQCStatus:      pqcStatus,
		PQCReplacement: pqcReplacement,
		Issuer:         cert.Issuer.CommonName,
		ExpiryDate:     &expiryTime,
		Confidence:     1.0,
		ContextSnippet: contextSnippet,
		DiscoveredAt:   time.Now(),
	}
}

// analyzePrivateKeyObject performs type-switch on decoded private key objects (RSA/ECDSA/Ed25519)
func (cp *CertParser) analyzePrivateKeyObject(key interface{}, relPath string, index int, scanID string) *models.CryptoAsset {
	algo := "Unknown"
	keySize := 0

	switch k := key.(type) {
	case *rsa.PrivateKey:
		algo = "RSA"
		keySize = k.N.BitLen()
	case *ecdsa.PrivateKey:
		algo = "ECDSA"
		keySize = k.Curve.Params().BitSize
	case ed25519.PrivateKey:
		algo = "Ed25519"
		keySize = 256
	default:
		return nil
	}

	return &models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           fmt.Sprintf("Keystore Private Key: %s (%d-bit)", algo, keySize),
		Type:           models.AssetKeyMaterial,
		AssetType:      models.AssetKeyMaterial,
		Location:       relPath,
		LineNumber:     index,
		Algorithm:      algo,
		KeySize:        keySize, // Actual bit size
		RiskLevel:      models.RiskCritical,
		RiskRationale:  fmt.Sprintf("Decrypted %s private key (%d-bit) found in keystore. Quantum-vulnerable asymmetric algorithm.", algo, keySize),
		PQCStatus:      models.StatusVulnerable,
		PQCReplacement: "ML-KEM / ML-DSA Post-Quantum Migration",
		Confidence:     1.0,
		ContextSnippet: fmt.Sprintf("Private Key Type: %s | Bit Length: %d", algo, keySize),
		DiscoveredAt:   time.Now(),
	}
}

// encryptedKeystoreAsset handles custom password protected keystores (KeySize: 0)
func (cp *CertParser) encryptedKeystoreAsset(path, relPath, format, scanID string) models.CryptoAsset {
	return models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           fmt.Sprintf("Encrypted Keystore: %s", filepath.Base(path)),
		Type:           models.AssetKeyStore,
		AssetType:      models.AssetKeyStore,
		Location:       relPath,
		LineNumber:     1,
		Algorithm:      format,
		KeySize:        0, // Fixed: 0 instead of byte length
		RiskLevel:      models.RiskHigh,
		RiskRationale:  fmt.Sprintf("Encrypted %s keystore found. Contents unreadable with standard default passwords.", format),
		PQCStatus:      models.StatusVulnerable,
		PQCReplacement: "Migrate to QuantumVault Cloud KMS",
		Confidence:     1.0,
		ContextSnippet: fmt.Sprintf("%s Keystore File: %s", format, filepath.Base(path)),
		DiscoveredAt:   time.Now(),
	}
}

// createPrivateKeyAsset flags a PEM private key block
func (cp *CertParser) createPrivateKeyAsset(block *pem.Block, relPath string, index int, scanID string) models.CryptoAsset {
	keyType := block.Type
	algo := "Unknown"
	keySize := 0

	if parsedKey, err := parsePrivateKeyBytes(block.Bytes); err == nil {
		switch k := parsedKey.(type) {
		case *rsa.PrivateKey:
			algo = "RSA"
			keySize = k.N.BitLen()
		case *ecdsa.PrivateKey:
			algo = "ECDSA"
			keySize = k.Curve.Params().BitSize
		case ed25519.PrivateKey:
			algo = "Ed25519"
			keySize = 256
		}
	} else if strings.Contains(keyType, "RSA") {
		algo = "RSA"
		keySize = 2048
	} else if strings.Contains(keyType, "EC") {
		algo = "ECDSA"
		keySize = 256
	}

	riskLevel := models.RiskCritical
	if strings.Contains(strings.ToUpper(block.Headers["Proc-Type"]), "ENCRYPTED") {
		riskLevel = models.RiskHigh
	}

	return models.CryptoAsset{
		ID:             models.GenerateUUID(),
		ScanID:         scanID,
		Name:           fmt.Sprintf("PEM Private Key #%d (%s)", index, algo),
		Type:           models.AssetKeyMaterial,
		AssetType:      models.AssetKeyMaterial,
		Location:       relPath,
		LineNumber:     index,
		Algorithm:      algo,
		KeySize:        keySize,
		RiskLevel:      riskLevel,
		RiskRationale:  fmt.Sprintf("Unencrypted %s private key found in PEM file. Key material should never be stored in project directories.", algo),
		PQCStatus:      models.StatusVulnerable,
		PQCReplacement: "Migrate to PQC key pair (ML-KEM / ML-DSA) via HSM/KMS",
		Confidence:     1.0,
		ContextSnippet: fmt.Sprintf("PEM Block Type: %s | Key Size: %d bits", keyType, keySize),
		DiscoveredAt:   time.Now(),
	}
}

func parsePrivateKeyBytes(der []byte) (interface{}, error) {
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("unknown private key format")
}

func getCertPublicKeySize(cert *x509.Certificate) int {
	switch k := cert.PublicKey.(type) {
	case *rsa.PublicKey:
		return k.N.BitLen()
	case *ecdsa.PublicKey:
		return k.Curve.Params().BitSize
	case ed25519.PublicKey:
		return 256
	default:
		return 256
	}
}
