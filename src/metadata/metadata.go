package metadata

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/fetcher"
	"github.com/Eyepan/yap/src/types"
	"github.com/Masterminds/semver/v3"
)

// FetchPackageMetadata retrieves metadata for a given package.
func FetchPackageMetadata(pkg types.Package, cache *fetcher.FSCache, npmrc types.Config) (types.VersionMetadata, error) {
	registryURL := npmrc["registry"]
	packageURL := fmt.Sprintf("%s/%s", registryURL, pkg.Name)
	data, err := cache.Fetch(packageURL, pkg.Name, config.ExtractAuthToken(npmrc))
	if err != nil {
		return types.VersionMetadata{}, err
	}

	var metadata types.Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return types.VersionMetadata{}, err
	}

	var exactVersion string
	if pkg.Version == "latest" {
		exactVersion = metadata.DistTags.Latest
	} else if pkg.Version == "next" {
		exactVersion = metadata.DistTags.Next
	} else {
		resolvedVersion, err := resolveVersion(pkg.Name, pkg.Version, getKeys(metadata.Versions))
		if err != nil {
			return types.VersionMetadata{}, err
		}
		exactVersion = resolvedVersion
	}

	return metadata.Versions[exactVersion], nil
}

// GetSubdependenciesFromMetadata extracts subdependencies from metadata.
func GetSubdependenciesFromMetadata(metadata types.VersionMetadata) []types.Package {
	var dependencies []types.Package
	for name, version := range metadata.Dependencies {
		dependencies = append(dependencies, types.Package{Name: name, Version: version})
	}
	return dependencies
}

// resolveVersion determines the appropriate version based on the constraint.
func resolveVersion(pkgName, version string, versions []string) (string, error) {
	var versionList []*semver.Version
	for _, v := range versions {
		ver, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		versionList = append(versionList, ver)
	}

	sort.Sort(semver.Collection(versionList))

	constraint, err := semver.NewConstraint(version)
	if err != nil {
		return "", err
	}

	for _, ver := range versionList {
		if constraint.Check(ver) {
			return ver.String(), nil
		}
	}

	return "", fmt.Errorf("no matching version found for %s %s", pkgName, version)
}

// getKeys returns the keys from a map.
func getKeys(m map[string]types.VersionMetadata) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
