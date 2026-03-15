package dotfiles

import (
	"errors"
	"strings"
	"testing"
)

func TestDeployFailedError(t *testing.T) {
	err := &DeployFailedError{Path: "/home/test/.config/opencode/AGENTS.md", Reason: "permission denied"}
	want := `deploy "/home/test/.config/opencode/AGENTS.md" failed: permission denied`
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
	if !errors.Is(err, ErrDeployFailed) {
		t.Error("Is(ErrDeployFailed) = false, want true")
	}
	if errors.Is(err, ErrReadEmbed) {
		t.Error("Is(ErrReadEmbed) = true, want false")
	}
}

func TestReadEmbedError(t *testing.T) {
	err := &ReadEmbedError{Path: "embed/opencode/AGENTS.md", Reason: "file not found"}
	if !strings.Contains(err.Error(), "embed/opencode/AGENTS.md") {
		t.Errorf("Error() = %q, want to contain path", err.Error())
	}
	if !errors.Is(err, ErrReadEmbed) {
		t.Error("Is(ErrReadEmbed) = false, want true")
	}
	if errors.Is(err, ErrDeployFailed) {
		t.Error("Is(ErrDeployFailed) = true, want false")
	}
}
