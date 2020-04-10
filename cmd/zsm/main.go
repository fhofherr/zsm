package main

import (
	"os"

	"github.com/fhofherr/zsm/internal/cmd"
)

func main() {
	zsmCmd := cmd.NewZSMCommand()
	if err := zsmCmd.Execute(); err != nil {
		// Cobra takes care of printing the error.
		os.Exit(1)
	}
}
