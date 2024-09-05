package logger

import (
	"fmt"
	"log/slog"
)

func PrettyPrintStats(resolveCount int, totalResolveCount int, downloadCount int, totalDownloadCount int) {
	// TODO: figure out a way to implement proper printing without inputting all the four numbers all the time
	fmt.Printf("\r🔍[%d/%d] 📦[%d/%d]", resolveCount, totalResolveCount, downloadCount, totalDownloadCount)
}

func InfoLogger(message string) {
	slog.Info(fmt.Sprintf("%s\n", message))
}
