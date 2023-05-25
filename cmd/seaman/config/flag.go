package config

import (
	"flag"
	"fmt"

	"github.com/cloudnativedaysjp/seaman/internal/version"
)

func ParseFlag() (*Config, error) {
	var (
		showVersion bool
		confFile    string
	)

	flag.BoolVar(&showVersion, "version", false,
		"show version")
	flag.StringVar(&confFile, "config", "",
		"filename of config (for example, refer to `example.yaml` on this repository)")
	flag.Parse()

	if showVersion {
		return nil, fmt.Errorf("%v", version.Information())
	}
	if confFile == "" {
		return nil, fmt.Errorf("flag --config must be specified")
	}

	conf, err := LoadConf(confFile)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
