package main

import (
	"fmt"
	"os"

	"github.com/mctofu/homekit/cmd/homekit/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}
