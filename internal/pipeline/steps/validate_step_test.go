package steps

import (
	"errors"
	"testing"
)

func TestValidateStep_AllFound(t *testing.T) {
	step := &ValidateStep{
		LookPath: func(string) (string, error) { return "/usr/bin/test", nil },
		Checks: []DependencyCheck{
			{Binary: "engram", Required: true, Message: "needed for memory"},
			{Binary: "npx", Required: true, Message: "needed for playwright"},
		},
	}

	if err := step.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(step.Warnings) != 0 {
		t.Fatalf("expected 0 warnings, got %d", len(step.Warnings))
	}
}

func TestValidateStep_RequiredMissing(t *testing.T) {
	step := &ValidateStep{
		LookPath: func(name string) (string, error) {
			if name == "engram" {
				return "", errors.New("not found")
			}
			return "/usr/bin/" + name, nil
		},
		Checks: []DependencyCheck{
			{Binary: "engram", Required: true, Message: "needed for memory"},
			{Binary: "npx", Required: true, Message: "needed for playwright"},
		},
	}

	err := step.Run()
	if err == nil {
		t.Fatal("expected error for missing required dep")
	}
}

func TestValidateStep_OptionalMissing_Warning(t *testing.T) {
	step := &ValidateStep{
		LookPath: func(name string) (string, error) {
			if name == "optional-tool" {
				return "", errors.New("not found")
			}
			return "/usr/bin/" + name, nil
		},
		Checks: []DependencyCheck{
			{Binary: "engram", Required: true, Message: "needed"},
			{Binary: "optional-tool", Required: false, Message: "nice to have"},
		},
	}

	if err := step.Run(); err != nil {
		t.Fatalf("optional missing should not error: %v", err)
	}
	if len(step.Warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d", len(step.Warnings))
	}
}

func TestValidateStep_EmptyChecks(t *testing.T) {
	step := NewValidateStep(nil)
	if err := step.Run(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateStep_ID(t *testing.T) {
	step := NewValidateStep(nil)
	if step.ID() != "validate-deps" {
		t.Fatalf("ID() = %q, want %q", step.ID(), "validate-deps")
	}
}
