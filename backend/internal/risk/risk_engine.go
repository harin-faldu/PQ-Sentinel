package risk

import (
	"fmt"

	"github.com/quantumpq/qv-cbom/internal/models"
)

type RiskEngine struct{}

func NewRiskEngine() *RiskEngine {
	return &RiskEngine{}
}

// CalculatePQCScore computes a weighted readiness score (0-100) using scope-aware normalization
// and populates NSA CNSA 2.0 compliance metadata on each asset.
//
// New formula: score = 100 * (1 - totalWeightedPenalty / totalWeight)
// where penalty per asset = riskFactor * scopeWeight
// Critical=1.0, High=0.7, Medium=0.3, Safe/Low=0.0
func (r *RiskEngine) CalculatePQCScore(assets []models.CryptoAsset) (int, int, int, int, int) {
	if len(assets) == 0 {
		return 100, 0, 0, 0, 0
	}

	critical, high, medium, safe := 0, 0, 0, 0
	totalWeightedPenalty := 0.0
	totalWeight := 0.0

	for i := range assets {
		a := &assets[i]

		// Get scope weight (default 1.0 if not populated)
		weight := a.ScopeWeight
		if weight == 0 {
			weight = 1.0
		}
		totalWeight += weight

		switch a.RiskLevel {
		case models.RiskCritical:
			critical++
			totalWeightedPenalty += 1.0 * weight // 100% penalty
		case models.RiskHigh:
			high++
			totalWeightedPenalty += 0.7 * weight // 70% penalty
		case models.RiskMedium:
			medium++
			totalWeightedPenalty += 0.3 * weight // 30% penalty
		case models.RiskSafe, models.RiskLow:
			safe++
			// No penalty
		}

		// Annotate NSA CNSA 2.0 Timelines (2025/2026/2030 deadlines)
		if a.PQCStatus == models.PQCStatusSafe {
			a.CNSA20Status = "COMPLIANT"
			a.CNSA20Deadline = "2025"
		} else {
			a.CNSA20Status = "NON_COMPLIANT"
			if a.AssetType == models.AssetProtocol {
				a.CNSA20Deadline = "2026 (Network Equipment Deadline)"
			} else if a.AssetType == models.AssetSecret || a.AssetType == models.AssetKeyMaterial {
				a.CNSA20Deadline = "2025 (Software & Web Services Deadline)"
			} else {
				a.CNSA20Deadline = "2030 (Mandatory Legacy Sunset)"
			}
		}
	}

	// Normalized weighted score: penalizes risk proportional to scope importance
	score := 100.0 * (1.0 - totalWeightedPenalty/totalWeight)

	if score < 0 {
		score = 0
	} else if score > 100 {
		score = 100
	}

	return int(score), critical, high, medium, safe
}

// GenerateMigrationPlan constructs step-by-step PQC remediation playbooks
func (r *RiskEngine) GenerateMigrationPlan(scanID string, assets []models.CryptoAsset) models.MigrationPlan {
	var steps []models.MigrationStep
	stepNum := 1

	for _, a := range assets {
		if a.PQCStatus == models.PQCStatusVulnerable {
			actionType := "REFACTOR_CODE"
			effort := "MEDIUM"

			if a.AssetType == models.AssetCertificate {
				actionType = "REISSUE_CERT"
				effort = "MEDIUM"
			} else if a.AssetType == models.AssetKeyStore || a.AssetType == models.AssetSecret || a.AssetType == models.AssetKeyMaterial {
				actionType = "GENERATE_KEY"
				effort = "HIGH"
			}

			step := models.MigrationStep{
				StepNumber:  stepNum,
				Title:       fmt.Sprintf("Migrate %s to Quantum-Safe %s (CNSA 2.0 Target: %s)", a.Name, a.PQCReplacement, a.CNSA20Deadline),
				Description: fmt.Sprintf("Replace legacy %s algorithm at %s with NIST PQC standard %s. NSA CNSA 2.0 Compliance: %s.", a.Algorithm, a.Location, a.PQCReplacement, a.CNSA20Deadline),
				TargetAsset: a.Name,
				CurrentAlgo: a.Algorithm,
				NewPQCAlgo:  a.PQCReplacement,
				Effort:      effort,
				ActionType:  actionType,
				Impact:      []string{a.Location},
			}
			steps = append(steps, step)
			stepNum++
		}
	}

	readiness := "SAFE"
	if len(steps) > 4 {
		readiness = "CRITICAL_NEEDS_MIGRATION"
	} else if len(steps) > 0 {
		readiness = "MODERATE_RISK"
	}

	return models.MigrationPlan{
		ID:               fmt.Sprintf("plan-%s", scanID),
		ScanID:           scanID,
		OverallReadiness: readiness,
		EstimatedHours:   len(steps) * 12,
		RecommendedPQC:   []string{"ML-DSA-65", "ML-KEM-768", "X25519MLKEM768"},
		Steps:            steps,
	}
}
