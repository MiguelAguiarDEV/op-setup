package installer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEngramInstaller_IDAndName(t *testing.T) {
	inst := &EngramInstaller{}
	if got := inst.ID(); got != "engram" {
		t.Errorf("ID() = %q, want %q", got, "engram")
	}
	if got := inst.Name(); got != "Engram" {
		t.Errorf("Name() = %q, want %q", got, "Engram")
	}
}

func TestEngramInstaller_Prerequisites(t *testing.T) {
	inst := &EngramInstaller{}
	prereqs := inst.Prerequisites()
	if len(prereqs) != 1 || prereqs[0] != "go" {
		t.Errorf("Prerequisites() = %v, want [go]", prereqs)
	}
}

func TestEngramInstaller_Detect(t *testing.T) {
	tests := []struct {
		name     string
		lookPath func(string) (string, error)
		want     bool
	}{
		{
			name: "found in PATH",
			lookPath: func(_ string) (string, error) {
				return "/usr/local/bin/engram", nil
			},
			want: true,
		},
		{
			name: "not found",
			lookPath: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &EngramInstaller{
				Cmd: &MockCommandRunner{LookPathFunc: tt.lookPath},
			}
			got, err := inst.Detect(context.Background())
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEngramInstaller_Install(t *testing.T) {
	tests := []struct {
		name    string
		runFunc func(context.Context, string, ...string) ([]byte, error)
		wantErr bool
	}{
		{
			name: "success",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return nil, nil
			},
			wantErr: false,
		},
		{
			name: "go install fails",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("module not found"), errors.New("exit 1")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{RunFunc: tt.runFunc}
			inst := &EngramInstaller{Cmd: mock, HomeDir: "/home/test"}
			err := inst.Install(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Install() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrInstallFailed) {
				t.Errorf("Install() error type = %T, want InstallFailedError", err)
			}
			if len(mock.RunCalls) != 1 {
				t.Fatalf("RunCalls = %d, want 1", len(mock.RunCalls))
			}
			call := mock.RunCalls[0]
			if call.Name != "go" {
				t.Errorf("command = %q, want %q", call.Name, "go")
			}
			if len(call.Args) < 2 || call.Args[0] != "install" {
				t.Errorf("args = %v, want [install ...]", call.Args)
			}
		})
	}
}

func TestEngramInstaller_Rollback(t *testing.T) {
	tests := []struct {
		name      string
		setupDir  bool
		removeErr error
		wantErr   bool
	}{
		{
			name:     "binary exists and removed",
			setupDir: true,
			wantErr:  false,
		},
		{
			name:      "binary does not exist",
			setupDir:  false,
			removeErr: os.ErrNotExist,
			wantErr:   false,
		},
		{
			name:      "remove fails with other error",
			setupDir:  false,
			removeErr: errors.New("permission denied"),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.setupDir {
				binDir := filepath.Join(tmpDir, "go", "bin")
				if err := os.MkdirAll(binDir, 0o755); err != nil {
					t.Fatal(err)
				}
				binPath := filepath.Join(binDir, "engram")
				if err := os.WriteFile(binPath, []byte("fake"), 0o755); err != nil {
					t.Fatal(err)
				}
			}

			var removeFunc func(string) error
			if tt.removeErr != nil {
				removeFunc = func(_ string) error { return tt.removeErr }
			} else {
				removeFunc = os.Remove
			}

			// Clear GOPATH so it falls back to HomeDir/go.
			t.Setenv("GOPATH", "")

			inst := &EngramInstaller{
				Cmd:        &MockCommandRunner{},
				HomeDir:    tmpDir,
				RemoveFunc: removeFunc,
			}
			err := inst.Rollback(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngramInstaller_Rollback_GOPATH(t *testing.T) {
	tmpDir := t.TempDir()
	gopathDir := filepath.Join(tmpDir, "custom-gopath")
	binDir := filepath.Join(gopathDir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	binPath := filepath.Join(binDir, "engram")
	if err := os.WriteFile(binPath, []byte("fake"), 0o755); err != nil {
		t.Fatal(err)
	}

	t.Setenv("GOPATH", gopathDir)

	inst := &EngramInstaller{
		Cmd:        &MockCommandRunner{},
		HomeDir:    "/should-not-be-used",
		RemoveFunc: os.Remove,
	}
	if err := inst.Rollback(context.Background()); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	if _, err := os.Stat(binPath); !os.IsNotExist(err) {
		t.Error("binary should have been removed")
	}
}
