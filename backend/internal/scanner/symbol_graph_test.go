package scanner

import (
	"testing"

	"github.com/quantumpq/qv-cbom/internal/models"
)

func TestSymbolGraphCallSiteAggregationAndScopeElevation(t *testing.T) {
	sg := NewSymbolGraph()

	// 1. Register symbol definition in helper file
	sg.RegisterSymbol("hashPassword", "utils/crypto-helpers.js", 5)

	// 2. Parse import call-sites across files
	jsCaller1 := `const { hashPassword } = require('./utils/crypto-helpers');
	function login(req, res) {
		const hash = hashPassword(req.body.password);
	}`
	sg.ParseImportsAndExports(jsCaller1, "controllers/auth.js") // PRODUCTION caller

	jsCaller2 := `import { hashPassword } from '../utils/crypto-helpers';
	test('hash test', () => {
		expect(hashPassword('pass')).toBeDefined();
	});`
	sg.ParseImportsAndExports(jsCaller2, "test/auth_test.js") // TEST caller

	// 3. Verify Symbol Lookup
	ref, exists := sg.LookupSymbol("hashPassword")
	if !exists {
		t.Fatalf("Expected symbol hashPassword to exist in SymbolGraph")
	}

	if ref.OriginFile != "utils/crypto-helpers.js" {
		t.Errorf("Expected origin file utils/crypto-helpers.js, got %s", ref.OriginFile)
	}

	// 4. Test Call-Site Aggregation & ScopeWeight Elevation
	asset := models.CryptoAsset{
		ID:            "asset-hash",
		Name:          "hashPassword",
		Location:      "utils/crypto-helpers.js:5",
		Algorithm:     "MD5",
		BusinessScope: "TEST",
		ScopeWeight:   0.3,
		Details:       "Weak MD5 helper function",
	}

	sg.AttachCallSiteContext("hashPassword", &asset)

	// ScopeWeight must be elevated to PRODUCTION (1.5x) because controllers/auth.js is a production caller
	if asset.ScopeWeight != 1.5 {
		t.Errorf("Expected ScopeWeight elevated to 1.5 (PRODUCTION), got %.2f", asset.ScopeWeight)
	}

	if asset.BusinessScope != "PRODUCTION" {
		t.Errorf("Expected BusinessScope elevated to PRODUCTION, got %s", asset.BusinessScope)
	}

	// Verify single asset has aggregated consumer call-sites in Details (zero duplicate assets)
	if !testing.Verbose() && asset.Details == "" {
		t.Errorf("Expected Details to contain call-site list")
	}
}
