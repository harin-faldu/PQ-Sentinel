//go:build cgo

package scanner

import (
	"context"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	"github.com/quantumpq/qv-cbom/internal/models"
)

// TreeSitterParser parses source files using concrete Tree-Sitter AST syntax trees
type TreeSitterParser struct {
	rules *RulesConfig
}

func NewTreeSitterParser(rules *RulesConfig) *TreeSitterParser {
	return &TreeSitterParser{rules: rules}
}

// GetLanguageForFile maps file extensions to Tree-Sitter language grammars
func GetLanguageForFile(ext string) *sitter.Language {
	switch ext {
	case ".js", ".mjs", ".cjs":
		return javascript.GetLanguage()
	case ".ts":
		return typescript.GetLanguage()
	case ".py":
		return python.GetLanguage()
	case ".go":
		return golang.GetLanguage()
	case ".java":
		return java.GetLanguage()
	default:
		return nil
	}
}

// ParseCodeWithTreeSitter parses code using a real AST syntax tree.
// Returns (assets, true) on successful AST match evaluation, or (nil, false) to fall back to line parser.
func (tsp *TreeSitterParser) ParseCodeWithTreeSitter(content []byte, relPath string, scanID string) (assets []models.CryptoAsset, supported bool) {
	ext := strings.ToLower(filepath.Ext(relPath))
	lang := GetLanguageForFile(ext)
	if lang == nil {
		return nil, false // Not a Tree-sitter supported language -> fallback
	}

	// Graceful recovery against panics on malformed/minified code
	defer func() {
		if r := recover(); r != nil {
			assets = nil
			supported = false
		}
	}()

	parser := sitter.NewParser()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil || tree == nil || tree.RootNode() == nil {
		return nil, false
	}

	tsp.walkASTNode(tree.RootNode(), content, relPath, scanID, ext, &assets)
	return assets, true
}

// walkASTNode walks syntax tree looking for call_expression nodes strictly (ignoring string literals and comments)
func (tsp *TreeSitterParser) walkASTNode(node *sitter.Node, source []byte, relPath string, scanID string, ext string, assets *[]models.CryptoAsset) {
	if node == nil {
		return
	}

	nodeType := node.Type()

	// Only process call expressions (ignore comments, raw string literals, plain identifiers)
	if nodeType == "call_expression" {
		fnNode := node.ChildByFieldName("function")
		if fnNode != nil {
			callText := fnNode.Content(source)
			callSnippet := node.Content(source)
			lineNo := int(node.StartPoint().Row) + 1

			if tsp.rules != nil {
				for _, rule := range tsp.rules.SourceCodeRules {
					if matchLanguage(ext, rule.Languages) && rule.CompiledRegex != nil {
						if rule.CompiledRegex.MatchString(callText) || rule.CompiledRegex.MatchString(callSnippet) {
							// Multi-line AST argument node walker for key size parameters
							keySize := rule.KeySize
							argsNode := node.ChildByFieldName("arguments")
							if argsNode != nil {
								if extracted := extractModulusLengthFromASTNode(argsNode, source); extracted > 0 {
									keySize = extracted
								}
							}

							*assets = append(*assets, models.CryptoAsset{
								ID:             models.GenerateUUID(),
								ScanID:         scanID,
								Name:           rule.Name,
								AssetType:      models.AssetAlgorithm,
								Type:           models.AssetAlgorithm,
								Location:       relPath,
								LineNumber:     lineNo,
								Algorithm:      rule.Algorithm,
								KeySize:        keySize,
								RiskLevel:      models.RiskLevel(rule.RiskLevel),
								RiskRationale:  rule.RiskRationale,
								PQCStatus:      models.StatusType(rule.PQCStatus),
								PQCReplacement: rule.PQCReplacement,
								Confidence:     1.0, // 100% Structural AST Certainty
								ContextSnippet: callSnippet,
								DiscoveredAt:   time.Now(),
							})
						}
					}
				}
			}
		}
	}

	// Walk all child AST nodes
	for i := 0; i < int(node.ChildCount()); i++ {
		tsp.walkASTNode(node.Child(i), source, relPath, scanID, ext, assets)
	}
}

// extractModulusLengthFromASTNode walks argument node children across line boundaries to extract key_size/modulusLength
func extractModulusLengthFromASTNode(argsNode *sitter.Node, source []byte) int {
	if argsNode == nil {
		return 0
	}

	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child == nil {
			continue
		}

		if child.Type() == "object" || child.Type() == "dictionary" || child.Type() == "argument_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				pair := child.Child(j)
				if pair != nil && (pair.Type() == "pair" || pair.Type() == "property") {
					keyNode := pair.ChildByFieldName("key")
					valNode := pair.ChildByFieldName("value")

					if keyNode == nil {
						keyNode = pair.Child(0)
					}
					if valNode == nil && pair.ChildCount() > 1 {
						valNode = pair.Child(pair.ChildCount() - 1)
					}

					if keyNode != nil && valNode != nil {
						keyName := strings.TrimSpace(keyNode.Content(source))
						if keyName == "modulusLength" || keyName == "key_size" || keyName == "rsa_bits" {
							valStr := strings.TrimSpace(valNode.Content(source))
							if size, err := strconv.Atoi(valStr); err == nil {
								return size
							}
						}
					}
				}
			}
		}
	}
	return 0
}
