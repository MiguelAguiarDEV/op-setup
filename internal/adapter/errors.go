package adapter

import (
	"errors"
	"fmt"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Sentinel errors for adapter operations.
var (
	ErrAgentNotSupported = errors.New("agent not supported")
	ErrDuplicateAdapter  = errors.New("adapter already registered")
	ErrConfigNotFound    = errors.New("config file not found")
	ErrConfigCorrupted   = errors.New("config file is not valid")
)

// AgentNotSupportedError is returned when an unknown AgentID is requested.
type AgentNotSupportedError struct {
	Agent model.AgentID
}

func (e *AgentNotSupportedError) Error() string {
	return fmt.Sprintf("agent not supported: %q", e.Agent)
}

func (e *AgentNotSupportedError) Is(target error) bool {
	return target == ErrAgentNotSupported
}

// DuplicateAdapterError is returned when registering an adapter that already exists.
type DuplicateAdapterError struct {
	Agent model.AgentID
}

func (e *DuplicateAdapterError) Error() string {
	return fmt.Sprintf("adapter already registered for agent %q", e.Agent)
}

func (e *DuplicateAdapterError) Is(target error) bool {
	return target == ErrDuplicateAdapter
}

// ConfigNotFoundError is returned when a config file does not exist.
type ConfigNotFoundError struct {
	Path string
}

func (e *ConfigNotFoundError) Error() string {
	return fmt.Sprintf("config file not found: %s", e.Path)
}

func (e *ConfigNotFoundError) Is(target error) bool {
	return target == ErrConfigNotFound
}

// ConfigCorruptedError is returned when a config file cannot be parsed.
type ConfigCorruptedError struct {
	Path   string
	Reason string
}

func (e *ConfigCorruptedError) Error() string {
	return fmt.Sprintf("config file is not valid: %s (%s)", e.Path, e.Reason)
}

func (e *ConfigCorruptedError) Is(target error) bool {
	return target == ErrConfigCorrupted
}
