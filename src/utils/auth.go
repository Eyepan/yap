package utils

import (
	"strings"

	"github.com/Eyepan/yap/src/types"
)

// ExtractAuthToken retrieves the authentication token from the configuration.
func ExtractAuthToken(config *types.Config) string {
	for key, value := range *config {
		if strings.HasSuffix(key, "_authToken") || strings.HasSuffix(key, "_auth") {
			return value
		}
	}
	return ""
}
