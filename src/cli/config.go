package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Eyepan/yap/src/config"
	"github.com/Eyepan/yap/src/utils"
)

func HandleConfig() {
	// get the subcommand. god this should be better implemented instead of me just fetching os.Args
	args := os.Args
	subCommand := ""
	if len(args) <= 2 {
		subCommand = "list"
	} else {
		subCommand = args[2]
	}
	conf, err := config.ReadYapConfig()
	if err != nil {
		log.Fatalf("Failed to load configurations: %v", err)
	}

	switch subCommand {
	case "list":
		{
			jsonData, err := json.MarshalIndent(conf, "", "\t")
			if err != nil {
				log.Fatalf("failed to parse config file into json %v", err)
			}
			fmt.Println(string(jsonData))
		}
	case "get":
		{
			if len(args) <= 3 {
				log.Fatalf("Must pass in key to get from config file")
			}
			switch args[3] {
			case "registry":
				{
					fmt.Println(conf.Registry)
				}
			case "authToken":
				{
					fmt.Println(conf.AuthToken)
				}
			case "logLevel":
				{
					fmt.Println(conf.LogLevel)
				}
			default:
				{
					log.Fatalf("unknown key in config %s", args[3])
				}
			}
		}
	case "set":
		{
			if len(args) <= 4 {
				log.Fatalf("Must pass in both key and the value to set to config file")
			}
			switch args[3] {
			case "registry":
				{
					conf.Registry = args[4]
				}
			case "authToken":
				{
					conf.AuthToken = args[4]
				}
			case "logLevel":
				{
					conf.LogLevel = args[4]
				}
			default:
				{
					log.Fatalf("unknown key in config %s", args[3])
				}
			}
			var buf bytes.Buffer
			if err := utils.WriteConfig(&buf, conf); err != nil {
				log.Fatalf("failed to write config to buffer: %v", err)
			}
			configFile, err := utils.GetGlobalConfigDir()
			if err != nil {
				log.Fatalf("failed to get config file path: %v", err)
			}
			file, err := os.Create(configFile)
			if err != nil {
				log.Fatalf("failed to open config file: %v", err)
			}

			defer file.Close()
			if _, err := file.Write(buf.Bytes()); err != nil {
				log.Fatalf("failed to write to config file: %v", err)
			}
		}
	}
}
