package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/Eyepan/yap/src/cacher"
	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/dependencies"
	"github.com/Eyepan/yap/src/types"
)

// Main execution
func main() {
	cache := cacher.FSCache{CacheDir: "cache"}

	npmrc, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalln("Error parsing config", err)
	}

	packageJSON, err := config.ParsePackageJSON()
	if err != nil {
		fmt.Println("Error parsing package.json:", err)
		return
	}

	allDependencies := []types.Dependency{}
	for key, value := range packageJSON.Dependencies {
		allDependencies = append(allDependencies, types.Dependency{Name: key, Version: value})
	}
	for key, value := range packageJSON.DevDependencies {
		allDependencies = append(allDependencies, types.Dependency{Name: key, Version: value})
	}
	for key, value := range packageJSON.PeerDependencies {
		allDependencies = append(allDependencies, types.Dependency{Name: key, Version: value})
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var totalSubdependencies []types.Dependency

	for _, dep := range allDependencies {
		wg.Add(1)
		go func(d types.Dependency) {
			defer wg.Done()
			subdeps, err := dependencies.GetAllSubdependencies(d, &cache, npmrc)
			if err != nil {
				fmt.Println("Error fetching subdependencies:", err)
				return
			}
			mu.Lock()
			totalSubdependencies = append(totalSubdependencies, subdeps...)
			mu.Unlock()
		}(dep)
	}

	wg.Wait()

	// Remove duplicates
	uniqueSubdeps := map[string]types.Dependency{}
	for _, dep := range totalSubdependencies {
		uniqueSubdeps[dep.Name] = dep
	}

	var finalSubdependencies []types.Dependency
	for _, dep := range uniqueSubdeps {
		finalSubdependencies = append(finalSubdependencies, dep)
	}

	fmt.Println("Total subdependencies:", len(finalSubdependencies))
}
