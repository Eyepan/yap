package config

import (
	"encoding/json"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Eyepan/yap/src/types"
)

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

// Utility function to read configuration files
func readConfigFile(path string) (types.Config, error) {
	config := make(types.Config)
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return config, err
	}

	lines := string(content)
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || line[0] == ';' || line[0] == '#' {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		config[key] = value
	}

	return config, nil
}

// Load configurations from global and local .npmrc files
func LoadConfigurations() (types.Config, error) {
	config := types.Config{"registry": "https://registry.npmjs.org"}
	homeDir, _ := os.UserHomeDir()
	globalConfigPath := path.Join(homeDir, ".npmrc")
	localConfigPath := path.Join(".", ".npmrc")

	if cfg, err := readConfigFile(globalConfigPath); err == nil {
		for k, v := range cfg {
			config[k] = v
		}
	}

	if cfg, err := readConfigFile(localConfigPath); err == nil {
		for k, v := range cfg {
			config[k] = v
		}
	}

	return config, nil
}

// Extract authentication token from configuration
func ExtractAuthToken(config types.Config) string {
	for key, value := range config {
		if strings.HasSuffix(key, "_authToken") || strings.HasSuffix(key, "_auth") {
			return value
		}
	}
	return ""
}
