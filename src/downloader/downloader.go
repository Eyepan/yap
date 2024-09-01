package downloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// DownloadTarballAndExtract downloads a tarball and extracts it to storeLocation
func DownloadTarballAndExtract(tarballURL, packageName, storeLocation string) error {
	// Fetch the tarball
	resp, err := http.Get(tarballURL)
	if err != nil {
		return fmt.Errorf("failed to fetch tarball: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch tarball: status code %d", resp.StatusCode)
	}

	// Create a buffer to read the tarball data into
	var tarballData bytes.Buffer
	if _, err := io.Copy(&tarballData, resp.Body); err != nil {
		return fmt.Errorf("failed to read tarball data: %w", err)
	}

	// Uncompress and extract the tarball
	if err := extractTarball(&tarballData, packageName, storeLocation); err != nil {
		return fmt.Errorf("failed to extract tarball: %w", err)
	}

	return nil
}

// extractTarball extracts tarball data from a buffer
func extractTarball(tarballData *bytes.Buffer, packageName, storeLocation string) error {
	gzipReader, err := gzip.NewReader(tarballData)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	// Create a directory for the package
	// homeDir, err := os.UserHomeDir()
	// packageDir := filepath.Join(homeDir, ".yap_store", packageName)
	packageDir := path.Join(storeLocation, packageName)
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	var topLevelDir string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tarball entry: %w", err)
		}

		// Identify the top-level directory
		if topLevelDir == "" && header.Typeflag == tar.TypeDir && strings.ContainsRune(header.Name, '@') {
			// Set the top-level directory based on the first directory encountered
			topLevelDir = header.Name
		} else {
			topLevelDir = "package"
		}

		// Handle paths based on the top-level directory
		if topLevelDir != "" {
			// Strip the top-level directory from the file path
			relativePath := strings.TrimPrefix(header.Name, topLevelDir)
			if relativePath == header.Name {
				// If there's no top-level prefix, use the original name
				relativePath = header.Name
			}

			// Construct the file path
			filePath := filepath.Join(packageDir, relativePath)

			switch header.Typeflag {
			case tar.TypeDir:
				// Create directories
				if err := os.MkdirAll(filePath, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", filePath, err)
				}
			case tar.TypeReg:
				// Create files
				if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
					return fmt.Errorf("failed to create directory for file %s: %w", filePath, err)
				}
				file, err := os.Create(filePath)
				if err != nil {
					return fmt.Errorf("failed to create file %s: %w", filePath, err)
				}
				if _, err := io.Copy(file, tarReader); err != nil {
					file.Close()
					return fmt.Errorf("failed to write to file %s: %w", filePath, err)
				}
				file.Close()
			default:
				fmt.Printf("Skipping unsupported tarball entry type %c: %s\n", header.Typeflag, header.Name)
			}
		}
	}

	return nil
}

func CheckIfPackageIsAlreadyDownloaded(packageName string) (bool, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, fmt.Errorf("could not get home directory: %v", err)
	}

	packagePath := filepath.Join(homeDir, ".yap_store", packageName)
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("could not check package path: %v", err)
	}

	return true, nil
}
