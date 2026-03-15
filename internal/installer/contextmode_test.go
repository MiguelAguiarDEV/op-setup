package installer

import (
	"context"
	"errors"
	"testing"
)

func TestContextModeInstaller_IDAndName(t *testing.T) {
	inst := &ContextModeInstaller{}
	if got := inst.ID(); got != "context-mode" {
		t.Errorf("ID() = %q, want %q", got, "context-mode")
	}
	if got := inst.Name(); got != "Context Mode" {
		t.Errorf("Name() = %q, want %q", got, "Context Mode")
	}
}

func TestContextModeInstaller_Prerequisites(t *testing.T) {
	inst := &ContextModeInstaller{}
	prereqs := inst.Prerequisites()
	if len(prereqs) != 1 || prereqs[0] != "npm" {
		t.Errorf("Prerequisites() = %v, want [npm]", prereqs)
	}
}

func TestContextModeInstaller_Detect(t *testing.T) {
	tests := []struct {
		name     string
		lookPath func(string) (string, error)
		runFunc  func(context.Context, string, ...string) ([]byte, error)
		want     bool
	}{
		{
			name: "found in PATH",
			lookPath: func(_ string) (string, error) {
				return "/usr/bin/context-mode", nil
			},
			want: true,
		},
		{
			name: "found via npm list",
			lookPath: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("/usr/lib/node_modules\n└── context-mode@1.0.0"), nil
			},
			want: true,
		},
		{
			name: "not found anywhere",
			lookPath: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return nil, errors.New("exit 1")
			},
			want: false,
		},
		{
			name: "npm list returns empty",
			lookPath: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte(""), nil
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &ContextModeInstaller{
				Cmd: &MockCommandRunner{
					LookPathFunc: tt.lookPath,
					RunFunc:      tt.runFunc,
				},
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

func TestContextModeInstaller_Install(t *testing.T) {
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
			name: "npm fails",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("npm ERR!"), errors.New("exit 1")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{RunFunc: tt.runFunc}
			inst := &ContextModeInstaller{Cmd: mock}
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
			if call.Name != "npm" {
				t.Errorf("command = %q, want %q", call.Name, "npm")
			}
		})
	}
}

func TestContextModeInstaller_Rollback(t *testing.T) {
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
			name: "uninstall fails",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("err"), errors.New("fail")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{RunFunc: tt.runFunc}
			inst := &ContextModeInstaller{Cmd: mock}
			err := inst.Rollback(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
