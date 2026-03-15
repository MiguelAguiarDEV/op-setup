package steps

import (
	"fmt"
	"os/exec"
	"strings"
)

// DependencyCheck describes a binary dependency to validate.
type DependencyCheck struct {
	// Binary is the name of the binary to look for in PATH.
	Binary string

	// Required means failure if missing. Optional means warning only.
	Required bool

	// Message is a user-facing description of why this is needed.
	Message string
}

// ValidateStep checks that required binary dependencies exist.
type ValidateStep struct {
	LookPath func(string) (string, error)
	Checks   []DependencyCheck
	Warnings []string // populated after Run
}

// NewValidateStep creates a ValidateStep with default dependencies.
func NewValidateStep(checks []DependencyCheck) *ValidateStep {
	return &ValidateStep{
		LookPath: exec.LookPath,
		Checks:   checks,
	}
}

func (s *ValidateStep) ID() string { return "validate-deps" }

// Run checks all dependencies. Returns error if any required dep is missing.
// Warnings for optional deps are collected in s.Warnings.
func (s *ValidateStep) Run() error {
	s.Warnings = nil
	var missing []string

	for _, check := range s.Checks {
		_, err := s.LookPath(check.Binary)
		if err != nil {
			if check.Required {
				missing = append(missing, fmt.Sprintf("%s: %s", check.Binary, check.Message))
			} else {
				s.Warnings = append(s.Warnings, fmt.Sprintf("%s: %s", check.Binary, check.Message))
			}
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required dependencies:\n  %s", strings.Join(missing, "\n  "))
	}

	return nil
}
