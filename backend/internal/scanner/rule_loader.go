package scanner

import (
	"fmt"
	"os"
	"regexp"

	"gopkg.in/yaml.v3"
)

type YAMLRule struct {
	ID             string   `yaml:"id"`
	Name           string   `yaml:"name"`
	Pattern        string   `yaml:"pattern"`
	Languages      []string `yaml:"languages"`
	Algorithm      string   `yaml:"algorithm"`
	KeySize        int      `yaml:"key_size"`
	KeySizeCapture string   `yaml:"key_size_capture"` // Regex with capture group to extract actual key size from code
	RiskLevel      string   `yaml:"risk_level"`
	RiskRationale  string   `yaml:"risk_rationale"`
	PQCStatus      string   `yaml:"pqc_status"`
	PQCReplacement string   `yaml:"pqc_replacement"`
	Confidence     float64  `yaml:"confidence"` // 0.0-1.0 match confidence (default 1.0)

	CompiledRegex   *regexp.Regexp `yaml:"-"`
	CompiledCapture *regexp.Regexp `yaml:"-"` // Compiled key_size_capture regex
}

type RulesConfig struct {
	Version         string     `yaml:"version"`
	SourceCodeRules []YAMLRule `yaml:"source_code_rules"`
	SecretRules     []YAMLRule `yaml:"secret_rules"`
	TLSConfigRules  []YAMLRule `yaml:"tls_config_rules"`
}

// LoadRulesFromYAML loads and compiles rules from a YAML file
func LoadRulesFromYAML(filepath string) (*RulesConfig, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read rules file %s: %w", filepath, err)
	}

	var config RulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse rules YAML: %w", err)
	}

	// Compile regexes for source code rules
	for i := range config.SourceCodeRules {
		r, err := regexp.Compile(config.SourceCodeRules[i].Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex in rule %s (%s): %w", config.SourceCodeRules[i].ID, config.SourceCodeRules[i].Pattern, err)
		}
		config.SourceCodeRules[i].CompiledRegex = r

		// Compile key size capture regex if specified
		if config.SourceCodeRules[i].KeySizeCapture != "" {
			cr, err := regexp.Compile(config.SourceCodeRules[i].KeySizeCapture)
			if err == nil {
				config.SourceCodeRules[i].CompiledCapture = cr
			}
		}

		// Default confidence to 1.0 if not specified
		if config.SourceCodeRules[i].Confidence == 0 {
			config.SourceCodeRules[i].Confidence = 1.0
		}
	}

	// Compile regexes for secret rules
	for i := range config.SecretRules {
		r, err := regexp.Compile(config.SecretRules[i].Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex in secret rule %s (%s): %w", config.SecretRules[i].ID, config.SecretRules[i].Pattern, err)
		}
		config.SecretRules[i].CompiledRegex = r
	}

	// Compile regexes for TLS config rules
	for i := range config.TLSConfigRules {
		r, err := regexp.Compile(config.TLSConfigRules[i].Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex in TLS rule %s (%s): %w", config.TLSConfigRules[i].ID, config.TLSConfigRules[i].Pattern, err)
		}
		config.TLSConfigRules[i].CompiledRegex = r
	}

	return &config, nil
}
