package main

import (
	"flag"
	"fmt"
)

var (
	VERSION = "__REPLACEMENT__"
	COMMIT  = "__REPLACEMENT__"
)

func argParse() (string, error) {
	version := flag.Bool("version", false, "show version")
	confFile := flag.String("config", "", "")
	flag.Parse()
	if *version {
		return "", fmt.Errorf("version %s (commit: %s)", VERSION, COMMIT)
	}
	if *confFile == "" {
		return "", fmt.Errorf("flag --config must be specified")
	}
	return *confFile, nil
}
