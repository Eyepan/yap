package main

import (
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/downloader"
	"github.com/Eyepan/yap/src/fetcher"
	"github.com/Eyepan/yap/src/metadata"
	"github.com/Eyepan/yap/src/types"
)

type Task struct {
	Package types.Dependency
	Result  chan<- []types.Dependency
}

type MetadataTask struct {
	Package types.Dependency
	Result  chan<- types.VersionMetadata
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
	if len(args) > 0 {
		handleInstallArgs(args, &cache, npmrc)
	} else {
		installListOfDependencies(allDependencies, &cache, npmrc, false)
	}
}

func handleInstallArgs(args []string, cache *fetcher.FSCache, npmrc types.Config) {
	if len(args) == 1 && args[0] == "--force" {
		packageJSON, _ := config.ParsePackageJSON()
		allDependencies := gatherAllDependencies(packageJSON)
		installListOfDependencies(allDependencies, cache, npmrc, true)
	} else {
		for _, arg := range args {
			if strings.HasPrefix(arg, "--force") {
				installPackage(arg[7:], cache, npmrc, true)
			} else {
				installPackage(arg, cache, npmrc, false)
			}
		}
	}
}

func installListOfDependencies(dependencies []types.Dependency, cache *fetcher.FSCache, npmrc types.Config, force bool) {
	taskQueue := make(chan MetadataTask, len(dependencies))
	results := make(chan types.VersionMetadata, len(dependencies))
	var wg sync.WaitGroup

	startMetadataWorkers(taskQueue, results, cache, npmrc, &wg)
	startDownloadWorkers(results, &wg)

	// Enqueue tasks for metadata fetching
	for _, dep := range dependencies {
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
		taskQueue <- MetadataTask{Package: dep, Result: results}
	}

	close(taskQueue) // Close the task queue once all tasks are enqueued

	// Wait for all metadata fetching to complete
	wg.Wait()
	close(results) // Close results channel after all workers are done

	log.Println("All tasks completed.")
}

func installPackage(arg string, cache *fetcher.FSCache, npmrc types.Config, force bool) {
	parts := strings.Split(arg, "@")
	if len(parts) != 2 {
		log.Printf("Invalid package format: %s\n", arg)
		return
	}
	dep := types.Dependency{Name: parts[0], Version: parts[1]}

	if !force {
		isDownloaded, err := downloader.CheckIfPackageIsAlreadyDownloaded(dep.Name)
		if err != nil {
			log.Printf("Error checking if package is already downloaded: %v", err)
			return
		}
		if isDownloaded {
			log.Printf("Package %s is already downloaded.\n", dep.Name)
			return
		}
	}

	installListOfDependencies([]types.Dependency{dep}, cache, npmrc, force)
}

func startMetadataWorkers(taskQueue <-chan MetadataTask, results chan<- types.VersionMetadata, cache *fetcher.FSCache, npmrc types.Config, wg *sync.WaitGroup) {
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go metadataWorker(taskQueue, results, cache, npmrc, wg)
	}
}

func startDownloadWorkers(results <-chan types.VersionMetadata, wg *sync.WaitGroup) {
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go downloadWorker(results, wg)
	}
}

func metadataWorker(taskQueue <-chan MetadataTask, results chan<- types.VersionMetadata, cache *fetcher.FSCache, npmrc types.Config, wg *sync.WaitGroup) {
	for task := range taskQueue {
		metadata, err := metadata.FetchPackageMetadata(task.Package, cache, npmrc)
		if err != nil {
			log.Printf("Error fetching metadata for package %s: %v\n", task.Package.Name, err)
			wg.Done()
			continue
		}
		results <- metadata
		wg.Done()
	}
}

func downloadWorker(results <-chan types.VersionMetadata, wg *sync.WaitGroup) {
	for metadata := range results {
		wg.Add(1)
		go func(md types.VersionMetadata) {
			defer wg.Done()
			tarballURL := md.Dist.Tarball
			homeDir, err := os.UserHomeDir()
			if err != nil {
				log.Println("Error getting home directory location")
			}
			storeLocation := path.Join(homeDir, ".yap_store")
			if err := downloader.DownloadTarballAndExtract(tarballURL, md.Name, storeLocation); err != nil {
				log.Printf("Error downloading or extracting tarball for package %s: %v\n", md.Name, err)
			}
		}(metadata)
	}
}

func gatherAllDependencies(packageJSON types.PackageJSON) []types.Dependency {
	allDependencies := make([]types.Dependency, 0, len(packageJSON.Dependencies)+len(packageJSON.DevDependencies)+len(packageJSON.PeerDependencies))
	for key, value := range packageJSON.Dependencies {
		allDependencies = append(allDependencies, types.Dependency{Name: key, Version: value})
	}
	for key, value := range packageJSON.DevDependencies {
		allDependencies = append(allDependencies, types.Dependency{Name: key, Version: value})
	}
	for key, value := range packageJSON.PeerDependencies {
		allDependencies = append(allDependencies, types.Dependency{Name: key, Version: value})
	}
	return allDependencies
}
