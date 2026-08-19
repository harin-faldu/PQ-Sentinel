package models

import (
	"crypto/rand"
	"encoding/xml"
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

// Normalize ensures both AssetType and Type are synchronized and non-empty
func (a *CryptoAsset) Normalize() {
	if a.AssetType == "" && a.Type != "" {
		a.AssetType = a.Type
	} else if a.Type == "" && a.AssetType != "" {
		a.Type = a.AssetType
	}
	if a.AssetType == "" {
		a.AssetType = AssetAlgorithm
		a.Type = AssetAlgorithm
	}
}

// GenerateUUID generates a quick pseudo-random UUID string for assets
func GenerateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

type TargetType string

const (
	TargetTypeRepo         TargetType = "REPO"
	TargetTypeHost         TargetType = "HOST"
	TargetTypeCloudAccount TargetType = "CLOUD_ACCOUNT"
)

type Target struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Type      TargetType `json:"type"`
	Location  string     `json:"location"`
	CreatedAt time.Time  `json:"created_at"`
}

// ScanSummary represents the metadata of a CBOM scan run
type ScanSummary struct {
	ID              string    `json:"id"`
	TargetID        string    `json:"target_id"`
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
	Notice          string    `json:"notice,omitempty"`
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
	XMLName      xml.Name               `xml:"bom" json:"-"`
	XMLNS        string                 `xml:"xmlns,attr" json:"-"`
	BOMFormat    string                 `xml:"-" json:"bomFormat"`
	SpecVersion  string                 `xml:"specVersion,attr,omitempty" json:"specVersion"`
	SerialNumber string                 `xml:"serialNumber,attr,omitempty" json:"serialNumber"`
	Version      int                    `xml:"version,attr,omitempty" json:"version"`
	Metadata     CycloneDXMetadata      `xml:"metadata" json:"metadata"`
	Components   []CycloneDXComponent   `xml:"components>component" json:"components"`
	Dependencies []CycloneDXDependency `xml:"dependencies>dependency,omitempty" json:"dependencies,omitempty"`
}

type CycloneDXMetadata struct {
	Timestamp string               `xml:"timestamp" json:"timestamp"`
	Tools     []CycloneDXTool      `xml:"tools>tool" json:"tools"`
	Component CycloneDXComponent   `xml:"component" json:"component"`
}

type CycloneDXTool struct {
	Vendor  string `xml:"vendor" json:"vendor"`
	Name    string `xml:"name" json:"name"`
	Version string `xml:"version" json:"version"`
}

type CycloneDXComponent struct {
	Type             string                `xml:"type,attr" json:"type"`
	Name             string                `xml:"name" json:"name"`
	Version          string                `xml:"version,omitempty" json:"version,omitempty"`
	PURL             string                `xml:"purl,omitempty" json:"purl,omitempty"`
	EvaluatedAt      string                `xml:"evaluatedAt,omitempty" json:"evaluatedAt,omitempty"`
	CryptoProperties *CycloneDXCryptoProps `xml:"cryptoProperties,omitempty" json:"cryptoProperties,omitempty"`
}

type CycloneDXCryptoProps struct {
	AssetType           string              `xml:"assetType" json:"assetType"`
	AlgorithmProperties *CycloneDXAlgoProps `xml:"algorithmProperties,omitempty" json:"algorithmProperties,omitempty"`
}

type CycloneDXAlgoProps struct {
	Primitive                string `xml:"primitive" json:"primitive"`
	ParameterSetID           string `xml:"parameterSetId,omitempty" json:"parameterSetId,omitempty"`
	ExecutionEnvironment     string `xml:"executionEnvironment,omitempty" json:"executionEnvironment,omitempty"`
	NISTQuantumSecurityLevel int    `xml:"nistQuantumSecurityLevel" json:"nistQuantumSecurityLevel"`
}

type CycloneDXDependency struct {
	Ref       string   `xml:"ref,attr" json:"ref"`
	DependsOn []string `xml:"dependency>ref,omitempty" json:"dependsOn,omitempty"`
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
