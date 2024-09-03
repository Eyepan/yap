package utils

import "strings"

// SanitizePackageName replaces slashes in the package name with a hyphen or another character
func SanitizePackageName(pkgName string) string {
	return strings.ReplaceAll(pkgName, "/", "-")
}
