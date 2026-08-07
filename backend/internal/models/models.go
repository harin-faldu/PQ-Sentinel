package models

import (
	"crypto/rand"
	"fmt"
	"time"
)

type RiskLevel string

const (
	RiskCritical RiskLevel = "CRITICAL"
	RiskHigh     RiskLevel = "HIGH"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskLow      RiskLevel = "LOW"
	RiskSafe     RiskLevel = "SAFE"
)

type AssetType string

const (
	AssetAlgorithm   AssetType = "ALGORITHM"
	AssetCertificate AssetType = "CERTIFICATE"
	AssetProtocol    AssetType = "PROTOCOL"
	AssetKeyStore    AssetType = "KEYSTORE"
	AssetLibrary     AssetType = "LIBRARY"
	AssetSecret      AssetType = "SECRET"
	AssetKeyMaterial AssetType = "KEY_MATERIAL"
)

type PQCStatus string

const (
	PQCStatusVulnerable   PQCStatus = "VULNERABLE"
	PQCStatusTransitional PQCStatus = "HYBRID_TRANSITIONAL"
	PQCStatusSafe         PQCStatus = "QUANTUM_SAFE"

	StatusVulnerable   = PQCStatusVulnerable
	StatusTransitional = PQCStatusTransitional
	StatusSafe         = PQCStatusSafe
)

type StatusType = PQCStatus

// CryptoAsset represents a single discovered cryptographic component
type CryptoAsset struct {
	ID               string    `json:"id"`
	ScanID           string    `json:"scan_id"`
	Name             string    `json:"name"`            // e.g., "RSA-2048", "SHA-1", "TLS 1.2", "AES-256-GCM"
	AssetType        AssetType `json:"asset_type"`      // ALGORITHM, CERTIFICATE, PROTOCOL, etc.
	Type             AssetType `json:"type"`            // Alias field for compatibility
	Location         string    `json:"location"`        // File path + line number or network host
	LineNumber       int       `json:"line_number"`     // 1-indexed
	Algorithm        string    `json:"algorithm"`       // e.g., RSA, AES, ECDSA, SHA256
	KeySize          int       `json:"key_size"`        // e.g., 2048, 256, 384
	RiskLevel        RiskLevel `json:"risk_level"`      // CRITICAL, HIGH, MEDIUM, LOW, SAFE
	RiskRationale    string    `json:"risk_rationale"`  // Explanation of WHY it was assigned this risk rating
	PQCStatus        PQCStatus `json:"pqc_status"`      // VULNERABLE, HYBRID_TRANSITIONAL, QUANTUM_SAFE
	PQCReplacement   string    `json:"pqc_replacement"` // e.g., "ML-DSA-65", "ML-KEM-768"
	CNSA20Status     string    `json:"cnsa_2_0_status,omitempty"`   // COMPLIANT, NON_COMPLIANT
	CNSA20Deadline   string    `json:"cnsa_2_0_deadline,omitempty"` // 2025, 2026, 2030
	BusinessScope    string    `json:"business_scope,omitempty"`   // PRODUCTION, STAGING, TEST
	ScopeWeight      float64   `json:"scope_weight,omitempty"`     // 2.0x, 1.0x, 0.5x
	Issuer           string    `json:"issuer,omitempty"`
	ExpiryDate       *time.Time`json:"expiry_date,omitempty"`
	Confidence       float64   `json:"confidence"`              // 0.0-1.0 match confidence
	ContextSnippet   string    `json:"context_snippet,omitempty"`
	Details          string    `json:"details"` // Context snippet or extra metadata
	CreatedAt        time.Time `json:"created_at"`
	DiscoveredAt     time.Time `json:"discovered_at"`
}

