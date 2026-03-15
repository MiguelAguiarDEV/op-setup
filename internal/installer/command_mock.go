package installer

import "context"

// MockCommandRunner records calls and returns preconfigured results.
type MockCommandRunner struct {
	// RunFunc is called for each Run invocation.
	RunFunc func(ctx context.Context, name string, args ...string) ([]byte, error)

	// LookPathFunc is called for each LookPath invocation.
	LookPathFunc func(name string) (string, error)

	// RunCalls records all Run invocations.
	RunCalls []CommandCall

	// LookPathCalls records all LookPath invocations.
	LookPathCalls []string
}

// CommandCall records a single Run invocation.
type CommandCall struct {
	Name string
	Args []string
}

// Run records the call and delegates to RunFunc if set.
func (m *MockCommandRunner) Run(ctx context.Context, name string, args ...string) ([]byte, error) {
	m.RunCalls = append(m.RunCalls, CommandCall{Name: name, Args: args})
	if m.RunFunc != nil {
		return m.RunFunc(ctx, name, args...)
	}
	return nil, nil
}

// LookPath records the call and delegates to LookPathFunc if set.
func (m *MockCommandRunner) LookPath(name string) (string, error) {
	m.LookPathCalls = append(m.LookPathCalls, name)
	if m.LookPathFunc != nil {
		return m.LookPathFunc(name)
	}
	return "/usr/bin/" + name, nil
}
