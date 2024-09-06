package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Eyepan/yap/src/logger"
)

func HandleArgs() {
	// slog.SetLogLoggerLevel(slog.LevelWarn)
	args := os.Args
	if len(args) < 2 {
		slog.Error("Must input a command")
		HandleHelp()
		os.Exit(-1)
	}
	switch args[1] {
	case "install":
		logger.PrintCurrentCommand(args[1])
		HandleInstall()
		// TODO: add/install packages
	case "list":
		HandleList()
	case "add":
		logger.PrintCurrentCommand(args[1])
		slog.Error("hasn't been implemented yet, sorry")
		os.Exit(-1)
	case "update":
		logger.PrintCurrentCommand(args[1])
		slog.Error("hasn't been implemented yet, sorry")
		os.Exit(-1)
	case "uninstall":
		logger.PrintCurrentCommand(args[1])
		slog.Error("hasn't been implemented yet, sorry")
		os.Exit(-1)
	case "help":
		HandleHelp()
	default:
		slog.Error("Unknown command")
		HandleHelp()
		os.Exit(-1)
	}

}

func HandleHelp() {
	fmt.Print(`YAP: Yet-Another-Package manager
	Usage:
		./yap <command>
	Commands:
		help
			prints this out!
		install
			installs a list of packages
		list
			list out packages that should be installed
		list --all
			list out all packages and their n-th dependencies
		add	<package-name>@<!version> 
			adds this particular package to package.json and install it in the repository
		update <package-name>
			updates the selected package to its latest version
		update --all
			updates all dependencies to its latest version
		uninstall <package-name>
			removes this package from the list of dependencies
`)
}
