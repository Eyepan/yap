package downloader

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

// ProgressReader wraps an io.Reader and tracks the progress of the read operation.
type ProgressReader struct {
	io.Reader
	total    int64
	progress int64
	name     string
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.Reader.Read(p)
	pr.progress += int64(n)

	// Report progress
	// fmt.Sprintf("\rDownloading from %s %.2f%% complete", pr.name, float64(pr.progress)/float64(pr.total)*100)

	return n, err
}

func DownloadPackage(pkg types.Package, tarballURL string, npmrc types.Config, force bool) (bool, error) {
	if check, _ := CheckIfPackageIsAlreadyDownloaded(pkg); !force && check {
		slog.Info(fmt.Sprintf("%s@%s has already been downloaded. Reusing this from the store", pkg.Name, pkg.Version))
		return true, nil
	}
	tarballData, err := DownloadTarball(tarballURL, npmrc)
	if err != nil {
		return false, fmt.Errorf("failed while downloading tarball: %w", err)
	}
	err = ExtractTarball(tarballData, fmt.Sprintf("%s@%s", pkg.Name, pkg.Version))
	if err != nil {
		return false, fmt.Errorf("failed while extracting tarball: %w", err)
	}
	return true, nil
}

func DownloadTarball(tarballURL string, npmrc types.Config) (*bytes.Buffer, error) {
	authToken := utils.ExtractAuthToken(npmrc)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", tarballURL, nil)
	if err != nil {
		return nil, err
	}

	// Add the auth token to the request headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch tarball: status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	// Get the content length (total size of the tarball)
	totalSize := resp.ContentLength

	// Wrap the response body in the ProgressReader
	progressReader := &ProgressReader{
		Reader: resp.Body,
		total:  totalSize,
		name:   tarballURL,
	}

	// Create a buffer to store the tarball data
	var tarballData bytes.Buffer
	if _, err := io.Copy(&tarballData, progressReader); err != nil {
		return nil, fmt.Errorf("failed to read tarball data: %w", err)
	}

	// Print a final newline after the progress is complete
	// fmt.Println("\nDownload complete")

	return &tarballData, nil
}

func ExtractTarball(tarballData *bytes.Buffer, packageID string) error {
	gzipReader, err := gzip.NewReader(tarballData)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	storeDir, err := utils.GetStoreDir()
	if err != nil {
		return fmt.Errorf("failed to get store directory: %w", err)
	}
	// create dir for package
	packageDir := filepath.Join(storeDir, utils.SanitizePackageName(packageID))

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

		if topLevelDir == "" && header.Typeflag == tar.TypeDir && strings.ContainsRune(header.Name, '@') {
			topLevelDir = header.Name
		} else {
			topLevelDir = "package"
		}

		if topLevelDir != "" {
			relativePath := strings.TrimPrefix(header.Name, topLevelDir)
			filePath := filepath.Join(packageDir, relativePath)

			switch header.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(filePath, 0755); err != nil {
					return fmt.Errorf("failed to create directory %s: %w", filePath, err)
				}
			case tar.TypeReg:
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
				slog.Warn(fmt.Sprintf("skipping unsupported tarball entry type %c: %s", header.Typeflag, header.Name))
			}
		}
	}
	return nil
}

func CheckIfPackageIsAlreadyDownloaded(pkg types.Package) (bool, error) {
	storeDir, err := utils.GetStoreDir()
	if err != nil {
		return false, fmt.Errorf("failed to get store directory: %w", err)
	}

	packagePath := filepath.Join(storeDir, utils.SanitizePackageName(fmt.Sprintf("%s@%s", pkg.Name, pkg.Version)))

	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check package path: %w", err)
	}

	return true, nil
}
