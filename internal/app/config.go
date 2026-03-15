package app

import (
	"fmt"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// RunConfig holds CLI configuration parsed from flags.
type RunConfig struct {
	// DryRun shows what would happen without executing side effects.
	DryRun bool

	// NonInteractive skips the TUI and runs headless.
	NonInteractive bool

	// Profile pre-selects the setup profile.
	// Empty means not set (TUI will show profile selection screen).
	Profile model.SetupProfile
}

// BuildConfig validates flag values and returns a RunConfig.
// Returns an error for invalid flag combinations.
func BuildConfig(dryRun bool, profileStr string, noInteractive bool) (RunConfig, error) {
	cfg := RunConfig{
		DryRun:         dryRun,
		NonInteractive: noInteractive,
	}

	// Parse profile if provided.
	if profileStr != "" {
		p, err := model.ParseProfile(profileStr)
		if err != nil {
			return cfg, err
		}
		cfg.Profile = p
	}

	// Validate flag combinations.
	if noInteractive && cfg.Profile == "" {
		return cfg, fmt.Errorf("--no-interactive requires --profile (e.g. --profile full)")
	}

	return cfg, nil
}
