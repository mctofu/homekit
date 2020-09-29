package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mctofu/homekit/gen"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	services := flag.Bool("services", false, "generate services")
	characteristics := flag.Bool("characteristics", false, "generate characteristics")

	flag.Parse()

	if *characteristics {
		if err := gen.GenerateCharacteristics(); err != nil {
			return fmt.Errorf("GenerateCharacteristics: %v", err)
		}
	}

	if *services {
		if err := gen.GenerateServices(); err != nil {
			return fmt.Errorf("GenerateServices: %v", err)
		}
	}

	return nil
}
