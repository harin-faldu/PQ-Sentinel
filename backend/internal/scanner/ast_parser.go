package scanner

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/quantumpq/qv-cbom/internal/models"
)

// ASTNode represents a simplified Abstract Syntax Tree node for code parsing
type ASTNode struct {
	Type     string            // "call_expression", "comment", "executable_statement"
	Value    string            // Exact code snippet or token string
	Line     int               // 1-indexed line number
	Imports  map[string]string // Module import map (alias -> package)
	Children []*ASTNode        // Nested child AST nodes
}

// ASTParser handles multi-language syntax tree parsing, comment filtering, and import state tracking
type ASTParser struct {
	inBlockComment bool
}

func NewASTParser() *ASTParser {
	return &ASTParser{inBlockComment: false}
}

// ResetState resets parser state between files
func (p *ASTParser) ResetState() {
	p.inBlockComment = false
}

// IsComment checks if a line or token is a code comment (SL or ML)
func (p *ASTParser) IsComment(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Block comment toggle check
	if p.inBlockComment {
		if strings.Contains(trimmed, "*/") {
			p.inBlockComment = false
		}
		return true
	}

	if strings.HasPrefix(trimmed, "/*") {
		if !strings.Contains(trimmed, "*/") {
			p.inBlockComment = true
		}
		return true
	}

	// JavaScript/Go/Java single-line comments
	if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "*") {
		return true
	}
	// Python / YAML / Shell comments
	if strings.HasPrefix(trimmed, "#") {
		return true
	}
	return false
}

// ParseLineToNode parses a single code line into AST Node representation
func (p *ASTParser) ParseLineToNode(line string, lineNumber int) *ASTNode {
	trimmed := strings.TrimSpace(line)
	if p.IsComment(trimmed) {
		return &ASTNode{
			Type:  "comment",
			Value: trimmed,
			Line:  lineNumber,
		}
	}

	return &ASTNode{
		Type:  "executable_statement",
		Value: trimmed,
		Line:  lineNumber,
	}
}

// MatchYAMLRuleAgainstAST evaluates a YAML scanner rule against an AST Node with context-window key size extraction
func (p *ASTParser) MatchYAMLRuleAgainstAST(node *ASTNode, rule YAMLRule, scanID string, relPath string, contextLines []string) *models.CryptoAsset {
	if node.Type == "comment" || isNonExecutableStatement(node.Value) {
		return nil
	}

	if rule.CompiledRegex != nil && rule.CompiledRegex.MatchString(node.Value) {
		assetType := models.AssetAlgorithm
		if strings.Contains(strings.ToLower(rule.Name), "cert") {
			assetType = models.AssetCertificate
		} else if strings.Contains(strings.ToLower(rule.Name), "key") || strings.Contains(strings.ToLower(rule.Name), "secret") {
			assetType = models.AssetKeyMaterial
		} else if strings.Contains(strings.ToLower(rule.Name), "tls") || strings.Contains(strings.ToLower(rule.Name), "protocol") {
			assetType = models.AssetProtocol
		}

		// Dynamic key size extraction from context window
		keySize := rule.KeySize
		if rule.CompiledCapture != nil && len(contextLines) > 0 {
			searchBlock := strings.Join(contextLines, " ")
			if captures := rule.CompiledCapture.FindStringSubmatch(searchBlock); len(captures) > 1 {
				if extracted, err := strconv.Atoi(captures[1]); err == nil {
					keySize = extracted
				}
			}
		}

		// Confidence scoring
		confidence := rule.Confidence
		if confidence == 0 {
			confidence = 1.0
		}

		return &models.CryptoAsset{
			ID:             models.GenerateUUID(),
			ScanID:         scanID,
			Name:           rule.Name,
			AssetType:      assetType,
			Type:           assetType,
			Location:       relPath,
			LineNumber:     node.Line,
			Algorithm:      rule.Algorithm,
			KeySize:        keySize,
			RiskLevel:      models.RiskLevel(rule.RiskLevel),
			RiskRationale:  rule.RiskRationale,
			PQCStatus:      models.StatusType(rule.PQCStatus),
			PQCReplacement: rule.PQCReplacement,
			Confidence:     confidence,
			ContextSnippet: node.Value,
		}
	}

	return nil
}

// MatchRuleAgainstAST evaluates legacy rule against AST Node (backward compatibility)
func (p *ASTParser) MatchRuleAgainstAST(node *ASTNode, pattern *regexp.Regexp, name string, relPath string) *models.CryptoAsset {
	if node.Type == "comment" {
		return nil
	}

	if pattern != nil && pattern.MatchString(node.Value) {
		return &models.CryptoAsset{
			ID:             models.GenerateUUID(),
			Name:           name,
			AssetType:      models.AssetAlgorithm,
			Type:           models.AssetAlgorithm,
			Location:       relPath,
			LineNumber:     node.Line,
			ContextSnippet: node.Value,
		}
	}

	return nil
}

// isNonExecutableStatement checks if a statement is a log message, print line, or pure control flow check
func isNonExecutableStatement(val string) bool {
	if isPureControlFlowCheck(val) {
		return true
	}

	trimmed := strings.TrimSpace(val)
	lower := strings.ToLower(trimmed)

	// Filter log/print statements (e.g. console.log("..."), logger.info("..."))
	logPrefixes := []string{"console.", "logger.", "log.", "fmt.print", "system.out.", "t.log", "t.error", "print(", "println("}
	for _, lp := range logPrefixes {
		if strings.HasPrefix(lower, lp) {
			return true
		}
	}

	// Filter string assignments (e.g. const info = "crypto.createCipheriv...")
	if (strings.HasPrefix(lower, "const ") || strings.HasPrefix(lower, "let ") || strings.HasPrefix(lower, "var ")) &&
		(strings.Contains(lower, "= \"") || strings.Contains(lower, "= '") || strings.Contains(lower, "= `")) {
		hasActiveCall := strings.Contains(lower, "= crypto.") || strings.Contains(lower, "= new ") || strings.Contains(lower, "= require(") || strings.Contains(lower, "= import(")
		if !hasActiveCall {
			return true
		}
	}

	return false
}

// isPureControlFlowCheck returns true if a line is a conditional statement comparing strings without active crypto invocations
func isPureControlFlowCheck(val string) bool {
	trimmed := strings.TrimSpace(val)
	if strings.HasPrefix(trimmed, "if ") || strings.HasPrefix(trimmed, "if(") ||
		strings.HasPrefix(trimmed, "else if") || strings.HasPrefix(trimmed, "switch") ||
		strings.HasPrefix(trimmed, "case ") {
		lower := strings.ToLower(trimmed)
		hasComparison := strings.Contains(lower, "==") || strings.Contains(lower, "!=") || strings.Contains(lower, "case ") || strings.Contains(lower, ".includes(")
		hasCryptoCall := strings.Contains(lower, "create") || strings.Contains(lower, "generate") ||
			strings.Contains(lower, "new ") || strings.Contains(lower, "sign(") ||
			strings.Contains(lower, "encrypt(") || strings.Contains(lower, "decrypt(") ||
			strings.Contains(lower, "hash(") || strings.Contains(lower, "digest(") ||
			strings.Contains(lower, "require(") || strings.Contains(lower, "import")
		if hasComparison && !hasCryptoCall {
			return true
		}
	}
	return false
}
