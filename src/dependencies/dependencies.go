package dependencies

import (
	"sync"

	"github.com/Eyepan/yap/src/fetcher"
	"github.com/Eyepan/yap/src/metadata"
	"github.com/Eyepan/yap/src/types"
)

var metadataCache sync.Map

// GetAllSubdependencies retrieves all subdependencies for a given package.
func GetAllSubdependencies(pkg types.Dependency, cache *fetcher.FSCache, npmrc types.Config) ([]types.Dependency, error) {
	stack := []types.Dependency{pkg}
	uniqueDeps := make(map[string]types.Dependency)

	for len(stack) > 0 {
		currentPkg := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		if _, exists := uniqueDeps[currentPkg.Name]; !exists {
			uniqueDeps[currentPkg.Name] = currentPkg

			var deps []types.Dependency
			if val, ok := metadataCache.Load(currentPkg.Name); ok {
				deps = val.([]types.Dependency)
			} else {
				md, err := metadata.FetchPackageMetadata(currentPkg, cache, npmrc)
				if err != nil {
					return nil, err
				}
				deps = metadata.GetSubdependenciesFromMetadata(md)
				metadataCache.Store(currentPkg.Name, deps)
			}

			for _, dep := range deps {
				if _, exists := uniqueDeps[dep.Name]; !exists {
					stack = append(stack, dep)
				}
			}
		}
	}

	var result []types.Dependency
	for _, dep := range uniqueDeps {
		result = append(result, dep)
	}
	return result, nil
}
