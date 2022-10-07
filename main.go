package main

import (
	"fmt"
	"os"

	"github.com/cloudnativedaysjp/seaman/config"
	"github.com/cloudnativedaysjp/seaman/seaman"
)

func main() {
	conf, err := config.ParseFlag()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := seaman.Run(conf); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
