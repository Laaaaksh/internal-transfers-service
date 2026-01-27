package config

import (
	"os"
	"regexp"
	"strings"
)

// envVarPattern matches ${VAR:-default} or ${VAR} syntax in TOML values
var envVarPattern = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)(:-([^}]*))?\}`)

// expandEnvVars expands environment variable templates in TOML content.
// Supports syntax: ${VAR} or ${VAR:-default}
func expandEnvVars(content []byte) []byte {
	expanded := envVarPattern.ReplaceAllFunc(content, func(match []byte) []byte {
		matchStr := string(match)
		submatches := envVarPattern.FindStringSubmatch(matchStr)

		if len(submatches) < 2 {
			return match
		}

		envVarName := submatches[1]
		defaultValue := ""
		if len(submatches) >= 4 {
			defaultValue = submatches[3]
		}

		if envValue := os.Getenv(envVarName); envValue != "" {
			return []byte(envValue)
		}

		return []byte(defaultValue)
	})

	return expanded
}

// expandEnvVarsInString is a convenience wrapper for string inputs.
func expandEnvVarsInString(content string) string {
	return string(expandEnvVars([]byte(content)))
}

// isTemplateValue checks if a string contains template syntax.
func isTemplateValue(value string) bool {
	return strings.Contains(value, "${")
}