// GenerateUUID generates a quick pseudo-random UUID string for assets
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ScanSummary represents the metadata of a CBOM scan run
type ScanSummary struct {
	ID              string    `json:"id"`
	TargetDirectory string    `json:"target_directory"`
	TotalFiles      int       `json:"total_files"`
	ScannedFiles    int       `json:"scanned_files"`
	TotalAssets     int       `json:"total_assets"`
	CriticalCount   int       `json:"critical_count"`
	HighCount       int       `json:"high_count"`
	MediumCount     int       `json:"medium_count"`
	SafeCount       int       `json:"safe_count"`
	PQCScore        int       `json:"pqc_score"`      // 0-100 score indicating quantum readiness
	CNSA20Score     int       `json:"cnsa_2_0_score"` // 0-100 score for NSA CNSA 2.0 readiness
	Status          string    `json:"status"`         // COMPLETED, SCANNING, FAILED
	DurationMs      int64     `json:"duration_ms"`
	CreatedAt       time.Time `json:"created_at"`
}

// ScanDetail includes summary and all discovered assets
type ScanDetail struct {
	Summary ScanSummary   `json:"summary"`
	Assets  []CryptoAsset `json:"assets"`
}

// MigrationStep represents a single actionable step in the PQC migration plan
type MigrationStep struct {
	StepNumber  int      `json:"step_number"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	TargetAsset string   `json:"target_asset"`
	CurrentAlgo string   `json:"current_algo"`
	NewPQCAlgo  string   `json:"new_pqc_algo"`
	Effort      string   `json:"effort"`      // LOW, MEDIUM, HIGH
	ActionType  string   `json:"action_type"` // GENERATE_KEY, REISSUE_CERT, REFACTOR_CODE
	Impact      []string `json:"impact"`
}

// MigrationPlan represents the full post-quantum migration strategy for a scan
type MigrationPlan struct {
	ID               string          `json:"id"`
	ScanID           string          `json:"scan_id"`
	OverallReadiness string          `json:"overall_readiness"` // CRITICAL_NEEDS_MIGRATION, MODERATE, SAFE
	EstimatedHours   int             `json:"estimated_hours"`
	RecommendedPQC   []string        `json:"recommended_pqc"`
	Steps            []MigrationStep `json:"steps"`
	CreatedAt        time.Time       `json:"created_at"`
}

// CycloneDX 1.6 CBOM standard format export structures
type CycloneDXCBOM struct {
	BOMFormat    string               `json:"bomFormat"`
	SpecVersion  string               `json:"specVersion"`
	SerialNumber string               `json:"serialNumber"`
	Version      int                  `json:"version"`
	Metadata     CycloneDXMetadata    `json:"metadata"`
	Components   []CycloneDXComponent `json:"components"`
	Dependencies []CycloneDXDependency`json:"dependencies,omitempty"`
}

type CycloneDXMetadata struct {
	Timestamp string             `json:"timestamp"`
	Tools     []CycloneDXTool    `json:"tools"`
	Component CycloneDXComponent `json:"component"`
}

type CycloneDXTool struct {
	Vendor  string `json:"vendor"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type CycloneDXComponent struct {
	Type             string                `json:"type"`
	Name             string                `json:"name"`
	Version          string                `json:"version,omitempty"`
	PURL             string                `json:"purl,omitempty"`
	CryptoProperties *CycloneDXCryptoProps `json:"cryptoProperties,omitempty"`
}

type CycloneDXCryptoProps struct {
	AssetType           string              `json:"assetType"`
	AlgorithmProperties *CycloneDXAlgoProps `json:"algorithmProperties,omitempty"`
}

type CycloneDXAlgoProps struct {
	Primitive                string `json:"primitive"`
	ParameterSetID           string `json:"parameterSetId,omitempty"`
	ExecutionEnvironment     string `json:"executionEnvironment,omitempty"`
	NISTQuantumSecurityLevel int    `json:"nistQuantumSecurityLevel"`
}

type CycloneDXDependency struct {
	Ref       string   `json:"ref"`
	DependsOn []string `json:"dependsOn"`
}

// ScoreHistoryEntry tracks PQC score over time for trend analysis
type ScoreHistoryEntry struct {
	ScanID    string    `json:"scan_id"`
	Target    string    `json:"target"`
	PQCScore  int       `json:"pqc_score"`
	Critical  int       `json:"critical"`
	High      int       `json:"high"`
	Medium    int       `json:"medium"`
	Safe      int       `json:"safe"`
	Timestamp time.Time `json:"timestamp"`
}
