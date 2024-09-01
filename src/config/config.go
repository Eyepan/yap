package config

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Eyepan/yap/src/types"
)

// ParsePackageJSON reads and parses package.json.
func ParsePackageJSON() (types.PackageJSON, error) {
	filePath := filepath.Join(".", "package.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return types.PackageJSON{}, err
	}

	var pkgJSON types.PackageJSON
	if err := json.Unmarshal(data, &pkgJSON); err != nil {
		return types.PackageJSON{}, err
	}

	return pkgJSON, nil
}

// readConfigFile reads and parses configuration files.
func readConfigFile(filePath string) (types.Config, error) {
	config := make(types.Config)
	file, err := os.Open(filePath)
	if err != nil {
		return config, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		config[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	return config, nil
}

// LoadConfigurations loads configurations from global and local .npmrc files.
func LoadConfigurations() (types.Config, error) {
	config := types.Config{"registry": "https://registry.npmjs.org"}

	homeDir, _ := os.UserHomeDir()
	globalConfigPath := filepath.Join(homeDir, ".npmrc")
	localConfigPath := filepath.Join(".", ".npmrc")

	for _, path := range []string{globalConfigPath, localConfigPath} {
		if cfg, err := readConfigFile(path); err == nil {
			for k, v := range cfg {
				config[k] = v
			}
		}
	}

	return config, nil
}

// ExtractAuthToken retrieves the authentication token from the configuration.
func ExtractAuthToken(config types.Config) string {
	for key, value := range config {
		if strings.HasSuffix(key, "_authToken") || strings.HasSuffix(key, "_auth") {
			return value
		}
	}
	return ""
}
