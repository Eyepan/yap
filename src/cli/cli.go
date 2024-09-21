package cli

import (
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/logger"
)

func HandleArgs() {
	conf, err := config.ReadYapConfig()
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	switch conf.LogLevel {
	case "debug":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "warn":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "error":
		slog.SetLogLoggerLevel(slog.LevelError)
	case "info":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "":
		slog.SetLogLoggerLevel(slog.LevelError)
	default:
		log.Fatalf("failed to read config file, logLevel is not 'debug', 'warn', 'info', 'error' or ''")
	}
	args := os.Args
	if len(args) < 2 {
		slog.Error("Must input a command")
		HandleHelp()
		os.Exit(-1)
	}
	switch args[1] {
	case "install":
		logger.PrintCurrentCommand(args[1])
		if len(args) > 2 && (args[2] == "-f" || args[2] == "--force") {
			HandleInstall(true)
		} else {
			HandleInstall(false)
		}
	case "list":
		HandleList()
	case "add":
		logger.PrintCurrentCommand(args[1])
		HandleAdd()
	case "config":
		HandleConfig()
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
			list out packages from lockfile
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
