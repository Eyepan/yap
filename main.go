package main

import (
	"fmt"
	"log"
	"runtime"
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

// MetadataTask represents a task for fetching metadata and downloading tarballs.
type MetadataTask struct {
	Package types.Dependency
	Result  chan<- types.VersionMetadata
}

func main() {
	cache := fetcher.FSCache{CacheDir: "cache"}

	npmrc, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalln("Error parsing config:", err)
	}

	packageJSON, err := config.ParsePackageJSON()
	if err != nil {
		log.Fatalln("Error parsing package.json:", err)
	}

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

	taskQueue := make(chan MetadataTask, len(allDependencies))
	results := make(chan types.VersionMetadata, len(allDependencies))
	var wg sync.WaitGroup

	// Start worker goroutines for fetching metadata
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go metadataWorker(taskQueue, results, &cache, npmrc, &wg)
	}

	// Start worker goroutines for downloading tarballs
	numDownloadWorkers := runtime.NumCPU()
	for i := 0; i < numDownloadWorkers; i++ {
		go downloadWorker(results, &wg)
	}

	// Enqueue tasks for metadata fetching
	for _, dep := range allDependencies {
		wg.Add(1)
		resultChan := make(chan types.VersionMetadata)
		taskQueue <- MetadataTask{Package: dep, Result: resultChan}
	}

	close(taskQueue)

	// Wait for all tasks to be completed
	wg.Wait()
	close(results)

	fmt.Println("All tasks completed.")
}

// metadataWorker processes tasks from the queue for fetching metadata.
func metadataWorker(taskQueue <-chan MetadataTask, results chan<- types.VersionMetadata, cache *fetcher.FSCache, npmrc types.Config, wg *sync.WaitGroup) {
	for task := range taskQueue {
		metadata, err := metadata.FetchPackageMetadata(task.Package, cache, npmrc)
		if err != nil {
			fmt.Printf("Error fetching metadata for package %s: %v\n", task.Package.Name, err)
			wg.Done()
			continue
		}
		results <- metadata
		wg.Done()
	}
}

// downloadWorker concurrently downloads and extracts tarballs based on metadata.
func downloadWorker(results <-chan types.VersionMetadata, wg *sync.WaitGroup) {
	for metadata := range results {
		wg.Add(1)
		go func(md types.VersionMetadata) {
			defer wg.Done()
			tarballURL := md.Dist.Tarball
			if err := downloader.DownloadTarballAndExtract(tarballURL, md.Name); err != nil {
				fmt.Printf("Error downloading or extracting tarball for package %s: %v\n", md.Name, err)
			}
		}(metadata)
	}
}
