package app

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

// stubAdapter is a minimal adapter for headless tests.
type stubAdapter struct {
	agent     model.AgentID
	name      string
	installed bool
}

func (s *stubAdapter) Name() string                                                    { return s.name }
func (s *stubAdapter) Agent() model.AgentID                                            { return s.agent }
func (s *stubAdapter) ConfigPath(homeDir string) string                                { return homeDir + "/.stub-config.json" }
func (s *stubAdapter) MCPStrategy() model.MCPStrategy                                  { return model.StrategyMergeIntoJSON }
func (s *stubAdapter) MCPConfigKey() string                                            { return "mcpServers" }
func (s *stubAdapter) PostInject(homeDir string, components []model.ComponentID) error { return nil }
func (s *stubAdapter) Detect(homeDir string) (model.DetectResult, error) {
	return model.DetectResult{Installed: s.installed, ConfigFound: s.installed}, nil
}

func newTestRegistry(adapters ...adapter.Adapter) *adapter.Registry {
	r, err := adapter.NewRegistry(adapters...)
	if err != nil {
		panic(err)
	}
	return r
}

func TestRunHeadless_DryRun_DotfilesOnly(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	var buf bytes.Buffer

	err := runHeadless(RunConfig{
		DryRun:  true,
		Profile: model.ProfileDotfilesOnly,
	}, headlessDeps{
		homeDir:         t.TempDir(),
		adapterRegistry: newTestRegistry(&stubAdapter{agent: "test", name: "Test", installed: true}),
		stdout:          &buf,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[DRY RUN]") {
		t.Fatalf("expected dry-run header in output, got: %s", output)
	}
	if !strings.Contains(output, "deploy-dotfiles") {
		t.Fatalf("expected deploy-dotfiles step in output, got: %s", output)
	}
}

func TestRunHeadless_DotfilesOnly_SkipsDetection(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	var buf bytes.Buffer

	// No adapters are "installed" — should still succeed for dotfiles-only.
	err := runHeadless(RunConfig{
		DryRun:  true,
		Profile: model.ProfileDotfilesOnly,
	}, headlessDeps{
		homeDir:         t.TempDir(),
		adapterRegistry: newTestRegistry(&stubAdapter{agent: "test", name: "Test", installed: false}),
		stdout:          &buf,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if strings.Contains(output, "detected:") {
		t.Fatalf("dotfiles-only should not detect agents, got: %s", output)
	}
}

func TestRunHeadless_NoAgentsDetected_Error(t *testing.T) {
	var buf bytes.Buffer

	err := runHeadless(RunConfig{
		Profile: model.ProfileMCPOnly,
	}, headlessDeps{
		homeDir:         t.TempDir(),
		adapterRegistry: newTestRegistry(&stubAdapter{agent: "test", name: "Test", installed: false}),
		stdout:          &buf,
	})
	if err == nil {
		t.Fatal("expected error when no agents detected")
	}
	if !strings.Contains(err.Error(), "no AI tools detected") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestRunHeadless_DetectsAgents(t *testing.T) {
	var buf bytes.Buffer

	// Two adapters: one installed, one not. Verify detection output.
	err := runHeadless(RunConfig{
		DryRun:  true,
		Profile: model.ProfileMCPOnly,
	}, headlessDeps{
		homeDir: t.TempDir(),
		adapterRegistry: newTestRegistry(
			&stubAdapter{agent: "test-a", name: "TestA", installed: true},
			&stubAdapter{agent: "test-b", name: "TestB", installed: false},
		),
		stdout: &buf,
	})

	output := buf.String()
	if !strings.Contains(output, "detected: TestA") {
		t.Fatalf("expected TestA detected in output, got: %s", output)
	}
	if strings.Contains(output, "detected: TestB") {
		t.Fatalf("TestB should not be detected, got: %s", output)
	}

	// BuildPlan uses the same registry passed in deps, so the stub adapter
	// "test-a" will fail resolution because it's not a real adapter with
	// a config path that pipeline can use. Error from build plan is expected.
	if err != nil && !strings.Contains(err.Error(), "build plan") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunHeadless_MCPOnly_NoAgentsDetected(t *testing.T) {
	var buf bytes.Buffer

	// All stub adapters report not installed — should get "no AI tools" error.
	err := runHeadless(RunConfig{
		DryRun:  true,
		Profile: model.ProfileMCPOnly,
	}, headlessDeps{
		homeDir: t.TempDir(),
		adapterRegistry: newTestRegistry(
			&stubAdapter{agent: "stub-a", name: "StubA", installed: false},
			&stubAdapter{agent: "stub-b", name: "StubB", installed: false},
		),
		stdout: &buf,
	})

	if err == nil {
		t.Fatal("expected error — no agents detected")
	}
	if !strings.Contains(err.Error(), "no AI tools detected") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFprintProgressEvent(t *testing.T) {
	tests := []struct {
		name   string
		status pipeline.StepStatus
		expect string
	}{
		{"running", pipeline.StatusRunning, "..."},
		{"succeeded", pipeline.StatusSucceeded, " ok"},
		{"failed", pipeline.StatusFailed, "ERR"},
		{"rolledback", pipeline.StatusRolledBack, "<--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			fprintProgressEvent(&buf, pipeline.ProgressEvent{
				Stage:  pipeline.StageApply,
				StepID: "test-step",
				Status: tt.status,
			})
			if !strings.Contains(buf.String(), tt.expect) {
				t.Fatalf("expected %q in output, got: %s", tt.expect, buf.String())
			}
		})
	}
}

func TestFprintProgressEvent_WithError(t *testing.T) {
	var buf bytes.Buffer
	fprintProgressEvent(&buf, pipeline.ProgressEvent{
		Stage:  pipeline.StageApply,
		StepID: "test-step",
		Status: pipeline.StatusFailed,
		Err:    fmt.Errorf("something broke"),
	})
	output := buf.String()
	if !strings.Contains(output, "something broke") {
		t.Fatalf("expected error in output, got: %s", output)
	}
}

func TestFprintResult_Success(t *testing.T) {
	var buf bytes.Buffer
	fprintResult(&buf, pipeline.ExecutionResult{})
	if !strings.Contains(buf.String(), "Setup complete!") {
		t.Fatalf("expected success message, got: %s", buf.String())
	}
}

func TestFprintResult_Failure(t *testing.T) {
	var buf bytes.Buffer
	fprintResult(&buf, pipeline.ExecutionResult{
		Err: fmt.Errorf("install stage failed"),
	})
	output := buf.String()
	if !strings.Contains(output, "Setup failed") {
		t.Fatalf("expected failure message, got: %s", output)
	}
}

func TestFprintResult_WithRollback(t *testing.T) {
	var buf bytes.Buffer
	fprintResult(&buf, pipeline.ExecutionResult{
		Err: fmt.Errorf("deploy failed"),
		Rollback: &pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "step-a", Status: pipeline.StatusRolledBack},
				{StepID: "step-b", Status: pipeline.StatusFailed},
			},
		},
	})
	output := buf.String()
	if !strings.Contains(output, "Rollback was executed") {
		t.Fatalf("expected rollback message, got: %s", output)
	}
	if !strings.Contains(output, "step-a") {
		t.Fatalf("expected step-a in rollback output, got: %s", output)
	}
	if !strings.Contains(output, "FAILED") {
		t.Fatalf("expected FAILED status for step-b, got: %s", output)
	}
}

func TestFprintPlan(t *testing.T) {
	var buf bytes.Buffer

	plan := pipeline.StagePlan{
		Prepare: []pipeline.Step{&fakeStep{id: "validate-deps"}},
		Deploy:  []pipeline.Step{&fakeStep{id: "deploy-dotfiles"}},
	}

	fprintPlan(&buf, plan, model.ProfileDotfilesOnly)
	output := buf.String()

	if !strings.Contains(output, "[DRY RUN]") {
		t.Fatalf("expected DRY RUN header, got: %s", output)
	}
	if !strings.Contains(output, "Dotfiles Only") {
		t.Fatalf("expected profile name, got: %s", output)
	}
	if !strings.Contains(output, "validate-deps") {
		t.Fatalf("expected validate-deps step, got: %s", output)
	}
	if !strings.Contains(output, "deploy-dotfiles") {
		t.Fatalf("expected deploy-dotfiles step, got: %s", output)
	}
	if !strings.Contains(output, "Total: 2 steps") {
		t.Fatalf("expected total 2 steps, got: %s", output)
	}
}

// fakeStep implements pipeline.Step for testing print functions.
type fakeStep struct {
	id string
}

func (f *fakeStep) ID() string { return f.id }
func (f *fakeStep) Run() error { return nil }
