package resolver

import (
	"fmt"

	"github.com/Eyepan/yap/config"
	"github.com/Eyepan/yap/dependencies"
	"github.com/Eyepan/yap/fetcher"
)

// GetAllSubdependencies recursively resolves all packages by fetching metadata for each and resolving their subdependencies.
func GetAllSubdependencies(packages []dependencies.Package, npmrc config.Config, fetchCache *fetcher.Cache) ([]dependencies.Package, error) {
	var resolvedPackages []dependencies.Package
	seen := make(map[string]bool)

	for _, pkg := range packages {
		if err := resolvePackage(pkg, npmrc, seen, &resolvedPackages, fetchCache); err != nil {
			return nil, err
		}
	}

	return resolvedPackages, nil
}

// resolvePackage resolves a single package and its subdependencies.
func resolvePackage(pkg dependencies.Package, npmrc config.Config, seen map[string]bool, resolvedPackages *[]dependencies.Package, fetchCache *fetcher.Cache) error {
	if seen[pkg.Name+"@"+pkg.Version] {
		return nil
	}
	seen[pkg.Name+"@"+pkg.Version] = true

	metadata, err := FetchPackageMetadata(pkg.Name, pkg.Version, npmrc, fetchCache)
	if err != nil {
		return fmt.Errorf("fetching metadata failed for package %s: \n%w", pkg.Name, err)
	}

	subdeps, err := GetSubdependencies(metadata)
	if err != nil {
		return fmt.Errorf("getting subdependencies failed: \n%w", err)
	}

	*resolvedPackages = append(*resolvedPackages, subdeps...)

	for _, subdep := range subdeps {
		if err := resolvePackage(subdep, npmrc, seen, resolvedPackages, fetchCache); err != nil {
			return fmt.Errorf("resolving package failed for package %s: \n%w", subdep.Name, err)
		}
	}

	return nil
}
