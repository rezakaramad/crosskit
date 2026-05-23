package nextinsight

import (
	"regexp"
	"strings"
)

const labelPrefix = "next-insight.io/"

// AppMetadata holds the resolved metadata for a Next-Insight application.
type AppMetadata struct {
	ApplicationID     string
	ApplicationName   string
	Lifecycle         string
	LifecycleDecision string
	Criticality       string
	Complexity        string
	Category          string
	DevelopmentType   string
	SourcingType      string
	FacingInternet    string // "true" | "false" | ""

	// Holds the single Agile Release Train associated with this application.
	AgileReleaseTrain string

	// Holds the single Agile Team associated with this application.
	AgileTeam string
}

// Labels returns a map of Kubernetes-safe labels derived from the metadata.
// Kubernetes label values must satisfy the following constraints:
//   - Must be 63 characters or less (can be empty)
//   - Must consist of lowercase alphanumeric characters, '-' or '.'
//   - Must start and end with an alphanumeric character
//
// normalize() is applied to all label values to ensure these are met.
func (m *AppMetadata) Labels() map[string]string {
	// We prefix all labels with "next-insight.io/" to avoid collisions with other labels
	labels := make(map[string]string)

	// If a metadata field is empty, we won't set a label for it
	set := func(key, value string) {
		if value != "" {
			labels[key] = value
		}
	}

	// Set labels for all metadata fields after normalization
	set(labelPrefix+"app-id", m.ApplicationID)
	set(labelPrefix+"app-name", normalize(m.ApplicationName))
	set(labelPrefix+"lifecycle", normalize(m.Lifecycle))
	set(labelPrefix+"lifecycle-decision", normalize(m.LifecycleDecision))
	set(labelPrefix+"criticality", normalize(m.Criticality))
	set(labelPrefix+"complexity", normalize(m.Complexity))
	set(labelPrefix+"category", normalize(m.Category))
	set(labelPrefix+"development-type", normalize(m.DevelopmentType))
	set(labelPrefix+"sourcing-type", normalize(m.SourcingType))
	set(labelPrefix+"facing-internet", m.FacingInternet)
	set(labelPrefix+"agile-release-train", normalize(m.AgileReleaseTrain))
	set(labelPrefix+"agile-team", normalize(m.AgileTeam))

	return labels
}

// normalize converts a string to a Kubernetes-safe label value.
//
// Rules applied:
//  1. Lowercase
//  2. Spaces, underscores, and forward-slashes become hyphens
//  3. Characters outside [a-z0-9-.] are dropped
//  4. Leading/trailing hyphens and dots are stripped
//  5. Consecutive hyphens are collapsed to one
//  6. Result is truncated to 63 characters (K8s label value max)

// Returns a regex object which matches two or more consecutive hyphens, used to collapse them into one
var multiHyphen = regexp.MustCompile(`-{2,}`)

func normalize(str string) string {
	// Step 1: Lowercase and trim whitespace
	// E.g., "  My App/Name_v2.0  " -> "my app/name_v2.0"
	str = strings.ToLower(strings.TrimSpace(str))

	var buffer strings.Builder
	// Allocate enough memory for the string
	buffer.Grow(len(str))

	// For each character in the string, apply rules 2 and 3:
	// - If it's a space, underscore, or forward-slash, write a hyphen
	// - If it's a lowercase letter, digit, hyphen, or dot, write it as is
	// - Otherwise, skip it
	//
	// E.g., "my app/name_v2.0" -> "my-app-name-v2.0"
	for _, char := range str {
		switch {
		case char >= 'a' && char <= 'z', char >= '0' && char <= '9':
			buffer.WriteRune(char)
		case char == ' ', char == '_', char == '/':
			buffer.WriteRune('-')
		case char == '-', char == '.':
			buffer.WriteRune(char)
		}
	}

	// Step 4: Trim leading/trailing hyphens and dots
	// E.g., "-my-app-name-v2.0-" -> "my-app-name-v2.0"
	result := strings.Trim(buffer.String(), "-.")
	// Step 5: Collapse multiple consecutive hyphens into one
	// E.g., "my--app---name-v2.0" -> "my-app-name-v2.0"
	result = multiHyphen.ReplaceAllString(result, "-")

	// Step 6: Truncate to 63 characters
	// E.g., "a-very-long-app-name-that-exceeds-kubernetes-label-length-limit" -> "a-very-long-app-name-that-exceeds-kubernetes-label-length-l"
	if len(result) > 63 {
		result = strings.TrimRight(result[:63], "-.")
	}

	return result
}
