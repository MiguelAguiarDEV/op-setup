package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/MiguelAguiarDEV/op-setup/internal/app"
)

var version = "dev"

func main() {
	app.Version = version

	showVersion := flag.Bool("version", false, "Print version and exit")
	dryRun := flag.Bool("dry-run", false, "Show what would happen without executing")
	profileStr := flag.String("profile", "", "Setup profile: full, mcp-only, dotfiles-only")
	noInteractive := flag.Bool("no-interactive", false, "Run headless without TUI")
	flag.Parse()

	if *showVersion {
		fmt.Printf("op-setup %s\n", app.Version)
		os.Exit(0)
	}

	cfg, err := app.BuildConfig(*dryRun, *profileStr, *noInteractive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		flag.Usage()
		os.Exit(1)
	}

	if err := app.Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
