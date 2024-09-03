package config

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Eyepan/yap/src/types"
)

// function that reads npmrc config
func parseNpmrc(filePath string) (types.Config, error) {
	config := make(types.Config)
	content, err := os.ReadFile(filePath)
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

func LoadConfigurations() (types.Config, error) {
	config := types.Config{"registry": "https://registry.npmjs.org"}

	homeDir, _ := os.UserHomeDir()
	globalConfigPath := filepath.Join(homeDir, ".npmrc")
	localConfigPath := filepath.Join(".", ".npmrc")

	for _, path := range []string{globalConfigPath, localConfigPath} {
		if cfg, err := parseNpmrc(path); err == nil {
			for k, v := range cfg {
				config[k] = v
			}
		}
	}

	return config, nil
}
