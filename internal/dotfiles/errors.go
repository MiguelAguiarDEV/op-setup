package dotfiles

import (
	"errors"
	"fmt"
)

// Sentinel errors for dotfiles operations.
var (
	ErrDeployFailed = errors.New("dotfile deployment failed")
	ErrReadEmbed    = errors.New("failed to read embedded file")
)

// DeployFailedError is returned when a file deployment fails.
type DeployFailedError struct {
	Path   string
	Reason string
}

func (e *DeployFailedError) Error() string {
	return fmt.Sprintf("deploy %q failed: %s", e.Path, e.Reason)
}

func (e *DeployFailedError) Is(target error) bool {
	return target == ErrDeployFailed
}

// ReadEmbedError is returned when an embedded file cannot be read.
type ReadEmbedError struct {
	Path   string
	Reason string
}

func (e *ReadEmbedError) Error() string {
	return fmt.Sprintf("read embedded file %q: %s", e.Path, e.Reason)
}

func (e *ReadEmbedError) Is(target error) bool {
	return target == ErrReadEmbed
}
