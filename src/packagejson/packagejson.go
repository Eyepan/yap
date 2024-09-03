package packagejson

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Eyepan/yap/src/types"
)

// ParsePackageJSON reads and parses package.json.
func ParsePackageJSON() (types.PackageJSON, error) {
	filePath := filepath.Join(".", "package.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return types.PackageJSON{}, err
	}

	var pkgJSON types.PackageJSON
	if err := json.Unmarshal(data, &pkgJSON); err != nil {
		return types.PackageJSON{}, err
	}

	return pkgJSON, nil
}

func GetAllDependencies(pkgJSON *types.PackageJSON) types.Dependencies {
	deps := make(types.Dependencies)
	for name, version := range pkgJSON.PeerDependencies {
		deps[name] = version
	}
	for name, version := range pkgJSON.DevDependencies {
		deps[name] = version
	}
	for name, version := range pkgJSON.Dependencies {
		deps[name] = version
	}
	return deps
}
