package main

import (
	"fmt"
	"os"
	"path"

	"github.com/pelletier/go-toml"
)

const (
	configOverride = "XDG_CONFIG_HOME"
	configFilePath = "nvidia-container-runtime/config.toml"
)

var (
	configDir = "/etc/"
)

var logger = NewLogger()

func main() {
	err := run(os.Args)
	if err != nil {
		logger.Errorf("Error running %v: %v", os.Args, err)
		os.Exit(1)
	}
}

// run is an entry point that allows for idiomatic handling of errors
// when calling from the main function.
func run(argv []string) (rerr error) {
	cfg, err := getConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	err = logger.LogToFile(cfg.debugFilePath)
	if err != nil {
		return fmt.Errorf("error opening debug log file: %v", err)
	}
	defer func() {
		// We capture and log a returning error before closing the log file.
		if rerr != nil {
			logger.Errorf("Error running %v: %v", argv, rerr)
		}
		logger.CloseFile()
	}()

	runtime, err := newNVIDIAContainerRuntime(logger.Logger, cfg, argv)
	if err != nil {
		return fmt.Errorf("error creating runtime: %v", err)
	}

	return runtime.Exec(argv)
}

type config struct {
	debugFilePath string
	Experimental  bool
}

// getConfig sets up the config struct. Values are read from a toml file
// or set via the environment.
func getConfig() (*config, error) {
	cfg := &config{}

	if XDGConfigDir := os.Getenv(configOverride); len(XDGConfigDir) != 0 {
		configDir = XDGConfigDir
	}

	configFilePath := path.Join(configDir, configFilePath)

	tomlContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	toml, err := toml.Load(string(tomlContent))
	if err != nil {
		return nil, err
	}

	cfg.debugFilePath = toml.GetDefault("nvidia-container-runtime.debug", "/dev/null").(string)
	cfg.Experimental = toml.GetDefault("nvidia-container-runtime.experimental", false).(bool)

	return cfg, nil
}
