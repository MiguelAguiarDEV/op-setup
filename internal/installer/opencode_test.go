package installer

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestOpenCodeInstaller_IDAndName(t *testing.T) {
	inst := &OpenCodeInstaller{}
	if got := inst.ID(); got != "opencode" {
		t.Errorf("ID() = %q, want %q", got, "opencode")
	}
	if got := inst.Name(); got != "OpenCode" {
		t.Errorf("Name() = %q, want %q", got, "OpenCode")
	}
}

func TestOpenCodeInstaller_Prerequisites(t *testing.T) {
	inst := &OpenCodeInstaller{}
	prereqs := inst.Prerequisites()
	if len(prereqs) != 1 || prereqs[0] != "npm" {
		t.Errorf("Prerequisites() = %v, want [npm]", prereqs)
	}
}

func TestOpenCodeInstaller_Detect(t *testing.T) {
	tests := []struct {
		name     string
		lookPath func(string) (string, error)
		statPath func(string) (os.FileInfo, error)
		want     bool
	}{
		{
			name: "found in PATH",
			lookPath: func(_ string) (string, error) {
				return "/usr/bin/opencode", nil
			},
			statPath: nil,
			want:     true,
		},
		{
			name: "found in home dir",
			lookPath: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			statPath: func(_ string) (os.FileInfo, error) {
				return nil, nil // exists
			},
			want: true,
		},
		{
			name: "not found anywhere",
			lookPath: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			statPath: func(_ string) (os.FileInfo, error) {
				return nil, os.ErrNotExist
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &OpenCodeInstaller{
				Cmd:      &MockCommandRunner{LookPathFunc: tt.lookPath},
				HomeDir:  "/home/test",
				StatPath: tt.statPath,
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

func TestOpenCodeInstaller_Install(t *testing.T) {
	tests := []struct {
		name    string
		runFunc func(context.Context, string, ...string) ([]byte, error)
		wantErr bool
	}{
		{
			name: "success",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("installed"), nil
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
			inst := &OpenCodeInstaller{Cmd: mock, HomeDir: "/home/test"}
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
			wantArgs := []string{"install", "-g", "opencode-ai"}
			for i, arg := range wantArgs {
				if i >= len(call.Args) || call.Args[i] != arg {
					t.Errorf("args[%d] = %q, want %q", i, call.Args[i], arg)
				}
			}
		})
	}
}

func TestOpenCodeInstaller_Rollback(t *testing.T) {
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
			inst := &OpenCodeInstaller{Cmd: mock, HomeDir: "/home/test"}
			err := inst.Rollback(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
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
