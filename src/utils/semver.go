package utils

import (
	"fmt"
	"sort"

	"github.com/Eyepan/yap/src/types"
	"github.com/Masterminds/semver/v3"
)

func ResolveVersionForPackage(pkg *types.Package, availableVersions []string) (string, error) {
	versionList := make([]*semver.Version, len(availableVersions))
	for i, v := range availableVersions {
		version, err := semver.NewVersion(v)
		if err != nil {
			continue
		}
		versionList[i] = version
	}
	sort.Sort(semver.Collection(versionList))

	constraint, err := semver.NewConstraint((*pkg).Version)
	if err != nil {
		return "", err
	}

	for _, ver := range versionList {
		if constraint.Check(ver) {
			return ver.String(), nil
		}
	}

	return "", fmt.Errorf("no matching version found for package %s@%s: found versions %v", pkg.Name, pkg.Version, availableVersions)
}

func DetermineIfPackageVersionIsResolvableDirectly(pkg types.Package) (string, error) {
	// TODO:
	return "", fmt.Errorf("hasn't been implemented yet")
}
