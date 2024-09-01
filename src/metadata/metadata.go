package metadata

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/Eyepan/yap/src/cacher"
	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/types"
	"github.com/Masterminds/semver/v3"
)

// Fetch metadata of a package
func FetchPackageMetadata(pkg types.Dependency, cache *cacher.FSCache) (types.VersionMetadata, error) {
	npmrc, err := config.LoadConfigurations()
	if err != nil {
		return types.VersionMetadata{}, fmt.Errorf("error loading configurations: \n%w", err)
	}

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

func GetSubdependenciesFromMetadata(metadata types.VersionMetadata) []types.Dependency {
	var dependencies []types.Dependency
	for name, version := range metadata.Dependencies {
		dependencies = append(dependencies, types.Dependency{Name: name, Version: version})
	}
	return dependencies
}

// Resolve version based on semver and metadata
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

// Fetch keys from a map
func getKeys(m map[string]types.VersionMetadata) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
