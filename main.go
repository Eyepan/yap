package main

import (
	"fmt"
	"log"
	"log/slog"
	"runtime"
	"sync"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/downloader"
	"github.com/Eyepan/yap/src/logger"
	"github.com/Eyepan/yap/src/metadata"
	"github.com/Eyepan/yap/src/packagejson"
	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

// Init application
func Init() {
}

func main() {
	slog.SetLogLoggerLevel(slog.LevelWarn)
	// Load configurations
	config, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	var resolveCount = 0
	var totalResolveCount = 0
	var downloadCount = 0
	var totalDownloadCount = 0

	// Parse the command input (for now, assume 'install')
	command := "install" // this would be dynamically set by the CLI parser
	if command == "install" {
		logger.PrettyPrintStats(resolveCount, totalResolveCount, downloadCount, totalDownloadCount)

		// Parse package.json to get core dependencies
		pkgJSON, err := packagejson.ParsePackageJSON()
		if err != nil {
			log.Fatalf("Failed to parse package.json: %v", err)
		}

		// Resolve dependencies concurrently
		var lockBin types.Lockfile
		lockBin.CoreDependencies = []types.Package{}

		results := make(chan *types.MPackage)
		downloads := make(chan types.MPackage)

		var wg sync.WaitGroup

		// Start download workers
		numWorkers := runtime.NumCPU() // Number of concurrent download workers
		go startDownloadWorkers(numWorkers, downloads, config, &resolveCount, &totalResolveCount, &downloadCount, &totalDownloadCount)

		// Start resolving core dependencies
		go func() {
			for name, version := range packagejson.GetAllDependencies(&pkgJSON) {
				pkg := types.Package{Name: name, Version: version}
				lockBin.CoreDependencies = append(lockBin.CoreDependencies, pkg)
				totalResolveCount += 1
				logger.PrettyPrintStats(resolveCount, totalResolveCount, downloadCount, totalDownloadCount)
				wg.Add(1)
				go func(pkg types.Package) {
					defer wg.Done()
					resolved, err := resolvePackage(pkg, config, &resolveCount, &totalResolveCount, &downloadCount, &totalDownloadCount, downloads)
					if err != nil {
						log.Printf("Failed to resolve package %s: %v", pkg.Name, err)
						return
					}
					results <- resolved
				}(pkg)
			}
			wg.Wait()
			close(results)
			close(downloads)
		}()

		// Build the lockfile
		lockBin.Resolutions = []*types.MPackage{}
		for res := range results {
			lockBin.Resolutions = append(lockBin.Resolutions, res)
		}

		utils.WriteLock(lockBin)

		fmt.Println("\n💫Done!")
	}
}

// Resolving process
func resolvePackage(pkg types.Package, config types.Config, resolveCount *int, totalResolveCount *int, downloadCount *int, totalDownloadCount *int, downloads chan<- types.MPackage) (*types.MPackage, error) {
	slog.Info(fmt.Sprintf("Fetching metadata for %s@%s", pkg.Name, pkg.Version))
	vmd, err := metadata.FetchVersionMetadata(pkg, config, false)
	if err != nil {
		log.Fatalln("failed while fetching the metadata", err)
	}
	slog.Info(fmt.Sprintf("Done fetching metadata for %s@%s", pkg.Name, pkg.Version))
	*resolveCount += 1
	logger.PrettyPrintStats(*resolveCount, *totalResolveCount, *downloadCount, *totalDownloadCount)

	// Send package for downloading
	downloads <- types.MPackage{Name: vmd.Name, Version: vmd.Version, Dist: vmd.Dist}

	resolvedPkg := &types.MPackage{
		Name:         vmd.Name,
		Version:      vmd.Version,
		Id:           vmd.ID,
		Dist:         vmd.Dist,
		Dependencies: []*types.MPackage{},
	}

	for depName, depVersion := range vmd.Dependencies {
		depPkg := types.Package{Name: depName, Version: depVersion}
		*totalResolveCount += 1
		logger.PrettyPrintStats(*resolveCount, *totalResolveCount, *downloadCount, *totalDownloadCount)
		subResolved, err := resolvePackage(depPkg, config, resolveCount, totalResolveCount, downloadCount, totalDownloadCount, downloads)
		if err != nil {
			return nil, err
		}
		resolvedPkg.Dependencies = append(resolvedPkg.Dependencies, subResolved)
	}

	return resolvedPkg, nil
}

// Downloading process with worker pool
func startDownloadWorkers(numWorkers int, downloads <-chan types.MPackage, config types.Config, resolveCount *int, totalResolveCount *int, downloadCount *int, totalDownloadCount *int) {
	for i := 0; i < numWorkers; i++ {
		go func() {
			for pkg := range downloads {
				*totalDownloadCount += 1
				logger.PrettyPrintStats(*resolveCount, *totalResolveCount, *downloadCount, *totalDownloadCount)
				slog.Info(fmt.Sprintf("Downloading tarball for %s@%s", pkg.Name, pkg.Version))
				err := downloader.DownloadPackage(types.Package{Name: pkg.Name, Version: pkg.Version}, pkg.Dist.Tarball, config, false)
				if err != nil {
					log.Printf("Failed to download package %s: %v", pkg.Name, err)
					continue
				}
				*downloadCount += 1
				logger.PrettyPrintStats(*resolveCount, *totalResolveCount, *downloadCount, *totalDownloadCount)
				slog.Info(fmt.Sprintf("Done downloading tarball for %s@%s", pkg.Name, pkg.Version))
			}
		}()
	}
}
