package cli

import (
	"fmt"
	"log"

	"github.com/Eyepan/yap/src/utils"
)

func HandleList() {
	lockBin, err := utils.ReadLock()
	if err != nil {
		log.Fatalf("failed to read lockfile %v", err)
	}
	// TODO: Pretty print this properly
	fmt.Printf("lockBin.CoreDependencies: %v\n", lockBin.CoreDependencies)
	fmt.Printf("lockBin.Resolutions: %v\n", lockBin.Resolutions)
}
