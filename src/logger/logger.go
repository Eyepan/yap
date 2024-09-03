package logger

import (
	"fmt"
)

func PrettyPrintStats(resolvCount int, totalResolvCount int, downloadCount int, totalDownloadCount int) {
	fmt.Printf("\r🔍[%d/%d] 📦[%d/%d]", resolvCount, totalResolvCount, downloadCount, totalDownloadCount)
}
