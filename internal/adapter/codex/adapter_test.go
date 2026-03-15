package codex

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestAdapter_Identity(t *testing.T) {
	a := NewAdapter()
	if a.Name() != "Codex" {
		t.Fatalf("Name() = %q, want %q", a.Name(), "Codex")
	}
	if a.Agent() != model.AgentCodex {
		t.Fatalf("Agent() = %q, want %q", a.Agent(), model.AgentCodex)
	}
	if a.MCPStrategy() != model.StrategyMergeIntoTOML {
		t.Fatalf("MCPStrategy() = %d, want %d", a.MCPStrategy(), model.StrategyMergeIntoTOML)
	}
	if a.MCPConfigKey() != "mcp_servers" {
		t.Fatalf("MCPConfigKey() = %q, want %q", a.MCPConfigKey(), "mcp_servers")
	}
}

func TestAdapter_ConfigPath(t *testing.T) {
	a := NewAdapter()
	got := a.ConfigPath("/home/test")
	want := filepath.Join("/home/test", ".codex", "config.toml")
	if got != want {
		t.Fatalf("ConfigPath() = %q, want %q", got, want)
	}
}

func TestAdapter_Detect(t *testing.T) {
	tests := []struct {
		name        string
		lookPath    func(string) (string, error)
		statPath    func(string) (os.FileInfo, error)
		wantInstall bool
		wantConfig  bool
		wantErr     bool
	}{
		{
			name:        "binary found and config exists",
			lookPath:    func(string) (string, error) { return "/usr/bin/codex", nil },
			statPath:    func(string) (os.FileInfo, error) { return nil, nil },
			wantInstall: true,
			wantConfig:  true,
		},
		{
			name:        "binary not found and config missing",
			lookPath:    func(string) (string, error) { return "", errors.New("not found") },
			statPath:    func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
			wantInstall: false,
			wantConfig:  false,
		},
		{
			name:     "stat error bubbles up",
			lookPath: func(string) (string, error) { return "/usr/bin/codex", nil },
			statPath: func(string) (os.FileInfo, error) {
				return nil, errors.New("permission denied")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Adapter{
				LookPath: tt.lookPath,
				StatPath: tt.statPath,
			}
			result, err := a.Detect("/home/test")
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Installed != tt.wantInstall {
				t.Fatalf("Installed = %v, want %v", result.Installed, tt.wantInstall)
			}
			if result.ConfigFound != tt.wantConfig {
				t.Fatalf("ConfigFound = %v, want %v", result.ConfigFound, tt.wantConfig)
			}
		})
	}
}

func TestAdapter_Detect_LookPathCalledWithCodex(t *testing.T) {
	var calledWith string
	a := &Adapter{
		LookPath: func(name string) (string, error) {
			calledWith = name
			return "", errors.New("not found")
		},
		StatPath: func(string) (os.FileInfo, error) { return nil, os.ErrNotExist },
	}
	_, _ = a.Detect("/home/test")
	if calledWith != "codex" {
		t.Fatalf("LookPath called with %q, want %q", calledWith, "codex")
	}
}

func TestAdapter_PostInject_NoOp(t *testing.T) {
	a := NewAdapter()
	err := a.PostInject("/home/test", []model.ComponentID{model.ComponentEngram})
	if err != nil {
		t.Fatalf("PostInject should be no-op, got error: %v", err)
	}
}
