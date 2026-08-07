//go:build !cgo

package scanner

import "github.com/quantumpq/qv-cbom/internal/models"

// TreeSitterParser stub for environments without CGO enabled.
// Seamlessly falls back to the enhanced line-based AST parser.
type TreeSitterParser struct {
	rules *RulesConfig
}

func NewTreeSitterParser(rules *RulesConfig) *TreeSitterParser {
	return &TreeSitterParser{rules: rules}
}

func (tsp *TreeSitterParser) ParseCodeWithTreeSitter(content []byte, relPath string, scanID string) ([]models.CryptoAsset, bool) {
	return nil, false // Fall back to enhanced line AST parser
}
