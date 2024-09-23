package ship

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Eyepan/yap/src/logger"
	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

func InstallPackageToDotYap(mPkg *types.MPackage, config *types.YapConfig, stats *logger.Stats) {
	if check, _ := CheckIfPackageIsAlreadyInstalled(mPkg); check {
		slog.Info(fmt.Sprintf("[SHIP] Package %s@%s already installed", mPkg.Name, mPkg.Version))
		return
	}

	dotYapDir, err := utils.GetDotYapDir()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to initialize the .yap directory inside node_modules:\t %v", err))
		return
	}

	packageDir := filepath.Join(dotYapDir, utils.SanitizePackageName(fmt.Sprintf("%s@%s", mPkg.Name, mPkg.Version)))

	if err := os.MkdirAll(packageDir, 0755); err != nil {
		slog.Error(fmt.Sprintf("failed to create package directory: %v", err))
		return
	}

	// At this point, the package is definitely in the cache, hardlinking from the cache should be fine
	// get the package from the cache

	storeDir, err := utils.GetStoreDir()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to get store directory: %v", err))
		return
	}
	// create dir for package
	sourcePackageDir := filepath.Join(storeDir, utils.SanitizePackageName(fmt.Sprintf("%s@%s", mPkg.Name, mPkg.Version)))

	if err := utils.HardLinkTwoDirectories(sourcePackageDir, packageDir); err != nil {
		slog.Error(fmt.Sprintf("failed to hardlink files: %v", err))
	}

	slog.Info(fmt.Sprintf("[SHIP] ✅ %s@%s", mPkg.Name, mPkg.Version))
	return
}

func CheckIfPackageIsAlreadyInstalled(mPkg *types.MPackage) (bool, error) {
	dotYapDir, err := utils.GetDotYapDir()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to initialize the .yap directory inside node_modules:\t %v", err))
		return false, err
	}

	packageDir := filepath.Join(dotYapDir, utils.SanitizePackageName(fmt.Sprintf("%s@%s", mPkg.Name, mPkg.Version)))

	if _, err := os.Stat(packageDir); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check package path: %w", err)
	}

	return true, nil
}
