package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cloudnativedaysjp/seaman/seaman"
)

func main() {
	var confFile string
	flag.StringVar(&confFile, "config", "", "filename of config (for example, refer to `example.yaml` on this repository)")
	flag.Parse()
	if confFile == "" {
		fmt.Println("flag --config must be specified")
		os.Exit(1)
	}

	conf, err := seaman.LoadConf(confFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := seaman.Run(conf); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
