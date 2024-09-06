package cli

import (
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/logger"
	"github.com/Eyepan/yap/src/types"
)

func HandleAdd() {
	args := os.Args
	if len(args) <= 2 {
		slog.Error("missing packages to install")
		os.Exit(-1)
	}
	config, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	stats := logger.Stats{}

	var lockBin types.Lockfile
	var lockBinMutex sync.Mutex

	packageName := args[2]
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
	if strings.ContainsRune(packageName, '@') {
		splits := strings.Split(packageName, "@")
		metadataWg.Add(1)
		metadataChannel <- &types.Package{Name: splits[0], Version: splits[1]}
	}

	metadataWg.Wait()
	close(metadataChannel)

	downloadWg.Wait()
	close(downloadChannel)

	// utils.WriteLock(lockBin)
	fmt.Println("\n💫 Done!")
}
