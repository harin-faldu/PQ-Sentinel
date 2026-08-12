package scanner

// DriftResult captures the alias-based comparison between live AWS KMS keys
// and IaC-declared keys.
type DriftResult struct {
	MatchedAliases  []string         // deployed AND declared — the healthy case
	ConsoleOnlyKeys []LiveKMSFinding // deployed, no matching IaC declaration
	IaCOnlyKeys     []IaCDeclaredKey // declared, not (yet) deployed
}

// DetectKMSDrift matches live findings against IaC declarations by alias name.
func DetectKMSDrift(live []LiveKMSFinding, declared []IaCDeclaredKey) DriftResult {
	declaredByAlias := make(map[string]IaCDeclaredKey)
	for _, d := range declared {
		if d.Alias != "" {
			declaredByAlias[d.Alias] = d
		}
	}

	seenAliases := make(map[string]bool)
	var result DriftResult

	for _, lf := range live {
		if lf.AccessDenied || lf.Alias == "" {
			continue // can't drift-match a key we couldn't read or that has no alias
		}
		if _, ok := declaredByAlias[lf.Alias]; ok {
			result.MatchedAliases = append(result.MatchedAliases, lf.Alias)
			seenAliases[lf.Alias] = true
		} else {
			result.ConsoleOnlyKeys = append(result.ConsoleOnlyKeys, lf)
		}
	}

	for alias, d := range declaredByAlias {
		if !seenAliases[alias] {
			result.IaCOnlyKeys = append(result.IaCOnlyKeys, d)
		}
	}

	return result
}
