package config

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Eyepan/yap/src/types"
	"github.com/Eyepan/yap/src/utils"
)

func ReadYapConfig() (*types.YapConfig, error) {
	configFile, err := utils.GetGlobalConfigDir()
	var config types.YapConfig
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(configFile); err != nil {
		// config file doesn't exist, create one
		config = types.YapConfig{Registry: "https://registry.npmjs.org", LogLevel: "warn"}
		var buf bytes.Buffer
		if err := utils.WriteConfig(&buf, &config); err != nil {
			return nil, fmt.Errorf("failed to write config to buffer: %w", err)
		}
		file, err := os.Create(configFile)
		if err != nil {
			return nil, fmt.Errorf("failed to create config file in %s: %w", configFile, err)
		}
		defer file.Close()
		if _, err := file.Write(buf.Bytes()); err != nil {
			return nil, fmt.Errorf("failed to write to config file in %s: %w", configFile, err)
		}
		return &config, nil
	}
	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file in %s: %w", configFile, err)
	}
	buf := bytes.NewReader(data)
	return utils.ReadConfig(buf)
}
