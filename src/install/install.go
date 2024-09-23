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

	numWorkers := 200 // arbitrarily huge number
	slog.Info(fmt.Sprintf("Running on %d CPU Cores", numWorkers))

	var metadataWg sync.WaitGroup
	var downloadWg sync.WaitGroup

	metadataChannel := make(chan *types.Package)
	downloadChannel := make(chan *types.MPackage)
	// nmChannel := make(chan *types.MPackage)

	var installedPackages sync.Map

	for i := 0; i < numWorkers; i++ {
		go func() {
			for pkg := range metadataChannel {
				ResolvePackageMetadata(&metadataWg, &downloadWg, pkg, config, downloadChannel, metadataChannel, &stats, &installedPackages, force)
			}
		}()
	}

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
		shouldCreateLockFile = false
		lockBin, err := utils.ReadLock()
		if err == nil {
			slog.Info("[INSTALL] Lockfile detected, skipping metadata resolution step")
			slog.Info(fmt.Sprintf("[INSTALL] There are %d packages in the lockfile to be installed", len(lockBin.Resolutions)))
			stats.TotalDownloadCount = len(lockBin.Resolutions)
			stats.TotalResolveCount = len(lockBin.Resolutions)
			stats.TotalMoveCount = len(lockBin.Resolutions)
			stats.ResolveCount = len(lockBin.Resolutions)
			for i := range lockBin.Resolutions {
				downloadWg.Add(1)
				downloadChannel <- &lockBin.Resolutions[i]
			}
		} else {
			shouldCreateLockFile = true
			slog.Warn("[INSTALL] Invalid lockfile, rewriting it now.")
		}
	} else {
		shouldCreateLockFile = true
		slog.Warn("[INSTALL] No lockfile detected, creating one now")
	}

	if shouldCreateLockFile || force {
		metadataWg.Add(len(baseDependencies))
		for name, version := range baseDependencies {
			go func(name, version string) {
				basePackage := types.Package{Name: name, Version: version}
				lockMutex.Lock()
				lockFile.CoreDependencies = append(lockFile.CoreDependencies, basePackage)
				lockMutex.Unlock()
				metadataChannel <- &basePackage
				stats.IncrementTotalResolveCount()
			}(name, version)
		}
	}

	metadataWg.Wait()
	close(metadataChannel)

	downloadWg.Wait()
	close(downloadChannel)

	// Write lockfile
	if shouldCreateLockFile {
		utils.WriteLock(lockFile)
	}

	fmt.Println("\n💫 Done!")
}

func ResolvePackageMetadata(metadataWg, downloadWg *sync.WaitGroup, pkg *types.Package, config *types.YapConfig, downloadChannel chan<- *types.MPackage, metadataChannel chan<- *types.Package, stats *logger.Stats, installedPackages *sync.Map, force bool) {
	defer metadataWg.Done()
	if _, loaded := installedPackages.LoadOrStore(pkg, true); loaded {
		stats.IncrementResolveCount()
		return
	}
	slog.Info(fmt.Sprintf("[METADATA] 🔃 %s@%s", pkg.Name, pkg.Version))

	vmd, err := metadata.FetchVersionMetadata(client, pkg, config, force)
	stats.IncrementResolveCount()

	if err != nil {
		slog.Error(fmt.Sprintf("[METADATA] ❌ %s@%s\t%v", pkg.Name, pkg.Version, err))
		return
	}

	slog.Info(fmt.Sprintf("[METADATA] ✅ %s@%s", vmd.Name, vmd.Version))

	var dependencies = make([]types.Package, 0, len(vmd.Dependencies))
	for name, version := range vmd.Dependencies {
		dependencies = append(dependencies, types.Package{Name: name, Version: version}) // this append doesn't have the append overhead in slices as make() already takes in the length and there's no need to reallocate memory for the entire slice!
	}

	packageToBeDownloaded := types.MPackage{
		Name:         vmd.Name,
		Version:      vmd.Version,
		Dist:         vmd.Dist,
		Dependencies: dependencies,
	}
	stats.IncrementTotalDownloadCount()
	stats.IncrementTotalMoveCount()

	downloadWg.Add(1)
	downloadChannel <- &packageToBeDownloaded

	for depName, depVersion := range vmd.Dependencies {
		depPkg := types.Package{Name: depName, Version: depVersion}
		if _, alreadyInstalled := installedPackages.Load(depPkg); alreadyInstalled {
			continue // skipping adding dependency if dependency is already installed
		}
		stats.IncrementTotalResolveCount()

		metadataWg.Add(1)
		go func(depPkg *types.Package) {
			metadataChannel <- depPkg
		}(&depPkg)
	}
}

func DownloadPackageTarball(downloadWg *sync.WaitGroup, mPkg *types.MPackage, config *types.YapConfig, stats *logger.Stats) {
	defer downloadWg.Done()
	slog.Info(fmt.Sprintf("[TARBALL] 🚚 %s@%s", mPkg.Name, mPkg.Version))

	if err := downloader.DownloadPackage(client, &types.Package{Name: mPkg.Name, Version: mPkg.Version}, &mPkg.Dist.Tarball, config, false); err != nil {
		slog.Error(fmt.Sprintf("[TARBALL] ❌ %s@%s\t%v", mPkg.Name, mPkg.Version, err))
		return
	}

	stats.IncrementDownloadCount()
	slog.Info(fmt.Sprintf("[TARBALL] ✅ %s@%s", mPkg.Name, mPkg.Version))

	ship.InstallPackageToDotYap(mPkg, config, stats)
	stats.IncrementMoveCount()
}
