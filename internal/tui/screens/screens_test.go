package screens

import (
	"errors"
	"strings"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

func TestRenderWelcome_ContainsVersion(t *testing.T) {
	view := RenderWelcome("1.2.3")
	if !strings.Contains(view, "1.2.3") {
		t.Fatal("welcome should contain version string")
	}
}

func TestRenderWelcome_ContainsLogo(t *testing.T) {
	view := RenderWelcome("dev")
	if !strings.Contains(view, "op-setup") {
		t.Fatal("welcome should contain op-setup text")
	}
}

func TestRenderDetection_InstalledTool(t *testing.T) {
	items := []DetectionItem{
		{Name: "Claude Code", Agent: model.AgentClaudeCode, Installed: true, ConfigFound: true},
	}
	view := RenderDetection(items)
	if !strings.Contains(view, "Claude Code") {
		t.Fatal("should contain tool name")
	}
	if !strings.Contains(view, "config found") {
		t.Fatal("should indicate config found")
	}
}

func TestRenderDetection_NotInstalledTool(t *testing.T) {
	items := []DetectionItem{
		{Name: "Codex", Agent: model.AgentCodex, Installed: false},
	}
	view := RenderDetection(items)
	if !strings.Contains(view, "Codex") {
		t.Fatal("should contain tool name")
	}
	if !strings.Contains(view, "not installed") {
		t.Fatal("should indicate not installed")
	}
}

func TestRenderDetection_InstalledNoConfig(t *testing.T) {
	items := []DetectionItem{
		{Name: "OpenCode", Agent: model.AgentOpenCode, Installed: true, ConfigFound: false},
	}
	view := RenderDetection(items)
	if !strings.Contains(view, "no config file") {
		t.Fatal("should indicate no config file")
	}
}

func TestRenderDetection_Empty(t *testing.T) {
	view := RenderDetection(nil)
	if view == "" {
		t.Fatal("should render even with no items")
	}
}

func TestRenderAgents_WithItems(t *testing.T) {
	items := []AgentItem{
		{Name: "Claude Code", Detected: true, Selected: true},
		{Name: "Codex", Detected: false, Selected: false},
	}
	view := RenderAgents(items, 0)
	if !strings.Contains(view, "Claude Code") {
		t.Fatal("should contain agent name")
	}
	if !strings.Contains(view, "not installed") {
		t.Fatal("should show disabled message for undetected")
	}
}

func TestRenderAgents_CursorPosition(t *testing.T) {
	items := []AgentItem{
		{Name: "A", Detected: true, Selected: false},
		{Name: "B", Detected: true, Selected: false},
	}
	view0 := RenderAgents(items, 0)
	view1 := RenderAgents(items, 1)
	// Views should differ because cursor is at different positions.
	if view0 == view1 {
		t.Fatal("cursor position should affect rendering")
	}
}

func TestRenderComponents_WithItems(t *testing.T) {
	items := []ComponentItem{
		{Name: "Engram", Desc: "Persistent memory", Selected: true},
		{Name: "Playwright", Desc: "Browser automation", Selected: false},
	}
	view := RenderComponents(items, 0)
	if !strings.Contains(view, "Engram") {
		t.Fatal("should contain component name")
	}
	if !strings.Contains(view, "Persistent memory") {
		t.Fatal("should contain description")
	}
}

func TestRenderReview_ShowsSelections(t *testing.T) {
	agents := []string{"Claude Code", "OpenCode"}
	components := []string{"Engram", "Context7"}
	view := RenderReview(agents, components)
	if !strings.Contains(view, "Claude Code") {
		t.Fatal("should contain agent name")
	}
	if !strings.Contains(view, "Engram") {
		t.Fatal("should contain component name")
	}
	if !strings.Contains(view, "backed up") {
		t.Fatal("should mention backup")
	}
}

func TestRenderInstalling_NoEvents(t *testing.T) {
	view := RenderInstalling(nil, 0)
	if !strings.Contains(view, "Installing") {
		t.Fatal("should contain header")
	}
}

func TestRenderInstalling_WithEvents(t *testing.T) {
	events := []pipeline.ProgressEvent{
		{Stage: pipeline.StageApply, StepID: "inject-claude-engram", Status: pipeline.StatusSucceeded},
		{Stage: pipeline.StageApply, StepID: "inject-claude-context7", Status: pipeline.StatusRunning},
		{Stage: pipeline.StageApply, StepID: "inject-opencode-engram", Status: pipeline.StatusFailed, Err: errors.New("write error")},
		{Stage: pipeline.StageRollback, StepID: "rollback-claude", Status: pipeline.StatusRolledBack},
	}
	view := RenderInstalling(events, 4)
	if !strings.Contains(view, "inject-claude-engram") {
		t.Fatal("should contain step ID")
	}
	if !strings.Contains(view, "write error") {
		t.Fatal("should contain error message")
	}
}

func TestRenderInstalling_WithProgressBar(t *testing.T) {
	events := []pipeline.ProgressEvent{
		{StepID: "step-1", Status: pipeline.StatusSucceeded},
		{StepID: "step-2", Status: pipeline.StatusSucceeded},
	}
	view := RenderInstalling(events, 4)
	if !strings.Contains(view, "2/4") {
		t.Fatal("should show progress count")
	}
}

func TestRenderComplete_Success(t *testing.T) {
	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "inject-claude-engram", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Complete") {
		t.Fatal("should indicate completion")
	}
	if !strings.Contains(view, "inject-claude-engram") {
		t.Fatal("should show step results")
	}
}

