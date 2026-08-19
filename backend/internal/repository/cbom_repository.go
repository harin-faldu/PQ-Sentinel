package repository

import (
	"fmt"
	"sync"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type CBOMRepository interface {
	CreateTarget(t models.Target) error
	ListTargets() ([]models.Target, error)
	GetTarget(id string) (*models.Target, error)

	SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error
	GetScans() ([]models.ScanSummary, error)
	GetScanByID(id string) (*models.ScanDetail, error)
	GetScansByTarget(targetID string) ([]models.ScanSummary, error)
	GetLatestScanByTargetID(targetID string) (*models.ScanDetail, error)
	SaveMigrationPlan(plan models.MigrationPlan) error
	GetMigrationPlanByScanID(scanID string) (*models.MigrationPlan, error)
}

// MemoryRepository implements pure in-memory, thread-safe storage (Zero-DB / Stateless mode)
type MemoryRepository struct {
	mu             sync.RWMutex
	targets        map[string]models.Target
	scans          map[string]models.ScanSummary
	assets         map[string][]models.CryptoAsset
	migrationPlans map[string]models.MigrationPlan
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		targets:        make(map[string]models.Target),
		scans:          make(map[string]models.ScanSummary),
		assets:         make(map[string][]models.CryptoAsset),
		migrationPlans: make(map[string]models.MigrationPlan),
	}
	repo.seedInitialData()
	return repo
}

func (r *MemoryRepository) CreateTarget(t models.Target) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if t.ID == "" {
		t.ID = fmt.Sprintf("target-%s", models.GenerateUUID()[:8])
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	r.targets[t.ID] = t
	return nil
}

func (r *MemoryRepository) ListTargets() ([]models.Target, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []models.Target
	for _, t := range r.targets {
		list = append(list, t)
	}
	return list, nil
}

func (r *MemoryRepository) GetTarget(id string) (*models.Target, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.targets[id]
	if !exists {
		return nil, fmt.Errorf("target not found: %s", id)
	}
	return &t, nil
}

func (r *MemoryRepository) GetScansByTarget(targetID string) ([]models.ScanSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []models.ScanSummary
	for _, scan := range r.scans {
		if scan.TargetID == targetID {
			result = append(result, scan)
		}
	}
	return result, nil
}

func (r *MemoryRepository) SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now()
	}
	for i := range assets {
		if assets[i].CreatedAt.IsZero() {
			assets[i].CreatedAt = time.Now()
		}
	}

	r.scans[summary.ID] = summary
	r.assets[summary.ID] = assets
	return nil
}

func (r *MemoryRepository) GetScans() ([]models.ScanSummary, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []models.ScanSummary
	for _, scan := range r.scans {
		result = append(result, scan)
	}
	return result, nil
}

func (r *MemoryRepository) GetScanByID(id string) (*models.ScanDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	summary, exists := r.scans[id]
	if !exists {
		return nil, fmt.Errorf("scan not found: %s", id)
	}

	assets := r.assets[id]
	return &models.ScanDetail{
		Summary: summary,
		Assets:  assets,
	}, nil
}

func (r *MemoryRepository) GetLatestScanByTargetID(targetID string) (*models.ScanDetail, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latestSummary models.ScanSummary
	found := false

	for _, s := range r.scans {
		if s.TargetID == targetID {
			if !found || s.CreatedAt.After(latestSummary.CreatedAt) {
				latestSummary = s
				found = true
			}
		}
	}

	if !found {
		return nil, fmt.Errorf("no scans found for target: %s", targetID)
	}

	assets := r.assets[latestSummary.ID]
	return &models.ScanDetail{
		Summary: latestSummary,
		Assets:  assets,
	}, nil
}

func (r *MemoryRepository) SaveMigrationPlan(plan models.MigrationPlan) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.migrationPlans[plan.ScanID] = plan
	return nil
}

func (r *MemoryRepository) GetMigrationPlanByScanID(scanID string) (*models.MigrationPlan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plan, exists := r.migrationPlans[scanID]
	if !exists {
		return nil, fmt.Errorf("migration plan not found for scan: %s", scanID)
	}
	return &plan, nil
}

