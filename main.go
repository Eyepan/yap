package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sync"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/downloader"
	"github.com/Eyepan/yap/src/metadata"
	"github.com/Eyepan/yap/src/packagejson"
	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

// init application
func Init() {

}

func Altmain() {
	config, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalln("failed while parsing the configs", err)
	}
	// parse package.json
	packageJSONfile, err := packagejson.ParsePackageJSON()
	if err != nil {
		log.Fatalln("failed while parsing package.json file", err)
	}
	slog.Info("Finished reading configs and package.json")
	fmt.Printf("config: %v\n", config)
	fmt.Printf("packageJSONfile: %v\n", packageJSONfile)

	typescriptPackage := types.Package{Name: "typescript", Version: "5.5.4"}
	slog.Info("Fetching metadata")
	vmd, err := metadata.FetchVersionMetadata(typescriptPackage, config, false)
	if err != nil {
		log.Fatalln("failed while fetching the metadata", err)
	}
	slog.Info("Fetched metadata")
	slog.Info("Downloading tarball")
	tarballData, err := downloader.DownloadTarball(vmd.Dist.Tarball, config)
	if err != nil {
		log.Fatalln("failed while downloading the tarball")
	}
	slog.Info("Downloaded tarball")
	slog.Info("Extracting tarball")
	err = downloader.ExtractTarball(tarballData, fmt.Sprintf("%s@%s", typescriptPackage.Name, typescriptPackage.Version))
	if err != nil {
		log.Fatalln("failed while extracting the tarball")
	}
	slog.Info("Extracted tarball")
}

func main() {
	// Step 1: Load configurations
	config, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	// Step 2: Parse the command input (for now, assume 'install')
	command := "install" // this would be dynamically set by your CLI parser
	if command == "install" {
		// check if lockfile already exists
		// if check, _ := utils.DoesLockfileExist(); check {
		// 	lockbin, _ := utils.ReadLock()
		// 	lockfileJSON, err := json.MarshalIndent(lockbin, "", "  ")
		// 	if err != nil {
		// 		log.Fatalf("Failed to read lockfile: %v", err)
		// 	}

		// 	// Print the JSON string
		// 	fmt.Println(string(lockfileJSON))
		// 	os.Exit(0)
		// }

		// Step 3: Parse package.json to get core dependencies
		pkgJSON, err := packagejson.ParsePackageJSON()
		if err != nil {
			log.Fatalf("Failed to parse package.json: %v", err)
		}

		// Step 4: Resolve dependencies concurrently
		var lockbin types.Lockfile
		lockbin.CoreDependencies = []types.Package{}

		// var mu sync.Mutex
		var wg sync.WaitGroup
		results := make(chan *types.MPackage)

		// Start resolving core dependencies
		for name, version := range pkgJSON.Dependencies {
			pkg := types.Package{Name: name, Version: version}
			lockbin.CoreDependencies = append(lockbin.CoreDependencies, pkg)

			wg.Add(1)
			go func(pkg types.Package) {
				defer wg.Done()
				resolved, err := resolvePackage(pkg, config)
				if err != nil {
					log.Printf("Failed to resolve package %s: %v", pkg.Name, err)
					return
				}
				results <- resolved
			}(pkg)
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		// Build the lockfile
		lockbin.Resolutions = []*types.MPackage{}
		for res := range results {
			lockbin.Resolutions = append(lockbin.Resolutions, res)
		}

		utils.WriteLock(lockbin)

		// TODO: Save or use the lockfile as needed
		lockfileJSON, err := json.MarshalIndent(lockbin, "", "  ")
		if err != nil {
			log.Fatalf("Failed to marshal lockfile: %v", err)
		}

		// Print the JSON string
		fmt.Println(string(lockfileJSON))

	}
}

func resolvePackage(pkg types.Package, config types.Config) (*types.MPackage, error) {
	slog.Info(fmt.Sprintf("Fetching metadata for %s@%s", pkg.Name, pkg.Version))
	vmd, err := metadata.FetchVersionMetadata(pkg, config, false)
	if err != nil {
		log.Fatalln("failed while fetching the metadata", err)
	}
	slog.Info(fmt.Sprintf("Done fetching metadata for %s@%s", pkg.Name, pkg.Version))
	downloader.DownloadPackage(types.Package{Name: vmd.Name, Version: vmd.Version}, vmd.Dist.Tarball, config, false)
	if err != nil {
		return nil, err
	}

	resolvedPkg := &types.MPackage{
		Name:         vmd.Name,
		Version:      vmd.Version,
		Id:           vmd.ID,
		Dist:         vmd.Dist,
		Dependencies: []*types.MPackage{},
	}

	for depName, depVersion := range vmd.Dependencies {
		depPkg := types.Package{Name: depName, Version: depVersion}
		subResolved, err := resolvePackage(depPkg, config)
		if err != nil {
			return nil, err
		}
		resolvedPkg.Dependencies = append(resolvedPkg.Dependencies, subResolved)
	}

	return resolvedPkg, nil
}
