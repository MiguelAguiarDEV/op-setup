package installer

import (
	"context"
	"errors"
	"testing"
)

func TestPlaywrightInstaller_IDAndName(t *testing.T) {
	inst := &PlaywrightInstaller{}
	if got := inst.ID(); got != "playwright" {
		t.Errorf("ID() = %q, want %q", got, "playwright")
	}
	if got := inst.Name(); got != "Playwright (Chromium)" {
		t.Errorf("Name() = %q, want %q", got, "Playwright (Chromium)")
	}
}

func TestPlaywrightInstaller_Prerequisites(t *testing.T) {
	inst := &PlaywrightInstaller{}
	prereqs := inst.Prerequisites()
	if len(prereqs) != 1 || prereqs[0] != "npx" {
		t.Errorf("Prerequisites() = %v, want [npx]", prereqs)
	}
}

func TestPlaywrightInstaller_Detect(t *testing.T) {
	tests := []struct {
		name     string
		globFunc func(string) ([]string, error)
		want     bool
		wantErr  bool
	}{
		{
			name: "chromium cache found",
			globFunc: func(_ string) ([]string, error) {
				return []string{"/home/test/.cache/ms-playwright/chromium-1234"}, nil
			},
			want: true,
		},
		{
			name: "no chromium cache",
			globFunc: func(_ string) ([]string, error) {
				return nil, nil
			},
			want: false,
		},
		{
			name: "glob error propagated",
			globFunc: func(_ string) ([]string, error) {
				return nil, errors.New("bad pattern")
			},
			want:    false,
			wantErr: true,
		},
		{
			name: "empty matches",
			globFunc: func(_ string) ([]string, error) {
				return []string{}, nil
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inst := &PlaywrightInstaller{
				Cmd:      &MockCommandRunner{},
				HomeDir:  "/home/test",
				GlobFunc: tt.globFunc,
			}
			got, err := inst.Detect(context.Background())
			if (err != nil) != tt.wantErr {
				t.Fatalf("Detect() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Detect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlaywrightInstaller_Install(t *testing.T) {
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
			name: "npx fails",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("error"), errors.New("exit 1")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCommandRunner{RunFunc: tt.runFunc}
			inst := &PlaywrightInstaller{Cmd: mock, HomeDir: "/home/test"}
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
			if call.Name != "npx" {
				t.Errorf("command = %q, want %q", call.Name, "npx")
			}
			wantArgs := []string{"playwright", "install", "chromium"}
			for i, arg := range wantArgs {
				if i >= len(call.Args) || call.Args[i] != arg {
					t.Errorf("args[%d] = %q, want %q", i, call.Args[i], arg)
				}
			}
		})
	}
}

func TestPlaywrightInstaller_Rollback(t *testing.T) {
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
			inst := &PlaywrightInstaller{Cmd: mock, HomeDir: "/home/test"}
			err := inst.Rollback(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Rollback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
