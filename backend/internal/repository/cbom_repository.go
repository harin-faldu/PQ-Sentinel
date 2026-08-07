package repository

import (
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type CBOMRepository interface {
	SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error
	GetScans() ([]models.ScanSummary, error)
	GetScanByID(id string) (*models.ScanDetail, error)
	SaveMigrationPlan(plan models.MigrationPlan) error
	GetMigrationPlanByScanID(scanID string) (*models.MigrationPlan, error)
}

type MemoryRepository struct {
	mu             sync.RWMutex
	scans          map[string]models.ScanSummary
	assets         map[string][]models.CryptoAsset
	migrationPlans map[string]models.MigrationPlan
}

func NewMemoryRepository() *MemoryRepository {
	repo := &MemoryRepository{
		scans:          make(map[string]models.ScanSummary),
		assets:         make(map[string][]models.CryptoAsset),
		migrationPlans: make(map[string]models.MigrationPlan),
	}
	repo.seedInitialData()
	return repo
}

func (r *MemoryRepository) SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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

// Seed initial realistic demonstration data so the UI displays rich results out-of-the-box
func (r *MemoryRepository) seedInitialData() {
	now := time.Now()
	expiryCert := now.AddDate(1, 2, 0)

	demoScanID := "scan-demo-001"
	summary := models.ScanSummary{
		ID:              demoScanID,
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
			Name:           "AWS KMS CMK Key",
			AssetType:      models.AssetKeyStore,
			Location:       "cloud/aws_kms_config.tf",
			LineNumber:     14,
			Algorithm:      "SYMMETRIC_DEFAULT",
			KeySize:        256,
			RiskLevel:      models.RiskMedium,
			RiskRationale:  "KMS key requires hybrid post-quantum TLS transport configuration.",
			PQCStatus:      models.PQCStatusTransitional,
			PQCReplacement: "AWS KMS Post-Quantum Hybrid TLS",
			Details:        "aws_kms_key.payment_key - AWS KMS Key with automatic rotation enabled.",
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

// MySQLRepository implementer for production MySQL databases
type MySQLRepository struct {
	db *sql.DB
}

func NewMySQLRepository(dbSource string) (*MySQLRepository, error) {
	db, err := sql.Open("mysql", dbSource)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("mysql connection failed: %w", err)
	}

	repo := &MySQLRepository{db: db}
	if err := repo.createTables(); err != nil {
		log.Printf("Warning: MySQL table creation failed: %v", err)
	}
	return repo, nil
}

func (r *MySQLRepository) createTables() error {
	scansTable := `
	CREATE TABLE IF NOT EXISTS scans (
		id VARCHAR(64) PRIMARY KEY,
		target_directory VARCHAR(255),
		total_files INT,
		scanned_files INT,
		total_assets INT,
		critical_count INT,
		high_count INT,
		medium_count INT,
		safe_count INT,
		pqc_score INT,
		status VARCHAR(32),
		duration_ms BIGINT,
		created_at DATETIME
	);`

	assetsTable := `
	CREATE TABLE IF NOT EXISTS crypto_assets (
		id VARCHAR(64) PRIMARY KEY,
		scan_id VARCHAR(64),
		name VARCHAR(255),
		asset_type VARCHAR(64),
		location VARCHAR(255),
		line_number INT,
		algorithm VARCHAR(64),
		key_size INT,
		risk_level VARCHAR(32),
		risk_rationale TEXT,
		pqc_status VARCHAR(32),
		pqc_replacement VARCHAR(128),
		issuer VARCHAR(255),
		details TEXT,
		created_at DATETIME,
		FOREIGN KEY (scan_id) REFERENCES scans(id) ON DELETE CASCADE
	);`

	_, err1 := r.db.Exec(scansTable)
	_, err2 := r.db.Exec(assetsTable)
	if err1 != nil {
		return err1
	}
	return err2
}

func (r *MySQLRepository) SaveScan(summary models.ScanSummary, assets []models.CryptoAsset) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryScan := `INSERT INTO scans (id, target_directory, total_files, scanned_files, total_assets, critical_count, high_count, medium_count, safe_count, pqc_score, status, duration_ms, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	_, err = tx.Exec(queryScan, summary.ID, summary.TargetDirectory, summary.TotalFiles, summary.ScannedFiles, summary.TotalAssets, summary.CriticalCount, summary.HighCount, summary.MediumCount, summary.SafeCount, summary.PQCScore, summary.Status, summary.DurationMs, summary.CreatedAt)
	if err != nil {
		return err
	}

	queryAsset := `INSERT INTO crypto_assets (id, scan_id, name, asset_type, location, line_number, algorithm, key_size, risk_level, risk_rationale, pqc_status, pqc_replacement, issuer, details, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	for _, a := range assets {
		_, err = tx.Exec(queryAsset, a.ID, a.ScanID, a.Name, a.AssetType, a.Location, a.LineNumber, a.Algorithm, a.KeySize, a.RiskLevel, a.RiskRationale, a.PQCStatus, a.PQCReplacement, a.Issuer, a.Details, a.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *MySQLRepository) GetScans() ([]models.ScanSummary, error) {
	rows, err := r.db.Query(`SELECT id, target_directory, total_files, scanned_files, total_assets, critical_count, high_count, medium_count, safe_count, pqc_score, status, duration_ms, created_at FROM scans ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scans []models.ScanSummary
	for rows.Next() {
		var s models.ScanSummary
		if err := rows.Scan(&s.ID, &s.TargetDirectory, &s.TotalFiles, &s.ScannedFiles, &s.TotalAssets, &s.CriticalCount, &s.HighCount, &s.MediumCount, &s.SafeCount, &s.PQCScore, &s.Status, &s.DurationMs, &s.CreatedAt); err != nil {
			return nil, err
		}
		scans = append(scans, s)
	}
	return scans, nil
}

func (r *MySQLRepository) GetScanByID(id string) (*models.ScanDetail, error) {
	var s models.ScanSummary
	err := r.db.QueryRow(`SELECT id, target_directory, total_files, scanned_files, total_assets, critical_count, high_count, medium_count, safe_count, pqc_score, status, duration_ms, created_at FROM scans WHERE id = ?`, id).Scan(
		&s.ID, &s.TargetDirectory, &s.TotalFiles, &s.ScannedFiles, &s.TotalAssets, &s.CriticalCount, &s.HighCount, &s.MediumCount, &s.SafeCount, &s.PQCScore, &s.Status, &s.DurationMs, &s.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(`SELECT id, scan_id, name, asset_type, location, line_number, algorithm, key_size, risk_level, risk_rationale, pqc_status, pqc_replacement, issuer, details, created_at FROM crypto_assets WHERE scan_id = ?`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []models.CryptoAsset
	for rows.Next() {
		var a models.CryptoAsset
		if err := rows.Scan(&a.ID, &a.ScanID, &a.Name, &a.AssetType, &a.Location, &a.LineNumber, &a.Algorithm, &a.KeySize, &a.RiskLevel, &a.RiskRationale, &a.PQCStatus, &a.PQCReplacement, &a.Issuer, &a.Details, &a.CreatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}

	return &models.ScanDetail{Summary: s, Assets: assets}, nil
}

func (r *MySQLRepository) SaveMigrationPlan(plan models.MigrationPlan) error {
	return nil
}

func (r *MySQLRepository) GetMigrationPlanByScanID(scanID string) (*models.MigrationPlan, error) {
	return nil, fmt.Errorf("migration plan not found")
}
