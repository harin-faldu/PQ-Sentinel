package scanner

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/quantumpq/qv-cbom/internal/models"
)

// CallSite tracks an active consumer invocation of an imported/exported symbol
type CallSite struct {
	SymbolName   string `json:"symbol_name"`
	ConsumerFile string `json:"consumer_file"`
	LineNumber   int    `json:"line_number"`
}

// SymbolRef tracks exported and imported symbols across file boundaries
type SymbolRef struct {
	Name       string     `json:"name"`
	OriginFile string     `json:"origin_file"`
	LineNumber int        `json:"line_number"`
	Package    string     `json:"package"`
	CallSites  []CallSite `json:"call_sites,omitempty"`
}

// SymbolGraph maintains project-wide symbol definitions, imports, and call-site contexts
type SymbolGraph struct {
	mu        sync.RWMutex
	symbols   map[string]SymbolRef
	consumers map[string][]CallSite // symbolName -> list of consumer call-sites across files
}

func NewSymbolGraph() *SymbolGraph {
	return &SymbolGraph{
		symbols:   make(map[string]SymbolRef),
		consumers: make(map[string][]CallSite),
	}
}

// RegisterSymbol registers an exported symbol or function definition
func (sg *SymbolGraph) RegisterSymbol(name string, originFile string, lineNo int) {
	if name == "" {
		return
	}
	sg.mu.Lock()
	defer sg.mu.Unlock()

	sg.symbols[name] = SymbolRef{
		Name:       name,
		OriginFile: originFile,
		LineNumber: lineNo,
	}
}

// RegisterConsumerCallSite registers a consumer call-site using an imported/exported symbol
func (sg *SymbolGraph) RegisterConsumerCallSite(name string, consumerFile string, lineNo int) {
	if name == "" || consumerFile == "" {
		return
	}
	sg.mu.Lock()
	defer sg.mu.Unlock()

	site := CallSite{
		SymbolName:   name,
		ConsumerFile: consumerFile,
		LineNumber:   lineNo,
	}

	// Avoid duplicate call-sites for the exact same file & line
	existing := sg.consumers[name]
	for _, s := range existing {
		if s.ConsumerFile == consumerFile && s.LineNumber == lineNo {
			return
		}
	}
	sg.consumers[name] = append(existing, site)
}

// LookupSymbol finds origin information and aggregated call-sites for a symbol name
func (sg *SymbolGraph) LookupSymbol(name string) (SymbolRef, bool) {
	sg.mu.RLock()
	defer sg.mu.RUnlock()

	ref, exists := sg.symbols[name]
	if exists {
		ref.CallSites = sg.consumers[name]
	}
	return ref, exists
}

// GetCallSites returns all aggregated call-sites for a given symbol
func (sg *SymbolGraph) GetCallSites(name string) []CallSite {
	sg.mu.RLock()
	defer sg.mu.RUnlock()
	return sg.consumers[name]
}

// ParseImportsAndExports parses JS/TS/Python/Go/Java import & export statements and function calls
func (sg *SymbolGraph) ParseImportsAndExports(content string, relPath string) {
	lines := strings.Split(content, "\n")

	// Export Regexes
	jsExportRegex := regexp.MustCompile(`(?:export\s+(?:function|const|let|var)|module\.exports\s*=\s*\{?)\s*([a-zA-Z0-9_]+)`)
	pyDefRegex := regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\(`)
	goFuncRegex := regexp.MustCompile(`func\s+([A-Z][a-zA-Z0-9_]+)\s*\(`)

	// Import & Consumer Call-Site Regexes
	jsImportRegex := regexp.MustCompile(`(?:import\s+\{?([^}]+)\}?\s+from|const\s+\{?([^}=]+)\}?\s*=\s*require)`)
	pyImportRegex := regexp.MustCompile(`(?:from\s+[a-zA-Z0-9_\.]+\s+import\s+([a-zA-Z0-9_,\s]+)|import\s+([a-zA-Z0-9_]+))`)
	goImportRegex := regexp.MustCompile(`import\s+["']([^"']+)["']`)
	javaImportRegex := regexp.MustCompile(`import\s+static?\s*([a-zA-Z0-9_\.]+);`)

	for i, line := range lines {
		lineNo := i + 1
		trimmed := strings.TrimSpace(line)

		// Parse Exports
		if matches := jsExportRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterSymbol(matches[1], relPath, lineNo)
		}
		if matches := pyDefRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterSymbol(matches[1], relPath, lineNo)
		}
		if matches := goFuncRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterSymbol(matches[1], relPath, lineNo)
		}

		// Parse Imports & Consumer Usage Call-Sites
		if matches := jsImportRegex.FindStringSubmatch(trimmed); len(matches) > 0 {
			rawImports := matches[1]
			if rawImports == "" && len(matches) > 2 {
				rawImports = matches[2]
			}
			for _, imp := range strings.Split(rawImports, ",") {
				sym := strings.TrimSpace(imp)
				if sym != "" {
					sg.RegisterConsumerCallSite(sym, relPath, lineNo)
				}
			}
		}
		if matches := pyImportRegex.FindStringSubmatch(trimmed); len(matches) > 0 {
			rawImports := matches[1]
			if rawImports == "" && len(matches) > 2 {
				rawImports = matches[2]
			}
			for _, imp := range strings.Split(rawImports, ",") {
				sym := strings.TrimSpace(imp)
				if sym != "" {
					sg.RegisterConsumerCallSite(sym, relPath, lineNo)
				}
			}
		}
		if matches := goImportRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterConsumerCallSite(matches[1], relPath, lineNo)
		}
		if matches := javaImportRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterConsumerCallSite(matches[1], relPath, lineNo)
		}
	}
}

// AttachCallSiteContext attaches aggregated call-site consumers and evaluates max ScopeWeight
func (sg *SymbolGraph) AttachCallSiteContext(symbolName string, asset *models.CryptoAsset) {
	callSites := sg.GetCallSites(symbolName)
	if len(callSites) == 0 {
		return
	}

	var callSiteStrs []string
	maxWeight := asset.ScopeWeight

	for _, cs := range callSites {
		callSiteStrs = append(callSiteStrs, fmt.Sprintf("%s:%d", cs.ConsumerFile, cs.LineNumber))

		// Check if any consumer call-site is in a PRODUCTION entry point
		scope, weight := classifyBusinessScope(cs.ConsumerFile)
		if weight > maxWeight {
			maxWeight = weight
			asset.BusinessScope = scope
			asset.ScopeWeight = weight
		}
	}

	// Attach aggregated call-site consumers into Asset Details (Zero Duplicate Assets)
	asset.Details = fmt.Sprintf("%s | Active Consumer Call-Sites (%d): [%s]",
		asset.Details, len(callSiteStrs), strings.Join(callSiteStrs, ", "))
}
