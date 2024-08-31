package config

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type Config map[string]string

func readConfigFile(path string) (Config, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		log.Printf("configuration file not found at %s, skipping", path)
		return Config{}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	config := make(Config)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			config[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return config, nil
}

// loads $HOME/.npmrc, .npmrc in that order. closest to the project takes precedence
func LoadConfigurations() (Config, error) {
	config := Config{
		"registry": "https://registry.npmjs.org/",
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to read user's home directory: \n%e", err)
	}

	globalConfig, err := readConfigFile(filepath.Join(homeDir, ".npmrc"))
	if err != nil {
		return nil, fmt.Errorf("failed to read $HOME/.npmrc: \n%e", err)
	}
	for k, v := range globalConfig {
		config[k] = v
	}

	localConfig, err := readConfigFile(filepath.Join(".", ".npmrc"))
	if err != nil {
		return nil, fmt.Errorf("failed to read .npmrc: \n%e", err)
	}
	for k, v := range localConfig {
		config[k] = v
	}
	return config, nil
}

// reads _authToken, _auth from .npmrc
func ExtractAuthToken(config Config) string {
	for key, value := range config {
		if strings.HasSuffix(key, "_authToken") || strings.HasSuffix(key, "_auth") {
			return value
		}
	}
	return ""
}
