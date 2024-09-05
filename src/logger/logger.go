package logger

import (
	"fmt"
)

func PrettyPrintStats(resolveCount int, totalResolveCount int, downloadCount int, totalDownloadCount int) {
	// TODO: figure out a way to implement this without inputting all the four numbers all the time
	fmt.Printf("\r🔍[%d/%d] 📦[%d/%d]", resolveCount, totalResolveCount, downloadCount, totalDownloadCount)
}
