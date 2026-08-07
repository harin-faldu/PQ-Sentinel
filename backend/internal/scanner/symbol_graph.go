package scanner

import (
	"regexp"
	"strings"
)

// SymbolRef tracks exported and imported symbols across file boundaries
type SymbolRef struct {
	Name       string // Symbol or function name (e.g. "generateRSAKey", "getSecret")
	OriginFile string // File where the symbol is defined
	LineNumber int    // Line number where symbol is defined
	Package    string // Package / Module path
}

// SymbolGraph maintains project-wide symbol definitions and import graphs
type SymbolGraph struct {
	symbols map[string]SymbolRef
}

func NewSymbolGraph() *SymbolGraph {
	return &SymbolGraph{
		symbols: make(map[string]SymbolRef),
	}
}

// RegisterSymbol registers a exported symbol or function definition
func (sg *SymbolGraph) RegisterSymbol(name string, originFile string, lineNo int) {
	if name == "" {
		return
	}
	sg.symbols[name] = SymbolRef{
		Name:       name,
		OriginFile: originFile,
		LineNumber: lineNo,
	}
}

// LookupSymbol finds origin information for a symbol name
func (sg *SymbolGraph) LookupSymbol(name string) (SymbolRef, bool) {
	ref, exists := sg.symbols[name]
	return ref, exists
}

// ParseImportsAndExports parses JS/Python/Go import & export statements
func (sg *SymbolGraph) ParseImportsAndExports(content string, relPath string) {
	lines := strings.Split(content, "\n")
	
	// JS export function / const regex
	jsExportRegex := regexp.MustCompile(`(?:export\s+(?:function|const|let|var)|module\.exports\s*=\s*\{?)\s*([a-zA-Z0-9_]+)`)
	// Python def / class regex
	pyDefRegex := regexp.MustCompile(`def\s+([a-zA-Z0-9_]+)\s*\(`)
	// Go func regex
	goFuncRegex := regexp.MustCompile(`func\s+([A-Z][a-zA-Z0-9_]+)\s*\(`)

	for i, line := range lines {
		lineNo := i + 1
		trimmed := strings.TrimSpace(line)

		if matches := jsExportRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterSymbol(matches[1], relPath, lineNo)
		}
		if matches := pyDefRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterSymbol(matches[1], relPath, lineNo)
		}
		if matches := goFuncRegex.FindStringSubmatch(trimmed); len(matches) > 1 {
			sg.RegisterSymbol(matches[1], relPath, lineNo)
		}
	}
}
