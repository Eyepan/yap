package dependencies

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

// this file contains functions associated with parsing package.json, listing out all the packages and (if possible) listing out the links to fetch metadata & tgz files from

type PackageJSON struct {
	Name             string            `json:"name"`
	Module           string            `json:"module"`
	Type             string            `json:"type"`
	DevDependencies  map[string]string `json:"devDependencies"`
	PeerDependencies map[string]string `json:"peerDependencies"`
	Dependencies     map[string]string `json:"dependencies"`
}

type Package struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// reads package.json and returns package slice
func ParsePackages() ([]Package, error) {
	filePath := path.Join(".", "package.json")
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading package.json failed: \n%e", err)
	}
	var data PackageJSON
	err = json.Unmarshal(fileBytes, &data)
	if err != nil {
		return nil, fmt.Errorf("parsing package.json failed: \n%e", err)
	}

	var packages []Package

	for k, v := range data.Dependencies {
		packages = append(packages, Package{Name: k, Version: v})
	}

	for k, v := range data.DevDependencies {
		packages = append(packages, Package{Name: k, Version: v})
	}

	for k, v := range data.PeerDependencies {
		packages = append(packages, Package{Name: k, Version: v})
	}

	return packages, nil
}
