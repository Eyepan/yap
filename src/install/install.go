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
	"github.com/Eyepan/yap/src/types"
)

var client = &http.Client{
	Timeout: time.Second * 30,
}

func InstallPackages(listOfPackages *types.Dependencies) {
	config, err := config.ReadYapConfig()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}
	stats := logger.Stats{}

	baseDependencies := *listOfPackages

	numWorkers := 200 // arbitrarily huge number
	slog.Info(fmt.Sprintf("Running on %d CPU Cores", numWorkers))

	var metadataWg sync.WaitGroup
	var downloadWg sync.WaitGroup

	metadataChannel := make(chan *types.Package)
	downloadChannel := make(chan *types.MPackage)
	// install map
	var installedPackages sync.Map

	for i := 0; i < numWorkers; i++ {
		go func() {
			for pkg := range metadataChannel {
				ResolvePackageMetadata(&metadataWg, &downloadWg, pkg, config, downloadChannel, metadataChannel, &stats, &installedPackages)
			}
		}()
	}

	for i := 0; i < numWorkers; i++ {
		go func() {
			for mPkg := range downloadChannel {
				DownloadPackageTarball(&downloadWg, mPkg, config, &stats)
			}
		}()
	}

	metadataWg.Add(len(baseDependencies))
	for name, version := range baseDependencies {
		go func(name, version string) {
			basePackage := types.Package{Name: name, Version: version}
			metadataChannel <- &basePackage
			stats.IncrementTotalResolveCount()
		}(name, version)
	}

	metadataWg.Wait()
	close(metadataChannel)

	downloadWg.Wait()
	close(downloadChannel)

	fmt.Println("\n💫 Done!")
}

func ResolvePackageMetadata(metadataWg, downloadWg *sync.WaitGroup, pkg *types.Package, config *types.YapConfig, downloadChannel chan<- *types.MPackage, metadataChannel chan<- *types.Package, stats *logger.Stats, installedPackages *sync.Map) {
	defer metadataWg.Done()
	if _, loaded := installedPackages.LoadOrStore(fmt.Sprintf("%s@%s", pkg.Name, pkg.Version), true); loaded {
		stats.IncrementResolveCount()
		return
	}
	slog.Info(fmt.Sprintf("[METADATA] 🔃 %s@%s", pkg.Name, pkg.Version))

	vmd, err := metadata.FetchVersionMetadata(client, pkg, config, false)
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
		// Dependencies: vmd.Dependencies, // make this return a types.MPackage and redo this?
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

func DownloadPackageTarball(downloadWg *sync.WaitGroup, mPkg *types.MPackage, config *types.YapConfig, stats *logger.Stats) {
	defer downloadWg.Done()
	slog.Info(fmt.Sprintf("[TARBALL] 🚚 %s@%s", mPkg.Name, mPkg.Version))

	if err := downloader.DownloadPackage(client, &types.Package{Name: mPkg.Name, Version: mPkg.Version}, &mPkg.Dist.Tarball, config, false); err != nil {
		slog.Error(fmt.Sprintf("[TARBALL] ❌ %s@%s\t%v", mPkg.Name, mPkg.Version, err))
		return
	}

	stats.IncrementDownloadCount()
	slog.Info(fmt.Sprintf("[TARBALL] ✅ %s@%s", mPkg.Name, mPkg.Version))
}
