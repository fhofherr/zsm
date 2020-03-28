package main

import (
	"os"

	"github.com/fhofherr/zsm/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		// Cobra takes care of printing the error.
		os.Exit(1)
	}
}
