package cli

import (
	"log"

	"github.com/Eyepan/yap/src/install"
	"github.com/Eyepan/yap/src/packagejson"
)

func HandleInstall() {
	pkgJSON, err := packagejson.ParsePackageJSON()
	if err != nil {
		log.Fatalf("Failed to parse package.json: %v", err)
	}

	baseDependencies := packagejson.GetAllDependencies(&pkgJSON)
	install.InstallPackages(&baseDependencies)
}
