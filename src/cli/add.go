package cli

import (
	"log/slog"
	"os"
)

func HandleAdd() {
	args := os.Args
	if len(args) <= 2 {
		slog.Error("missing packages to install")
		os.Exit(-1)
	}
	// TODO: install packages by getting them from args

}
