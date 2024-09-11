package metadata

import (
	"bytes"
	"encoding/json"
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

func FetchMetadata(client *http.Client, pkg *types.Package, conf *types.YapConfig, forceFetch bool) (*types.Metadata, error) {
	cacheDir, err := utils.GetCacheDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get cache directory: %w", err)
	}
	cacheFile := filepath.Join(cacheDir, utils.SanitizePackageName(pkg.Name))

	if !forceFetch {
		if data, err := os.ReadFile(cacheFile); err == nil {
			return utils.ReadMetadata(bytes.NewReader(data))
		}
	}

	registryURL := fmt.Sprintf("%s/%s", conf.Registry, pkg.Name)
	req, err := http.NewRequest("GET", registryURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", conf.AuthToken))
	req.Header.Add("Accept", "application/vnd.npm.install-v1+json; q=1.0, application/json; q=0.8, */*")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var metadata types.Metadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, err
	}

	if err := os.MkdirAll(cacheDir, os.ModePerm); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := utils.WriteMetadata(&buf, metadata); err != nil {
		return nil, err
	}

	if err := os.WriteFile(cacheFile, buf.Bytes(), os.ModePerm); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func FetchVersionMetadata(client *http.Client, pkg *types.Package, conf *types.YapConfig, forceFetch bool) (types.VersionMetadata, error) {
	metadata, err := FetchMetadata(client, pkg, conf, forceFetch)
	if err != nil {
		return types.VersionMetadata{}, fmt.Errorf("failed to fetch metadata for package %s@%s: %w", pkg.Name, pkg.Version, err)
	}

	resolvedVersion, err := resolveVersion(pkg, metadata)
	if err != nil {
		slog.Warn(fmt.Sprintf("Failed to resolve version for %s@%s, refetching metadata...", pkg.Name, pkg.Version))
		metadata, err = FetchMetadata(client, pkg, conf, true)
		if err != nil {
			return types.VersionMetadata{}, fmt.Errorf("failed to refetch metadata for package %s@%s: %w", pkg.Name, pkg.Version, err)
		}

		resolvedVersion, err = resolveVersion(pkg, metadata)
		if err != nil {
			return types.VersionMetadata{}, fmt.Errorf("failed to resolve version for package %s after refetch: %w", pkg.Name, err)
		}
	}

	slog.Info(fmt.Sprintf("[METADATA] Successfully resolved package %s@%s with version %s", pkg.Name, pkg.Version, resolvedVersion))

	versionMetadata, ok := metadata.Versions[resolvedVersion]
	if !ok {
		return types.VersionMetadata{}, fmt.Errorf("version %s not found in metadata for package %s", resolvedVersion, pkg.Name)
	}

	return versionMetadata, nil
}

func resolveVersion(pkg *types.Package, metadata *types.Metadata) (string, error) {
	switch pkg.Version {
	case "latest":
		return metadata.DistTags.Latest, nil
	case "next":
		return metadata.DistTags.Next, nil
	case "":
		return metadata.DistTags.Latest, nil
	case "*":
		return metadata.DistTags.Latest, nil
	default:
		// handle edge case scenarios: https://docs.npmjs.com/cli/v10/configuring-npm/package-json#dependencies
		if strings.HasPrefix(pkg.Version, "http://") || strings.HasPrefix(pkg.Version, "https://") || strings.HasPrefix(pkg.Version, "git://") || strings.HasPrefix(pkg.Version, "git+") || strings.HasPrefix(pkg.Version, "npm:") {
			return "", fmt.Errorf("[METADATA] URL Prefixes aren't being handled now. Failed to handle %s", pkg.Version)
		}
		if strings.ContainsRune(pkg.Version, '/') {
			return "", fmt.Errorf("[METADATA] Github URLs aren't being handled now. Failed to handle %s", pkg.Version)
		}
		versions := make([]string, len(metadata.Versions))
		i := 0
		for k := range metadata.Versions {
			versions[i] = k
			i++
		}
		resolved, _ := utils.ResolveVersionForPackage(pkg, versions)
		return resolved, nil
	}
}

func GetListOfDependenciesFromVersionMetadata(md *types.VersionMetadata) []types.Package {
	deps := make([]types.Package, len(md.Dependencies))
	i := 0
	for name, version := range md.Dependencies {
		deps[i] = types.Package{Name: name, Version: version}
		i++
	}
	return deps
}