// Seed initial demonstration data so the UI displays rich results out-of-the-box
func (r *MemoryRepository) seedInitialData() {
	now := time.Now()
	expiryCert := now.AddDate(1, 2, 0)

	demoTargetID := "target-demo-001"
	r.targets[demoTargetID] = models.Target{
		ID:        demoTargetID,
		Name:      "Payment Gateway Service",
		Type:      models.TargetTypeRepo,
		Location:  "PaymentGatewayService/v2",
		CreatedAt: now.Add(-24 * time.Hour),
	}

	demoScanID := "scan-demo-001"
	summary := models.ScanSummary{
		ID:              demoScanID,
		TargetID:        demoTargetID,
		TargetDirectory: "PaymentGatewayService/v2",
		TotalFiles:      142,
		ScannedFiles:    142,
		TotalAssets:     8,
		CriticalCount:   2,
		HighCount:       3,
		MediumCount:     2,
		SafeCount:       1,
		PQCScore:        35, // Low score = high quantum vulnerability
		Status:          "COMPLETED",
		DurationMs:      1240,
		CreatedAt:       now.Add(-2 * time.Hour),
	}

	assets := []models.CryptoAsset{
		{
			ID:             "asset-01",
			ScanID:         demoScanID,
			Name:           "RSA-2048 JWT Signature",
			AssetType:      models.AssetAlgorithm,
			Location:       "services/auth/jwt_provider.go",
			LineNumber:     45,
			Algorithm:      "RS256",
			KeySize:        2048,
			RiskLevel:      models.RiskHigh,
			RiskRationale:  "RSA 2048-bit digital signature scheme for tokens can be forged by a quantum computer via Shor's Algorithm.",
			PQCStatus:      models.PQCStatusVulnerable,
			PQCReplacement: "ML-DSA-65",
			Details:        "Jwt.Sign(claims, rsaPrivateKey) - Vulnerable to Shor's algorithm.",
			CreatedAt:      now,
		},
		{
			ID:             "asset-02",
			ScanID:         demoScanID,
			Name:           "SHA-1 Message Digest",
			AssetType:      models.AssetAlgorithm,
			Location:       "utils/hash_checker.py",
			LineNumber:     18,
			Algorithm:      "SHA1",
			KeySize:        160,
			RiskLevel:      models.RiskCritical,
			RiskRationale:  "Cryptographically broken hash digest vulnerable to practical collision attacks (NIST SP 800-131A mandatory deprecation).",
			PQCStatus:      models.PQCStatusVulnerable,
			PQCReplacement: "SHA-384 / SHAKE256",
			Details:        "hashlib.sha1(data).hexdigest() - Cryptographically broken digest.",
			CreatedAt:      now,
		},
		{
			ID:             "asset-03",
			ScanID:         demoScanID,
			Name:           "TLS 1.2 ECDHE_RSA Endpoint",
			AssetType:      models.AssetProtocol,
			Location:       "config/nginx/ingress.conf",
			LineNumber:     88,
			Algorithm:      "TLS 1.2",
			KeySize:        2048,
			RiskLevel:      models.RiskHigh,
			RiskRationale:  "TLS 1.2 ECDHE_RSA key exchange is vulnerable to Harvest-Now-Decrypt-Later quantum attack.",
			PQCStatus:      models.PQCStatusVulnerable,
			PQCReplacement: "TLS 1.3 with X25519MLKEM768",
			Details:        "ssl_protocols TLSv1.2; ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256;",
			CreatedAt:      now,
		},
		{
			ID:             "asset-04",
			ScanID:         demoScanID,
			Name:           "TLS Certificate (RSA 2048)",
			AssetType:      models.AssetCertificate,
			Location:       "certs/api_domain.pem",
			LineNumber:     1,
			Algorithm:      "RSA-2048",
			KeySize:        2048,
			RiskLevel:      models.RiskHigh,
			RiskRationale:  "X.509 RSA public key identity certificate is vulnerable to quantum decryption via Shor's Algorithm.",
			PQCStatus:      models.PQCStatusVulnerable,
			PQCReplacement: "Hybrid ML-DSA-65 + ECDSA P-256",
			Issuer:         "DigiCert Global Root CA",
			ExpiryDate:     &expiryCert,
			Details:        "X.509 Server Identity Certificate expiring 2027.",
			CreatedAt:      now,
		},
		{
			ID:             "asset-05",
			ScanID:         demoScanID,
			Name:           "AES-256-GCM DB Encryption",
			AssetType:      models.AssetAlgorithm,
			Location:       "storage/encrypter.go",
			LineNumber:     102,
			Algorithm:      "AES-256-GCM",
			KeySize:        256,
			RiskLevel:      models.RiskSafe,
			RiskRationale:  "Symmetric 256-bit authenticated encryption provides 128-bit quantum security against Grover's search algorithm.",
			PQCStatus:      models.PQCStatusSafe,
			PQCReplacement: "None Required (Symmetric AES-256 is PQC Safe)",
			Details:        "cipher.NewGCM(block) - Symmetric 256-bit encryption withstands Grover's attack.",
			CreatedAt:      now,
		},
		{
			ID:             "asset-06",
			ScanID:         demoScanID,
			Name:           "PKCS#12 Production Keystore",
			AssetType:      models.AssetKeyStore,
			Location:       "certs/keystore.p12",
			LineNumber:     1,
			Algorithm:      "AES-256-CBC",
			KeySize:        256,
			RiskLevel:      models.RiskMedium,
			RiskRationale:  "Keystore uses legacy PKCS#12 container format.",
			PQCStatus:      models.PQCStatusTransitional,
			PQCReplacement: "QuantumVault Key Vault",
			Details:        "keystore.p12 - Encrypted PKCS#12 certificate and private key store.",
			CreatedAt:      now,
		},
		{
			ID:             "asset-07",
			ScanID:         demoScanID,
			Name:           "OpenSSL 1.1.1 Library",
			AssetType:      models.AssetLibrary,
			Location:       "Dockerfile",
			LineNumber:     12,
			Algorithm:      "OpenSSL",
			KeySize:        0,
			RiskLevel:      models.RiskCritical,
			RiskRationale:  "EOL OpenSSL version lacks provider support for NIST Post-Quantum algorithms.",
			PQCStatus:      models.PQCStatusVulnerable,
			PQCReplacement: "OpenSSL 3.2+ / oqs-provider",
			Details:        "libssl1.1 installed via apt-get - EOL version lacking PQC provider support.",
			CreatedAt:      now,
		},
		{
			ID:             "asset-08",
			ScanID:         demoScanID,
			Name:           "Hardcoded RSA Private Key Secret",
			AssetType:      models.AssetSecret,
			Location:       "config/keys/dev_key.pem",
			LineNumber:     1,
			Algorithm:      "RSA Private Key",
			KeySize:        2048,
			RiskLevel:      models.RiskMedium,
			RiskRationale:  "Hardcoded private key secret exposed directly in source control.",
			PQCStatus:      models.PQCStatusVulnerable,
			PQCReplacement: "QuantumVault Key Vault",
			Details:        "-----BEGIN RSA PRIVATE KEY----- exposed in project repository.",
			CreatedAt:      now,
		},
	}

	r.scans[demoScanID] = summary
	r.assets[demoScanID] = assets

	// Seed Migration Plan
	r.migrationPlans[demoScanID] = models.MigrationPlan{
		ID:               "mig-demo-001",
		ScanID:           demoScanID,
		OverallReadiness: "CRITICAL_NEEDS_MIGRATION",
		EstimatedHours:   45,
		RecommendedPQC:   []string{"ML-DSA-65", "ML-KEM-768", "X25519MLKEM768"},
		Steps: []models.MigrationStep{
			{
				StepNumber:  1,
				Title:       "Upgrade OpenSSL & Crypto Providers",
				Description: "Replace EOL OpenSSL 1.1.1 in Dockerfile with OpenSSL 3.2+ and LibOQS liboqs provider for PQC algorithm support.",
				TargetAsset: "OpenSSL 1.1.1 Library",
				CurrentAlgo: "OpenSSL 1.1.1",
				NewPQCAlgo:  "OpenSSL 3.2 + oqs-provider",
				Effort:      "LOW",
				ActionType:  "REFACTOR_CODE",
				Impact:      []string{"Dockerfile", "CI/CD Pipeline"},
			},
			{
				StepNumber:  2,
				Title:       "Generate QuantumVault ML-DSA Signature Key",
				Description: "Provision a post-quantum ML-DSA-65 digital signature key in QuantumVault KMS to replace legacy RSA-2048.",
				TargetAsset: "RSA-2048 JWT Signature",
				CurrentAlgo: "RS256 (RSA-2048)",
				NewPQCAlgo:  "ML-DSA-65",
				Effort:      "MEDIUM",
				ActionType:  "GENERATE_KEY",
				Impact:      []string{"services/auth/jwt_provider.go"},
			},
			{
				StepNumber:  3,
				Title:       "Issue Hybrid X.509 Identity Certificate",
				Description: "Issue a dual-signed hybrid certificate (ML-DSA-65 + ECDSA P-256) for NGINX ingress endpoints.",
				TargetAsset: "TLS Certificate (RSA 2048)",
				CurrentAlgo: "RSA-2048",
				NewPQCAlgo:  "Hybrid ML-DSA-65 / ECDSA",
				Effort:      "MEDIUM",
				ActionType:  "REISSUE_CERT",
				Impact:      []string{"certs/api_domain.pem", "config/nginx/ingress.conf"},
			},
			{
				StepNumber:  4,
				Title:       "Replace Legacy SHA-1 Digest with SHA-384",
				Description: "Refactor python hash verification script from broken SHA-1 to SHA-384.",
				TargetAsset: "SHA-1 Message Digest",
				CurrentAlgo: "SHA1",
				NewPQCAlgo:  "SHA-384",
				Effort:      "LOW",
				ActionType:  "REFACTOR_CODE",
				Impact:      []string{"utils/hash_checker.py"},
			},
		},
		CreatedAt: now,
	}
}
