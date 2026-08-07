package risk

import (
	"testing"

	"github.com/quantumpq/qv-cbom/internal/models"
)

func TestRiskEngine(t *testing.T) {
	engine := NewRiskEngine()

	assets := []models.CryptoAsset{
		{RiskLevel: models.RiskCritical},
		{RiskLevel: models.RiskHigh},
		{RiskLevel: models.RiskSafe},
	}

	score, critical, high, _, safe := engine.CalculatePQCScore(assets)

	if critical != 1 || high != 1 || safe != 1 {
		t.Errorf("Mismatch in risk counts: critical=%d, high=%d, safe=%d", critical, high, safe)
	}

	if score >= 100 || score <= 0 {
		t.Errorf("PQC score out of bounds: %d", score)
	}

	plan := engine.GenerateMigrationPlan("test-scan", assets)
	if plan.ScanID != "test-scan" {
		t.Errorf("Expected scan ID test-scan, got %s", plan.ScanID)
	}
}
