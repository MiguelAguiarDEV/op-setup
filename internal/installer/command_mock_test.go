package installer

import (
	"context"
	"errors"
	"testing"
)

func TestMockCommandRunner_Run(t *testing.T) {
	tests := []struct {
		name     string
		runFunc  func(context.Context, string, ...string) ([]byte, error)
		wantOut  string
		wantErr  bool
		wantCall CommandCall
	}{
		{
			name:     "nil func returns nil",
			runFunc:  nil,
			wantOut:  "",
			wantErr:  false,
			wantCall: CommandCall{Name: "echo", Args: []string{"hello"}},
		},
		{
			name: "custom func returns output",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("ok"), nil
			},
			wantOut:  "ok",
			wantErr:  false,
			wantCall: CommandCall{Name: "echo", Args: []string{"hello"}},
		},
		{
			name: "custom func returns error",
			runFunc: func(_ context.Context, _ string, _ ...string) ([]byte, error) {
				return []byte("fail"), errors.New("boom")
			},
			wantOut:  "fail",
			wantErr:  true,
			wantCall: CommandCall{Name: "echo", Args: []string{"hello"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockCommandRunner{RunFunc: tt.runFunc}
			out, err := m.Run(context.Background(), "echo", "hello")

			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if string(out) != tt.wantOut {
				t.Errorf("Run() output = %q, want %q", string(out), tt.wantOut)
			}
			if len(m.RunCalls) != 1 {
				t.Fatalf("RunCalls = %d, want 1", len(m.RunCalls))
			}
			if m.RunCalls[0].Name != tt.wantCall.Name {
				t.Errorf("RunCalls[0].Name = %q, want %q", m.RunCalls[0].Name, tt.wantCall.Name)
			}
			if len(m.RunCalls[0].Args) != len(tt.wantCall.Args) {
				t.Errorf("RunCalls[0].Args = %v, want %v", m.RunCalls[0].Args, tt.wantCall.Args)
			}
		})
	}
}

func TestMockCommandRunner_LookPath(t *testing.T) {
	tests := []struct {
		name         string
		lookPathFunc func(string) (string, error)
		binary       string
		wantPath     string
		wantErr      bool
	}{
		{
			name:         "nil func returns default path",
			lookPathFunc: nil,
			binary:       "git",
			wantPath:     "/usr/bin/git",
			wantErr:      false,
		},
		{
			name: "custom func returns path",
			lookPathFunc: func(name string) (string, error) {
				return "/custom/" + name, nil
			},
			binary:   "node",
			wantPath: "/custom/node",
			wantErr:  false,
		},
		{
			name: "custom func returns error",
			lookPathFunc: func(_ string) (string, error) {
				return "", errors.New("not found")
			},
			binary:   "missing",
			wantPath: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MockCommandRunner{LookPathFunc: tt.lookPathFunc}
			path, err := m.LookPath(tt.binary)

			if (err != nil) != tt.wantErr {
				t.Errorf("LookPath() error = %v, wantErr %v", err, tt.wantErr)
			}
			if path != tt.wantPath {
				t.Errorf("LookPath() = %q, want %q", path, tt.wantPath)
			}
			if len(m.LookPathCalls) != 1 || m.LookPathCalls[0] != tt.binary {
				t.Errorf("LookPathCalls = %v, want [%q]", m.LookPathCalls, tt.binary)
			}
		})
	}
}

func TestMockCommandRunner_MultipleCalls(t *testing.T) {
	m := &MockCommandRunner{}
	_, _ = m.Run(context.Background(), "a")
	_, _ = m.Run(context.Background(), "b", "1", "2")
	_, _ = m.LookPath("x")
	_, _ = m.LookPath("y")

	if len(m.RunCalls) != 2 {
		t.Errorf("RunCalls = %d, want 2", len(m.RunCalls))
	}
	if len(m.LookPathCalls) != 2 {
		t.Errorf("LookPathCalls = %d, want 2", len(m.LookPathCalls))
	}
}
