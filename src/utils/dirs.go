package utils

import (
	"fmt"
	"io"
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

// hardlinkOrCopyRecursively tries to hardlink directories/files and falls back to copying if hardlinking fails.
// If the error is that the file already exists, it skips the fallback to copying.
func HardLinkOrCopyRecursively(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Construct the target path by replacing the source prefix with target prefix
		relPath, err := filepath.Rel(sourceDir, srcPath)
		if err != nil {
			return err
		}
		destPath := filepath.Join(targetDir, relPath)

		// If it's a directory, create it in the target
		if info.IsDir() {
			if err := os.MkdirAll(destPath, info.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", destPath, err)
			}
			return nil
		}

		// Try to hardlink the file
		if err := os.Link(srcPath, destPath); err != nil {
			// If the error is that the file already exists, log a warning and skip copying
			if os.IsExist(err) {
				slog.Warn(fmt.Sprintf("file already exists, skipping: %s", destPath))
				return nil
			}

			// Log a warning and fallback to copying the file
			slog.Warn(fmt.Sprintf("hardlink failed for %s, falling back to copy: %v", srcPath, err))

			// Fall back to copying the file if hardlink fails for another reason
			if err := CopyFile(srcPath, destPath); err != nil {
				return fmt.Errorf("failed to copy file %s to %s: %v", srcPath, destPath, err)
			}
		}

		return nil
	})
}

func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// Ensure the copied file has the same permissions as the source
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	return os.Chmod(dst, sourceInfo.Mode())
}
