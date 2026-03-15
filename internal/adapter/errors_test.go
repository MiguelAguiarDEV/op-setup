package adapter

import (
	"errors"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestAgentNotSupportedError_Is(t *testing.T) {
	err := &AgentNotSupportedError{Agent: "unknown"}
	if !errors.Is(err, ErrAgentNotSupported) {
		t.Fatal("AgentNotSupportedError should match ErrAgentNotSupported")
	}
	if errors.Is(err, ErrDuplicateAdapter) {
		t.Fatal("AgentNotSupportedError should not match ErrDuplicateAdapter")
	}
}

func TestAgentNotSupportedError_Message(t *testing.T) {
	err := &AgentNotSupportedError{Agent: "foo"}
	want := `agent not supported: "foo"`
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestDuplicateAdapterError_Is(t *testing.T) {
	err := &DuplicateAdapterError{Agent: model.AgentClaudeCode}
	if !errors.Is(err, ErrDuplicateAdapter) {
		t.Fatal("DuplicateAdapterError should match ErrDuplicateAdapter")
	}
	if errors.Is(err, ErrAgentNotSupported) {
		t.Fatal("DuplicateAdapterError should not match ErrAgentNotSupported")
	}
}

func TestDuplicateAdapterError_Message(t *testing.T) {
	err := &DuplicateAdapterError{Agent: model.AgentClaudeCode}
	want := `adapter already registered for agent "claude-code"`
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestConfigNotFoundError_Is(t *testing.T) {
	err := &ConfigNotFoundError{Path: "/tmp/missing.json"}
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatal("ConfigNotFoundError should match ErrConfigNotFound")
	}
	if errors.Is(err, ErrConfigCorrupted) {
		t.Fatal("ConfigNotFoundError should not match ErrConfigCorrupted")
	}
}

func TestConfigNotFoundError_Message(t *testing.T) {
	err := &ConfigNotFoundError{Path: "/tmp/missing.json"}
	want := "config file not found: /tmp/missing.json"
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestConfigCorruptedError_Is(t *testing.T) {
	err := &ConfigCorruptedError{Path: "/tmp/bad.json", Reason: "invalid JSON"}
	if !errors.Is(err, ErrConfigCorrupted) {
		t.Fatal("ConfigCorruptedError should match ErrConfigCorrupted")
	}
	if errors.Is(err, ErrConfigNotFound) {
		t.Fatal("ConfigCorruptedError should not match ErrConfigNotFound")
	}
}

func TestConfigCorruptedError_Message(t *testing.T) {
	err := &ConfigCorruptedError{Path: "/tmp/bad.json", Reason: "invalid JSON"}
	want := "config file is not valid: /tmp/bad.json (invalid JSON)"
	if err.Error() != want {
		t.Fatalf("got %q, want %q", err.Error(), want)
	}
}

func TestSentinelErrors_NotEqual(t *testing.T) {
	sentinels := []error{
		ErrAgentNotSupported,
		ErrDuplicateAdapter,
		ErrConfigNotFound,
		ErrConfigCorrupted,
	}
	for i := 0; i < len(sentinels); i++ {
		for j := i + 1; j < len(sentinels); j++ {
			if errors.Is(sentinels[i], sentinels[j]) {
				t.Fatalf("sentinel %v should not match %v", sentinels[i], sentinels[j])
			}
		}
	}
}
