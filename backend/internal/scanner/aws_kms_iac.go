package scanner

import "regexp"

var (
	tfKMSKeyBlockRegex   = regexp.MustCompile(`resource\s+"aws_kms_key"\s+"(\w+)"\s*\{`)
	tfKMSAliasBlockRegex = regexp.MustCompile(`resource\s+"aws_kms_alias"\s+"(\w+)"\s*\{`)
	tfAliasNameRegex     = regexp.MustCompile(`name\s*=\s*"([^"]+)"`)
	tfTargetKeyIDRegex   = regexp.MustCompile(`target_key_id\s*=\s*aws_kms_key\.(\w+)\.(id|arn|key_id)`)
	tfIaCRotEnableRegex  = regexp.MustCompile(`enable_key_rotation\s*=\s*(true|false)`)
)

// IaCDeclaredKey represents a KMS key found via Terraform IaC parsing
type IaCDeclaredKey struct {
	LocalName       string // Terraform resource local name, e.g. "payment_key"
	RotationEnabled bool
	RotationStated  bool // whether enable_key_rotation was found at all
	Alias           string
}

// parseTerraformKMSBlocks walks .tf content, extracting both key and alias
// blocks, and resolves each alias's target_key_id reference back to the
// key block's local name — the actual correlation step the old version skipped.
func parseTerraformKMSBlocks(content string) []IaCDeclaredKey {
	keysByLocalName := make(map[string]*IaCDeclaredKey)

	// Pass 1: find every aws_kms_key block and its rotation setting.
	keyBlocks := splitIntoBlocks(content, tfKMSKeyBlockRegex)
	for localName, blockBody := range keyBlocks {
		k := &IaCDeclaredKey{LocalName: localName}
		if m := tfIaCRotEnableRegex.FindStringSubmatch(blockBody); m != nil {
			k.RotationStated = true
			k.RotationEnabled = m[1] == "true"
		}
		keysByLocalName[localName] = k
	}

	// Pass 2: find every aws_kms_alias block, resolve target_key_id back
	// to the key's local name, and attach the alias name to that key.
	aliasBlocks := splitIntoBlocks(content, tfKMSAliasBlockRegex)
	for _, blockBody := range aliasBlocks {
		nameMatch := tfAliasNameRegex.FindStringSubmatch(blockBody)
		targetMatch := tfTargetKeyIDRegex.FindStringSubmatch(blockBody)
		if nameMatch == nil || targetMatch == nil {
			continue
		}
		aliasName := nameMatch[1]
		targetLocalName := targetMatch[1]
		if k, ok := keysByLocalName[targetLocalName]; ok {
			k.Alias = aliasName
		}
	}

	result := make([]IaCDeclaredKey, 0, len(keysByLocalName))
	for _, k := range keysByLocalName {
		result = append(result, *k)
	}
	return result
}

// splitIntoBlocks extracts the body of each resource block keyed by its
// captured local name. Scans forward tracking brace depth until the
// matching closing brace.
func splitIntoBlocks(content string, startPattern *regexp.Regexp) map[string]string {
	blocks := make(map[string]string)
	matches := startPattern.FindAllStringSubmatchIndex(content, -1)
	for _, m := range matches {
		localName := content[m[2]:m[3]]
		braceStart := -1
		// Find the opening brace after the match
		for i := m[1] - 1; i >= m[0]; i-- {
			if content[i] == '{' {
				braceStart = i
				break
			}
		}
		if braceStart < 0 {
			// Brace might be right at match end
			for i := m[1]; i < len(content); i++ {
				if content[i] == '{' {
					braceStart = i
					break
				}
			}
		}
		if braceStart < 0 {
			continue
		}
		depth := 0
		end := braceStart
		for i := braceStart; i < len(content); i++ {
			switch content[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					end = i
					goto done
				}
			}
		}
	done:
		blocks[localName] = content[braceStart : end+1]
	}
	return blocks
}
