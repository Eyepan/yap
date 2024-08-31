package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Eyepan/yap/config"
	"github.com/Eyepan/yap/dependencies"
	"github.com/Eyepan/yap/fetcher"
	"github.com/Eyepan/yap/resolver"
)

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	packages, err := dependencies.ParsePackages()
	fetchCache := fetcher.NewCache()

	if err != nil {
		log.Fatalln("parsing packages failed\n", err)
	}
	npmrc, err := config.LoadConfigurations()
	if err != nil {
		log.Fatalln("loading configurations failed\n", err)
	}
	// fmt.Println(packages, npmrc, authToken)

	if len(os.Args) < 2 {
		fmt.Println("Usage: install <package-name>@<version> or install to install dependencies from package.json")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "install":
		fmt.Println("not implemented yet")
	case "list":
		fmt.Println("Packages to be installed", packages)
	case "test":
		metadata, err := resolver.FetchPackageMetadata("express", "latest", npmrc, fetchCache)
		if err != nil {
			log.Fatalln("fetching package metadata failed\n", err)
		}
		subdependencies, err := resolver.GetSubdependencies(metadata)
		if err != nil {
			log.Fatalf("getting subdependencies for package %s failed \n%s", metadata.Name, err)
		}
		allSubdependencies, err := resolver.GetAllSubdependencies(subdependencies, npmrc, fetchCache)
		if err != nil {
			log.Fatalf("getting all subdependencies for package %s failed \n%s", subdependencies, err)
		}
		fmt.Println("All subdependencies:", allSubdependencies)
	default:
		fmt.Println("Unknown command")
	}
}
