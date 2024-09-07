package metadata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

func FetchMetadata(pkg *types.Package, conf *types.YapConfig, forceFetchAndRefresh bool) (*types.Metadata, error) {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}
	cacheFile := filepath.Join(cacheDir, utils.SanitizePackageName(pkg.Name))

	// this if should only happen if force is false

	// Check if the cache file exists
	if _, err := os.Stat(cacheFile); !forceFetchAndRefresh && err == nil {
		// Cache file exists, read its contents
		data, err := os.ReadFile(cacheFile)
		if err != nil {
			return nil, err
		}

		buf := bytes.NewReader(data)
		return utils.ReadMetadata(buf)
	}

	// Cache file does not exist, fetch metadata from the server
	registryURL := (*conf).Registry
	authToken := (*conf).AuthToken
	packageURL := fmt.Sprintf("%s/%s", registryURL, pkg.Name)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", packageURL, nil)
	if err != nil {
		return nil, err
	}

	// Add the auth token to the request headers
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	req.Header.Add("Accept", "application/vnd.npm.install-v1+json")

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Unmarshal the response body into a types.Metadata variable
	var metadata types.Metadata
	err = json.Unmarshal(body, &metadata)
	if err != nil {
		return nil, err
	}

	// Ensure the cache directory exists
	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, err
	}

	// Write the metadata to the cache file in binary format
	var buf bytes.Buffer
	if err := utils.WriteMetadata(&buf, metadata); err != nil {
		return nil, err
	}

	file, err := os.Create(cacheFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	if _, err := file.Write(buf.Bytes()); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func FetchVersionMetadata(pkg *types.Package, npmrc *types.YapConfig, forceFetchAndRefresh bool) (types.VersionMetadata, error) {
	md, err := FetchMetadata(pkg, npmrc, forceFetchAndRefresh)
	if err != nil {
		return types.VersionMetadata{}, fmt.Errorf("failed to fetch metadata for package %s@%s: %w", pkg.Name, pkg.Version, err)
	}
	versionsList := make([]string, len(md.Versions))
	i := 0
	for k := range md.Versions {
		versionsList[i] = k
		i++
	}
	resolvedVersion, err := utils.ResolveVersionForPackage(pkg, versionsList)
	if err != nil {
		return types.VersionMetadata{}, fmt.Errorf("failed to resolve version for package %s@%s: %w", pkg.Name, pkg.Version, err)
	}

	return md.Versions[resolvedVersion], nil
}

func GetListOfDependenciesFromVersionMetadata(md *types.VersionMetadata) []types.Package {
	deps := make([]types.Package, len(md.Dependencies))
	i := 0
	for key, value := range (*md).Dependencies {
		deps[i] = types.Package{Name: key, Version: value}
		i++
	}

	return deps
}
