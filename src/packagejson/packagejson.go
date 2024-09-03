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
