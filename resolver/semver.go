package resolver

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func ResolveVersion(version string, versions map[string]VersionMetadata) string {
	// Direct match
	if _, exists := versions[version]; exists {
		return version
	}

	// Handle tilde (~) and caret (^) ranges
	if strings.HasPrefix(version, "~") {
		return resolveTildeVersion(version[1:], versions)
	} else if strings.HasPrefix(version, "^") {
		return resolveCaretVersion(version[1:], versions)
	}

	// Handle wildcards (e.g., "1.2.x", "1.x", "x")
	if strings.Contains(version, "x") {
		return resolveWildcardVersion(version, versions)
	}

	// Handle range versions (e.g., ">= 2.1.2 < 3.0.0")
	if strings.Contains(version, ">") || strings.Contains(version, "<") {
		return resolveRangeVersion(version, versions)
	}

	// Fallback to empty string if no version matches
	return ""
}

func resolveRangeVersion(version string, versions map[string]VersionMetadata) string {
	// Parse the range string
	rangeRegex := regexp.MustCompile(`(>=|<=|>|<)\s*([\d\.]+)`)
	matches := rangeRegex.FindAllStringSubmatch(version, -1)
	if len(matches) == 0 {
		return ""
	}

	// Extract bounds
	var lowerBound, upperBound string
	for _, match := range matches {
		if match[1] == ">=" || match[1] == ">" {
			lowerBound = match[2]
		} else if match[1] == "<=" || match[1] == "<" {
			upperBound = match[2]
		}
	}

	// Collect versions within the range
	var validVersions []string
	for v := range versions {
		if (lowerBound == "" || compareVersions(v, lowerBound) >= 0) &&
			(upperBound == "" || compareVersions(v, upperBound) < 0) {
			validVersions = append(validVersions, v)
		}
	}

	// Sort and return the highest version
	sort.Strings(validVersions)
	if len(validVersions) > 0 {
		return validVersions[len(validVersions)-1]
	}

	return ""
}

func compareVersions(v1, v2 string) int {
	segments1 := strings.Split(v1, ".")
	segments2 := strings.Split(v2, ".")

	maxLen := len(segments1)
	if len(segments2) > maxLen {
		maxLen = len(segments2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		var err error

		if i < len(segments1) {
			num1, err = strconv.Atoi(segments1[i])
			if err != nil {
				return 0
			}
		}

		if i < len(segments2) {
			num2, err = strconv.Atoi(segments2[i])
			if err != nil {
				return 0
			}
		}

		if num1 < num2 {
			return -1
		} else if num1 > num2 {
			return 1
		}
	}

	return 0
}

func resolveTildeVersion(version string, versions map[string]VersionMetadata) string {
	for v := range versions {
		if matchTilde(version, v) {
			return v
		}
	}
	return ""
}

func resolveCaretVersion(version string, versions map[string]VersionMetadata) string {
	for v := range versions {
		if matchCaret(version, v) {
			return v
		}
	}
	return ""
}

func resolveWildcardVersion(version string, versions map[string]VersionMetadata) string {
	for v := range versions {
		if matchWildcard(version, v) {
			return v
		}
	}
	return ""
}

func matchTilde(baseVersion, candidateVersion string) bool {
	// Example: baseVersion = "1.2.3", candidateVersion = "1.2.4"
	baseParts := strings.Split(baseVersion, ".")
	candidateParts := strings.Split(candidateVersion, ".")

	// Major and minor versions must match
	if baseParts[0] != candidateParts[0] || baseParts[1] != candidateParts[1] {
		return false
	}

	// Candidate patch version must be >= base patch version
	basePatch, _ := strconv.Atoi(baseParts[2])
	candidatePatch, _ := strconv.Atoi(candidateParts[2])
	return candidatePatch >= basePatch
}

func matchCaret(baseVersion, candidateVersion string) bool {
	// Example: baseVersion = "1.2.3", candidateVersion = "1.3.0"
	baseParts := strings.Split(baseVersion, ".")
	candidateParts := strings.Split(candidateVersion, ".")

	// Major version must match, minor and patch can be greater or equal
	if baseParts[0] != candidateParts[0] {
		return false
	}

	// Check if minor or patch version is greater or equal
	baseMinor, _ := strconv.Atoi(baseParts[1])
	candidateMinor, _ := strconv.Atoi(candidateParts[1])
	return candidateMinor >= baseMinor
}

func matchWildcard(baseVersion, candidateVersion string) bool {
	// Example: baseVersion = "1.2.x", candidateVersion = "1.2.3"
	baseParts := strings.Split(baseVersion, ".")
	candidateParts := strings.Split(candidateVersion, ".")

	for i := range baseParts {
		if baseParts[i] == "x" {
			continue
		}
		if baseParts[i] != candidateParts[i] {
			return false
		}
	}

	return true
}
