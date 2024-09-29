package ship

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

func InstallPackageToDotYap(mPkg *types.MPackage, config *types.YapConfig) {
	// Check if package is already installed
	if check, _ := CheckIfPackageIsAlreadyInstalled(mPkg); check {
		slog.Info(fmt.Sprintf("[SHIP] Package %s@%s already installed", mPkg.Name, mPkg.Version))
		return
	}

	// Initialize .yap directory
	dotYapDir, err := utils.GetDotYapDir()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to initialize the .yap directory inside node_modules:\t %v", err))
		return
	}

	// Create package directory
	packageDir := filepath.Join(dotYapDir, utils.SanitizePackageName(fmt.Sprintf("%s@%s", mPkg.Name, mPkg.Version)))
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		slog.Error(fmt.Sprintf("failed to create package directory: %v", err))
		return
	}

	// Get store directory
	storeDir, err := utils.GetStoreDir()
	if err != nil {
		slog.Error(fmt.Sprintf("failed to get store directory: %v", err))
		return
	}

	// Define source package directory
	sourcePackageDir := filepath.Join(storeDir, utils.SanitizePackageName(fmt.Sprintf("%s@%s", mPkg.Name, mPkg.Version)))

	// Try to hardlink files recursively, fallback to copying if it fails
	err = utils.HardLinkOrCopyRecursively(sourcePackageDir, packageDir)
	if err != nil {
		slog.Error(fmt.Sprintf("failed to hardlink or copy files: %v", err))
		return
	}

	slog.Info(fmt.Sprintf("[SHIP] ✅ %s@%s", mPkg.Name, mPkg.Version))
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
