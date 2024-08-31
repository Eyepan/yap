package resolver

import (
	"fmt"

	"github.com/Eyepan/yap/config"
	"github.com/Eyepan/yap/dependencies"
	"github.com/Eyepan/yap/fetcher"
)

type MetadataDist struct {
	Shasum    string `json:"shasum"`
	Tarball   string `json:"tarball"`
	FileCount int    `json:"fileCount"`
}

type VersionMetadata struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Id           string            `json:"_id"`
	Dist         MetadataDist      `json:"dist"`
	Dependencies map[string]string `json:"dependencies"`
}

type Tags struct {
	Latest string `json:"latest"`
	Next   string `json:"next"`
}

type Metadata struct {
	Name     string                     `json:"name"`
	Id       string                     `json:"_id"`
	DistTags Tags                       `json:"dist-tags"`
	Versions map[string]VersionMetadata `json:"versions"`
}

func FetchPackageMetadata(name, version string, npmrc config.Config, fetchCache *fetcher.Cache) (*VersionMetadata, error) {
	registryURL := npmrc["registry"]
	authToken := config.ExtractAuthToken(npmrc)

	// Build the package URL
	packageURL := fmt.Sprintf("%s/%s", registryURL, name)

	var metadata Metadata
	fetchCache.Fetch(packageURL, authToken, &metadata)
	// Resolve the semver to the exact version
	var exactVersion string
	if version == "latest" {
		exactVersion = metadata.DistTags.Latest
	} else if version == "next" {
		exactVersion = metadata.DistTags.Next
	} else {
		// Manually resolve the version
		exactVersion = ResolveVersion(version, metadata.Versions)
		if exactVersion == "" {
			return nil, fmt.Errorf("could not resolve version for package %s with version %s", name, version)
		}
	}

	// Get the VersionMetadata from the resolved version
	versionMetadata, exists := metadata.Versions[exactVersion]
	if !exists {
		return nil, fmt.Errorf("version not found for %s at %s", name, version)
	}
	return &versionMetadata, nil
}

func GetSubdependencies(metadata *VersionMetadata) ([]dependencies.Package, error) {
	var packages []dependencies.Package
	for k, v := range metadata.Dependencies {
		packages = append(packages, dependencies.Package{Name: k, Version: v})
	}
	return packages, nil
}
