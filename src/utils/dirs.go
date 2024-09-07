package utils

import (
	"fmt"
	"os"
	"path/filepath"
)

func GetCacheDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory for user: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".yap_store", ".yap_cache")
	return cacheDir, nil
}

func GetStoreDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory for user: %w", err)
	}
	storeDir := filepath.Join(homeDir, ".yap_store")
	return storeDir, nil
}

func GetGlobalConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory for user: %w", err)
	}
	globalConfigDir := filepath.Join(homeDir, ".yap_config")
	return globalConfigDir, nil
}

// TODO: local configuration
// func GetLocalConfigDir() (string, error) {

// }
