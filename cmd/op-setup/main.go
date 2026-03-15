package main

import (
	"fmt"
	"os"

	"github.com/MiguelAguiarDEV/op-setup/internal/app"
)

var version = "dev"

func main() {
	app.Version = version
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
