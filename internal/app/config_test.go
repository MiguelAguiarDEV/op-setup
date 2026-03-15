package app

import (
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestBuildConfig_DefaultsToTUI(t *testing.T) {
	cfg, err := BuildConfig(false, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DryRun {
		t.Fatal("DryRun should be false")
	}
	if cfg.NonInteractive {
		t.Fatal("NonInteractive should be false")
	}
	if cfg.Profile != "" {
		t.Fatalf("Profile should be empty, got %q", cfg.Profile)
	}
}

func TestBuildConfig_DryRunAlone(t *testing.T) {
	cfg, err := BuildConfig(true, "", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.DryRun {
		t.Fatal("DryRun should be true")
	}
	if cfg.NonInteractive {
		t.Fatal("NonInteractive should be false")
	}
}

func TestBuildConfig_ProfileAlone(t *testing.T) {
	cfg, err := BuildConfig(false, "mcp-only", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Profile != model.ProfileMCPOnly {
		t.Fatalf("Profile = %q, want %q", cfg.Profile, model.ProfileMCPOnly)
	}
	if cfg.NonInteractive {
		t.Fatal("NonInteractive should be false")
	}
}

func TestBuildConfig_NonInteractiveWithProfile(t *testing.T) {
	cfg, err := BuildConfig(false, "full", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.NonInteractive {
		t.Fatal("NonInteractive should be true")
	}
	if cfg.Profile != model.ProfileFull {
		t.Fatalf("Profile = %q, want %q", cfg.Profile, model.ProfileFull)
	}
}

func TestBuildConfig_NonInteractiveWithDryRun(t *testing.T) {
	cfg, err := BuildConfig(true, "dotfiles-only", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.DryRun {
		t.Fatal("DryRun should be true")
	}
	if !cfg.NonInteractive {
		t.Fatal("NonInteractive should be true")
	}
	if cfg.Profile != model.ProfileDotfilesOnly {
		t.Fatalf("Profile = %q, want %q", cfg.Profile, model.ProfileDotfilesOnly)
	}
}

func TestBuildConfig_NonInteractiveWithoutProfile_Error(t *testing.T) {
	_, err := BuildConfig(false, "", true)
	if err == nil {
		t.Fatal("expected error for --no-interactive without --profile")
	}
}

func TestBuildConfig_InvalidProfile_Error(t *testing.T) {
	_, err := BuildConfig(false, "invalid", false)
	if err == nil {
		t.Fatal("expected error for invalid profile")
	}
}

func TestBuildConfig_AllValidProfiles(t *testing.T) {
	tests := []struct {
		input string
		want  model.SetupProfile
	}{
		{"full", model.ProfileFull},
		{"mcp-only", model.ProfileMCPOnly},
		{"dotfiles-only", model.ProfileDotfilesOnly},
	}
	for _, tt := range tests {
		cfg, err := BuildConfig(false, tt.input, true)
		if err != nil {
			t.Fatalf("BuildConfig(%q) error: %v", tt.input, err)
		}
		if cfg.Profile != tt.want {
			t.Fatalf("Profile = %q, want %q", cfg.Profile, tt.want)
		}
	}
}
