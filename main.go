package main

import (
	"log"
	"os"
	"path"
	"runtime"
	"sync"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/downloader"
	"github.com/Eyepan/yap/src/fetcher"
	"github.com/Eyepan/yap/src/metadata"
	"github.com/Eyepan/yap/src/types"
)

type MetadataTask struct {
	Package types.Package
	Result  chan<- DownloadTask
}

type DownloadTask struct {
	Metadata types.VersionMetadata
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Expected 'install' command")
		return
	}

	command := os.Args[1]

	switch command {
	case "install":
		handleInstall()
	default:
		log.Printf("Unknown command: %s\n", command)
	}
}

func handleInstall() {
	cache := fetcher.FSCache{CacheDir: "cache"}

	npmrc, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalf("Error parsing config: %v", err)
	}

	packageJSON, err := config.ParsePackageJSON()
	if err != nil {
		log.Fatalf("Error parsing package.json: %v", err)
	}

	allDependencies := gatherAllDependencies(packageJSON)
	args := os.Args[2:]
	if len(args) > 0 && args[0] == "--force" {
		installListOfPackages(allDependencies, &cache, npmrc, true)
	} else {
		installListOfPackages(allDependencies, &cache, npmrc, false)
	}
}

func installListOfPackages(packages []types.Package, cache *fetcher.FSCache, npmrc types.Config, force bool) {
	metadataQueue := make(chan MetadataTask, len(packages))
	downloadQueue := make(chan DownloadTask, len(packages))
	var wg sync.WaitGroup

	startMetadataWorkers(metadataQueue, downloadQueue, cache, npmrc, &wg)
	startDownloadWorkers(downloadQueue)

	for _, dep := range packages {
		// Check if the package is already downloaded
		if !force {
			isDownloaded, err := downloader.CheckIfPackageIsAlreadyDownloaded(dep.Name)
			if err != nil {
				log.Printf("Error checking if package %s is already downloaded: %v", dep.Name, err)
				continue
			}
			if isDownloaded {
				log.Printf("Package %s is already downloaded. Skipping installation.", dep.Name)
				continue
			}
		}
		wg.Add(1)
		metadataQueue <- MetadataTask{Package: dep, Result: downloadQueue}
	}
	wg.Wait()
	close(metadataQueue)
	close(downloadQueue)
}

func startMetadataWorkers(metadataTaskQueue chan MetadataTask, downloadTaskQueue chan<- DownloadTask, cache *fetcher.FSCache, npmrc types.Config, wg *sync.WaitGroup) {
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go func(metadataTaskQueue chan MetadataTask, downloadQueue chan<- DownloadTask) {
			for task := range metadataTaskQueue {
				md, err := metadata.FetchPackageMetadata(task.Package, cache, npmrc)
				if err != nil {
					log.Printf("Error fetching metadata for package %s: %v\n", task.Package.Name, err)
					wg.Done()
					continue
				}
				dependencies := metadata.GetSubdependenciesFromMetadata(md)
				for _, dep := range dependencies {
					metadataTaskQueue <- MetadataTask{Package: dep, Result: downloadQueue}
					wg.Add(1)
				}
				downloadQueue <- DownloadTask{Metadata: md}
				wg.Done()
			}
		}(metadataTaskQueue, downloadTaskQueue)
	}
}

func startDownloadWorkers(downloadTaskQueue <-chan DownloadTask) {
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go func(downloadTaskQueue <-chan DownloadTask) {
			for md := range downloadTaskQueue {
				tarballURL := md.Metadata.Dist.Tarball
				homeDir, err := os.UserHomeDir()
				if err != nil {
					log.Println("Error getting home directory")
				}
				storeLocation := path.Join(homeDir, ".yap_store")
				if err := downloader.DownloadTarballAndExtract(tarballURL, md.Metadata.Name, storeLocation); err != nil {
					log.Printf("Error downloading tarball for package %s: %v\n", md.Metadata.Name, err)
				}
				// log.Printf("Downloaded %s@%s", md.Metadata.Name, md.Metadata.Version)
			}
		}(downloadTaskQueue)
	}
}

func gatherAllDependencies(packageJSON types.PackageJSON) []types.Package {
	allDependencies := make([]types.Package, 0, len(packageJSON.Dependencies)+len(packageJSON.DevDependencies)+len(packageJSON.PeerDependencies))
	for key, value := range packageJSON.Dependencies {
		allDependencies = append(allDependencies, types.Package{Name: key, Version: value})
	}
	for key, value := range packageJSON.DevDependencies {
		allDependencies = append(allDependencies, types.Package{Name: key, Version: value})
	}
	for key, value := range packageJSON.PeerDependencies {
		allDependencies = append(allDependencies, types.Package{Name: key, Version: value})
	}
	return allDependencies
}
