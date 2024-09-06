package cli

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

func HandleInstall() {
	// TODO: Read lockfile if exists, and use it for faster resolution
	config, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}
	pkgJSON, err := packagejson.ParsePackageJSON()
	if err != nil {
		log.Fatalf("Failed to parse package.json: %v", err)
	}
	stats := logger.Stats{}

	var lockBin types.Lockfile
	var lockBinMutex sync.Mutex

	baseDependencies := packagejson.GetAllDependencies(&pkgJSON)

	numWorkers := runtime.NumCPU()
	slog.Info(fmt.Sprintf("Running on %d CPU Cores", numWorkers))

	var metadataWg sync.WaitGroup
	var downloadWg sync.WaitGroup

	metadataChannel := make(chan *types.Package)
	downloadChannel := make(chan *types.MPackage)

	for i := 0; i < numWorkers; i++ {
		go func() {
			for pkg := range metadataChannel {
				DownloadPackageMetadata(&metadataWg, &downloadWg, pkg, &config, downloadChannel, metadataChannel, &stats)
			}
		}()
	}

	for i := 0; i < numWorkers; i++ {
		go func() {
			for mPkg := range downloadChannel {
				DownloadPackageTarball(&downloadWg, mPkg, &config, &stats, &lockBin, &lockBinMutex)
			}
		}()
	}

	metadataWg.Add(len(baseDependencies))
	for name, version := range baseDependencies {
		go func(name, version string) {
			basePackage := types.Package{Name: name, Version: version}
			metadataChannel <- &basePackage
			lockBinMutex.Lock()
			lockBin.CoreDependencies = append(lockBin.CoreDependencies, basePackage)
			lockBinMutex.Unlock()
			stats.IncrementTotalResolveCount()
		}(name, version)
	}

	metadataWg.Wait()
	close(metadataChannel)

	downloadWg.Wait()
	close(downloadChannel)

	utils.WriteLock(lockBin)
	fmt.Println("\n💫 Done!")
}

func DownloadPackageMetadata(metadataWg, downloadWg *sync.WaitGroup, pkg *types.Package, config *types.Config, downloadChannel chan<- *types.MPackage, metadataChannel chan<- *types.Package, stats *logger.Stats) {
	defer metadataWg.Done()
	slog.Info(fmt.Sprintf("[METADATA] 🔃 %s@%s", pkg.Name, pkg.Version))

	vmd, err := metadata.FetchVersionMetadata(pkg, config, false)
	stats.IncrementResolveCount()

	if err != nil {
		slog.Error(fmt.Sprintf("[METADATA] ❌ %s@%s\t%v", pkg.Name, pkg.Version, err))
		return
	}

	slog.Info(fmt.Sprintf("[METADATA] ✅ %s@%s", vmd.Name, vmd.Version))

	packageToBeDownloaded := types.MPackage{
		Name:    vmd.Name,
		Version: vmd.Version,
		Dist:    vmd.Dist,
	}
	stats.IncrementTotalDownloadCount()

	downloadWg.Add(1)
	downloadChannel <- &packageToBeDownloaded

	for depName, depVersion := range vmd.Dependencies {
		depPkg := &types.Package{Name: depName, Version: depVersion}
		stats.IncrementTotalResolveCount()

		metadataWg.Add(1)
		go func(depPkg *types.Package) {
			metadataChannel <- depPkg
		}(depPkg)
	}
}

func DownloadPackageTarball(downloadWg *sync.WaitGroup, mPkg *types.MPackage, config *types.Config, stats *logger.Stats, lockBin *types.Lockfile, lockBinMutex *sync.Mutex) {
	defer downloadWg.Done()
	lockBinMutex.Lock()
	lockBin.Resolutions = append(lockBin.Resolutions, *mPkg)
	lockBinMutex.Unlock()
	slog.Info(fmt.Sprintf("[TARBALL] 🚚 %s@%s", mPkg.Name, mPkg.Version))

	if err := downloader.DownloadPackage(&types.Package{Name: mPkg.Name, Version: mPkg.Version}, &mPkg.Dist.Tarball, config, false); err != nil {
		slog.Error(fmt.Sprintf("[TARBALL] ❌ %s@%s\t%v", mPkg.Name, mPkg.Version, err))
	} else {
		stats.IncrementDownloadCount()
		slog.Info(fmt.Sprintf("[TARBALL] ✅ %s@%s", mPkg.Name, mPkg.Version))
	}
}
