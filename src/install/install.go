package install

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/downloader"
	"github.com/Eyepan/yap/src/logger"
	"github.com/Eyepan/yap/src/metadata"
	"github.com/Eyepan/yap/src/ship"
	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

var client = &http.Client{
	Timeout: time.Second * 30,
}

func InstallPackages(listOfPackages *types.Dependencies, force bool) {
	config, err := config.ReadYapConfig()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}
	stats := logger.Stats{}

	lockFile := types.Lockfile{}
	var lockMutex sync.Mutex
	baseDependencies := *listOfPackages

	numWorkers := 200
	slog.Info(fmt.Sprintf("Running on %d CPU Cores", numWorkers))

	var metadataWg sync.WaitGroup
	var downloadWg sync.WaitGroup

	metadataChannel := make(chan *types.Package)
	downloadChannel := make(chan *types.MPackage)

	installedPackages := sync.Map{} // Track installed packages to avoid redundant work

	// Worker for resolving package metadata
	for i := 0; i < numWorkers; i++ {
		go func() {
			for pkg := range metadataChannel {
				ResolvePackageMetadata(&metadataWg, &downloadWg, pkg, config, downloadChannel, metadataChannel, &stats, &installedPackages, force)
			}
		}()
	}

	// Worker for downloading packages
	for i := 0; i < numWorkers; i++ {
		go func() {
			for mPkg := range downloadChannel {
				DownloadPackageTarball(&downloadWg, mPkg, config, &stats)
				lockMutex.Lock()
				lockFile.Resolutions = append(lockFile.Resolutions, *mPkg)
				lockMutex.Unlock()
			}
		}()
	}

	lockfileExists, _ := utils.DoesLockfileExist()
	var shouldCreateLockFile bool
	if lockfileExists && !force {
		lockBin, err := utils.ReadLock()
		if err == nil {
			// Use the lockfile if valid
			slog.Info("[INSTALL] Lockfile detected, skipping metadata resolution step")
			for _, resolution := range lockBin.Resolutions {
				installedPackages.Store(types.Package{Name: resolution.Name, Version: resolution.Version}, true) // Mark packages as installed
				downloadWg.Add(1)
				downloadChannel <- &resolution
				stats.IncrementResolveCount()
				stats.IncrementTotalResolveCount()
				stats.IncrementTotalDownloadCount()
				stats.IncrementTotalMoveCount()
			}
		} else {
			shouldCreateLockFile = true
			slog.Warn("[INSTALL] Invalid lockfile, rewriting it now.")
		}
	} else {
		shouldCreateLockFile = true
		slog.Warn("[INSTALL] No lockfile detected, creating one now.")
	}

	if shouldCreateLockFile || force {
		metadataWg.Add(len(baseDependencies))
		for name, version := range baseDependencies {
			basePackage := types.Package{Name: name, Version: version}
			if _, loaded := installedPackages.LoadOrStore(basePackage, true); !loaded {
				lockMutex.Lock()
				lockFile.CoreDependencies = append(lockFile.CoreDependencies, basePackage)
				lockMutex.Unlock()
				metadataChannel <- &basePackage
				stats.IncrementTotalResolveCount()
				stats.IncrementTotalDownloadCount()
				stats.IncrementTotalMoveCount()
			} else {
				metadataWg.Done() // If already installed, reduce the waitgroup counter
			}
		}
	}

	metadataWg.Wait()
	close(metadataChannel)

	downloadWg.Wait()
	close(downloadChannel)

	// Write lockfile if necessary
	if shouldCreateLockFile {
		utils.WriteLock(lockFile)
	}

	fmt.Println("\n💫 Done!")
}

func ResolvePackageMetadata(metadataWg, downloadWg *sync.WaitGroup, pkg *types.Package, config *types.YapConfig, downloadChannel chan<- *types.MPackage, metadataChannel chan<- *types.Package, stats *logger.Stats, installedPackages *sync.Map, force bool) {
	defer metadataWg.Done()
	defer stats.IncrementResolveCount()
	if _, loaded := installedPackages.LoadOrStore(pkg, true); loaded {
		return
	}

	slog.Info(fmt.Sprintf("[METADATA] 🔃 %s@%s", pkg.Name, pkg.Version))

	vmd, err := metadata.FetchVersionMetadata(client, pkg, config, force)
	if err != nil {
		slog.Error(fmt.Sprintf("[METADATA] ❌ %s@%s\t%v", pkg.Name, pkg.Version, err))
		return
	}

	var dependencies = make([]types.Package, 0, len(vmd.Dependencies))
	for name, version := range vmd.Dependencies {
		dependencies = append(dependencies, types.Package{Name: name, Version: version}) // this append doesn't have the append overhead in slices as make() already takes in the length and there's no need to reallocate memory for the entire slice!
	}

	// Package metadata resolved, queue for download
	packageToBeDownloaded := types.MPackage{
		Name:         vmd.Name,
		Version:      vmd.Version,
		Dist:         vmd.Dist,
		Dependencies: dependencies,
	}

	downloadWg.Add(1)
	downloadChannel <- &packageToBeDownloaded

	// Resolve dependencies recursively
	for _, dep := range packageToBeDownloaded.Dependencies {
		if _, alreadyInstalled := installedPackages.LoadOrStore(dep, true); !alreadyInstalled {
			stats.IncrementTotalResolveCount()
			stats.IncrementTotalDownloadCount()
			stats.IncrementTotalMoveCount()
			metadataWg.Add(1)
			go func(dep *types.Package) {
				metadataChannel <- dep
			}(&dep)
		}
	}
}

func DownloadPackageTarball(downloadWg *sync.WaitGroup, mPkg *types.MPackage, config *types.YapConfig, stats *logger.Stats) {
	defer downloadWg.Done()
	defer stats.IncrementDownloadCount()
	slog.Info(fmt.Sprintf("[TARBALL] 🚚 %s@%s", mPkg.Name, mPkg.Version))

	if err := downloader.DownloadPackage(client, &types.Package{Name: mPkg.Name, Version: mPkg.Version}, &mPkg.Dist.Tarball, config, false); err != nil {
		slog.Error(fmt.Sprintf("[TARBALL] ❌ %s@%s\t%v", mPkg.Name, mPkg.Version, err))
		return
	}

	slog.Info(fmt.Sprintf("[TARBALL] ✅ %s@%s", mPkg.Name, mPkg.Version))

	ship.InstallPackageToDotYap(mPkg, config, stats)
	stats.IncrementMoveCount()
}
