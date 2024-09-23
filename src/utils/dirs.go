package utils

import (
	"fmt"
	"log/slog"
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

func GetNodeModulesDir() (string, error) {
	nodeModulesDir := "node_modules" // TODO: should be a better way of doing this,
	return nodeModulesDir, nil
}

func GetDotYapDir() (string, error) {
	nodeModulesDir := filepath.Join("node_modules", ".yap")
	return nodeModulesDir, nil
}

// TODO: local configuration
// func GetLocalConfigDir() (string, error) {

// }

func HardLinkTwoDirectories(sourceDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("[SHIP] error creating destination directory: %w", err)
	}

	// Read files from the source directory
	files, err := os.ReadDir(sourceDir)
	if err != nil {
		return fmt.Errorf("[SHIP] error reading source directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() {
			sourcePath := filepath.Join(sourceDir, file.Name())
			destPath := filepath.Join(destDir, file.Name())

			// Create a hard link
			if err := os.Link(sourcePath, destPath); err != nil {
				return fmt.Errorf("[SHIP] error creating hard link for %s: %w", file.Name(), err)
			}
			slog.Info(fmt.Sprintf("[SHIP] Created hard link: %s -> %s\n", sourcePath, destPath))
		}
	}
	return nil
}