func TestRenderComplete_Failure(t *testing.T) {
	result := pipeline.ExecutionResult{
		Err: errors.New("pipeline failed"),
		Apply: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "inject-claude-engram", Status: pipeline.StatusFailed, Err: errors.New("write error")},
			},
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Failed") {
		t.Fatal("should indicate failure")
	}
	if !strings.Contains(view, "pipeline failed") {
		t.Fatal("should show error message")
	}
}

func TestRenderComplete_WithRollback(t *testing.T) {
	result := pipeline.ExecutionResult{
		Err: errors.New("failed"),
		Rollback: &pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "rollback-1", Status: pipeline.StatusRolledBack},
			},
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Rollback") {
		t.Fatal("should mention rollback")
	}
}

func TestRenderComplete_EmptySteps(t *testing.T) {
	result := pipeline.ExecutionResult{}
	view := RenderComplete(result)
	if view == "" {
		t.Fatal("should render even with empty result")
	}
}

func TestRenderHeader_WithSubtitle(t *testing.T) {
	view := RenderHeader("Title", "Subtitle")
	if !strings.Contains(view, "Title") {
		t.Fatal("should contain title")
	}
}

func TestRenderHeader_WithoutSubtitle(t *testing.T) {
	view := RenderHeader("Title", "")
	if !strings.Contains(view, "Title") {
		t.Fatal("should contain title")
	}
}

func TestRenderFooter(t *testing.T) {
	view := RenderFooter("Press q to quit")
	if !strings.Contains(view, "Press q to quit") {
		t.Fatal("should contain help text")
	}
}

func TestRenderMultiSelect_AllStates(t *testing.T) {
	items := []SelectItem{
		{Label: "Selected", Selected: true},
		{Label: "Unselected", Selected: false},
		{Label: "Disabled", Disabled: true, DisabledMsg: "unavailable"},
		{Label: "DisabledNoMsg", Disabled: true},
		{Label: "WithDesc", Description: "a description", Selected: false},
	}
	view := RenderMultiSelect(items, 0)
	if !strings.Contains(view, "Selected") {
		t.Fatal("should contain selected item")
	}
	if !strings.Contains(view, "unavailable") {
		t.Fatal("should contain disabled message")
	}
	if !strings.Contains(view, "a description") {
		t.Fatal("should contain description")
	}
}

func TestProgressBar_Zero(t *testing.T) {
	result := progressBar(0, 0, 40)
	if result != "" {
		t.Fatalf("expected empty for zero total, got %q", result)
	}
}

func TestProgressBar_Partial(t *testing.T) {
	result := progressBar(2, 4, 40)
	if !strings.Contains(result, "2/4") {
		t.Fatalf("should contain 2/4, got %q", result)
	}
}

func TestProgressBar_Full(t *testing.T) {
	result := progressBar(4, 4, 40)
	if !strings.Contains(result, "4/4") {
		t.Fatalf("should contain 4/4, got %q", result)
	}
}

func TestProgressBar_Overflow(t *testing.T) {
	// current > total should not panic.
	result := progressBar(10, 4, 40)
	if !strings.Contains(result, "10/4") {
		t.Fatalf("should handle overflow, got %q", result)
	}
}
