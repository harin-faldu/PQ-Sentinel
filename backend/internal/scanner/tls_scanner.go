package scanner

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type TLSScanner struct {
	Timeout time.Duration
}

func NewTLSScanner(timeout time.Duration) *TLSScanner {
	if timeout == 0 {
		timeout = 4 * time.Second
	}
	return &TLSScanner{Timeout: timeout}
}

// AutoDiscoverTLSEndpoints scans standard TLS ports concurrently on a target host
func (ts *TLSScanner) AutoDiscoverTLSEndpoints(targetHost string) []string {
	host := targetHost
	if strings.Contains(host, ":") {
		return []string{targetHost}
	}

	standardPorts := []int{443, 8443, 636, 993, 995, 8080}
	var activeTargets []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, port := range standardPorts {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			addr := net.JoinHostPort(host, strconv.Itoa(p))
			conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
			if err == nil {
				conn.Close()
				mu.Lock()
				activeTargets = append(activeTargets, addr)
				mu.Unlock()
			}
		}(port)
	}

	wg.Wait()

	if len(activeTargets) == 0 {
		activeTargets = append(activeTargets, net.JoinHostPort(host, "443"))
	}
	return activeTargets
}

// ScanTLSEndpoint connects to a host:port via active TLS handshake and inspects protocol, cipher, cert chain, and vulnerabilities
func (ts *TLSScanner) ScanTLSEndpoint(target string) ([]models.CryptoAsset, error) {
	var assets []models.CryptoAsset

	if !strings.Contains(target, ":") {
		target = target + ":443"
	}

	dialer := &net.Dialer{Timeout: ts.Timeout}
	config := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", target, config)
	if err != nil {
		return nil, fmt.Errorf("active TLS connection failed to %s: %w", target, err)
	}
	defer conn.Close()

	state := conn.ConnectionState()

	// 1. Inspect TLS Version & Cipher Suite
	tlsVerName := getTLSVersionName(state.Version)
	cipherName := tls.CipherSuiteName(state.CipherSuite)

	tlsRisk := models.RiskSafe
	tlsRationale := "TLS protocol version " + tlsVerName + " with cipher " + cipherName
	pqcStatus := models.StatusSafe
	pqcReplacement := "PQC Hybrid KEM (X25519_Kyber768)"

	// Known Vulnerability Probes (POODLE, BEAST, FREAK)
	vulnFlags := []string{}
	if state.Version < tls.VersionTLS12 {
		tlsRisk = models.RiskCritical
		pqcStatus = models.StatusVulnerable
		vulnFlags = append(vulnFlags, "Vulnerable to POODLE/BEAST Attacks")
		tlsRationale = fmt.Sprintf("Insecure protocol version %s on %s. Vulnerable to POODLE/BEAST downgrade exploits.", tlsVerName, target)
		pqcReplacement = "Upgrade to TLS 1.3 with Hybrid PQC Exchange"
	} else if state.Version == tls.VersionTLS12 {
		tlsRisk = models.RiskMedium
		pqcStatus = models.StatusTransitional
		tlsRationale = fmt.Sprintf("TLS 1.2 in use on %s. Classical key exchange (ECDHE/RSA) vulnerable to store-now-decrypt-later PQC threat.", target)
		pqcReplacement = "Upgrade server to TLS 1.3 with PQC hybrid extension"
	}

	if strings.Contains(strings.ToUpper(cipherName), "EXPORT") || strings.Contains(strings.ToUpper(cipherName), "DES") || strings.Contains(strings.ToUpper(cipherName), "RC4") {
		tlsRisk = models.RiskCritical
		vulnFlags = append(vulnFlags, "Vulnerable to FREAK/SWEET32 cipher attacks")
	}

	contextSnippet := fmt.Sprintf("Protocol: %s | Cipher: %s | Server: %s", tlsVerName, cipherName, state.ServerName)
	if len(vulnFlags) > 0 {
		contextSnippet += " | Vulnerabilities: " + strings.Join(vulnFlags, ", ")
	}

	assets = append(assets, models.CryptoAsset{
		ID:             models.GenerateUUID(),
		Name:           "Active TLS Endpoint: " + target,
		Type:           models.AssetProtocol,
		Algorithm:      tlsVerName + " / " + cipherName,
		KeySize:        256,
		Location:       target,
		LineNumber:     0,
		ContextSnippet: contextSnippet,
		RiskLevel:      tlsRisk,
		PQCStatus:      pqcStatus,
		PQCReplacement: pqcReplacement,
		RiskRationale:  tlsRationale,
		DiscoveredAt:   time.Now(),
	})

	// 2. Inspect Peer Certificate Chain
	for i, cert := range state.PeerCertificates {
		certAsset := processCertificate(cert, target, i)
		assets = append(assets, certAsset)
	}

	return assets, nil
}

func processCertificate(cert *x509.Certificate, target string, index int) models.CryptoAsset {
	keyAlgo := cert.PublicKeyAlgorithm.String()
	keySize := getPublicKeySize(cert.PublicKey)

	riskLevel := models.RiskHigh
	pqcStatus := models.StatusVulnerable
	rationale := fmt.Sprintf("Certificate #%d (%s) uses classical public key algorithm %s (%d bits) which is vulnerable to Shor's algorithm.", index+1, cert.Subject.CommonName, keyAlgo, keySize)
	pqcReplacement := "ML-DSA (Dilithium) or Hybrid X.509 Certificate"

	if cert.NotAfter.Before(time.Now()) {
		riskLevel = models.RiskCritical
		rationale += " CERTIFICATE IS EXPIRED."
	}

	certName := fmt.Sprintf("TLS Cert: %s", cert.Subject.CommonName)
	if cert.Subject.CommonName == "" {
		certName = fmt.Sprintf("TLS Cert #%d (%s)", index+1, target)
	}

	return models.CryptoAsset{
		ID:             models.GenerateUUID(),
		Name:           certName,
		Type:           models.AssetCertificate,
		Algorithm:      fmt.Sprintf("%s-%d", keyAlgo, keySize),
		KeySize:        keySize,
		Location:       fmt.Sprintf("%s (Chain #%d)", target, index+1),
		LineNumber:     0,
		ContextSnippet: fmt.Sprintf("Subject: %s | Issuer: %s | Expires: %s", cert.Subject.CommonName, cert.Issuer.CommonName, cert.NotAfter.Format("2006-01-02")),
		RiskLevel:      riskLevel,
		PQCStatus:      pqcStatus,
		PQCReplacement: pqcReplacement,
		RiskRationale:  rationale,
		DiscoveredAt:   time.Now(),
	}
}

func getPublicKeySize(pub interface{}) int {
	switch k := pub.(type) {
	case *rsa.PublicKey:
		return k.N.BitLen()
	default:
		return 256
	}
}

func getTLSVersionName(ver uint16) string {
	switch ver {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("TLS 0x%04x", ver)
	}
}
