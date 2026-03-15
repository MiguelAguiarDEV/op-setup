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

// --- Profile screen tests ---

func TestRenderProfile_ShowsAllProfiles(t *testing.T) {
	items := []ProfileItem{
		{Name: "Full Setup", Description: "Install tools, deploy dotfiles, and configure MCP servers"},
		{Name: "MCP Servers Only", Description: "Only configure MCP servers"},
		{Name: "Dotfiles Only", Description: "Only deploy dotfiles"},
	}
	view := RenderProfile(items, 0)
	if !strings.Contains(view, "Full Setup") {
		t.Fatal("should contain Full Setup")
	}
	if !strings.Contains(view, "MCP Servers Only") {
		t.Fatal("should contain MCP Servers Only")
	}
	if !strings.Contains(view, "Dotfiles Only") {
		t.Fatal("should contain Dotfiles Only")
	}
	if !strings.Contains(view, "Select Setup Profile") {
		t.Fatal("should contain header")
	}
}

func TestRenderProfile_CursorPosition(t *testing.T) {
	items := []ProfileItem{
		{Name: "A", Description: "desc A"},
		{Name: "B", Description: "desc B"},
	}
	view0 := RenderProfile(items, 0)
	view1 := RenderProfile(items, 1)
	if view0 == view1 {
		t.Fatal("cursor position should affect rendering")
	}
}

func TestRenderProfile_Empty(t *testing.T) {
	view := RenderProfile(nil, 0)
	if view == "" {
		t.Fatal("should render even with no items")
	}
}

// --- SingleSelect tests ---

func TestRenderSingleSelect_AllStates(t *testing.T) {
	items := []SingleSelectItem{
		{Label: "Selected", Description: "first item"},
		{Label: "Unselected", Description: "second item"},
		{Label: "NoDesc"},
	}
	view := RenderSingleSelect(items, 0)
	if !strings.Contains(view, "Selected") {
		t.Fatal("should contain selected item")
	}
	if !strings.Contains(view, "first item") {
		t.Fatal("should contain description")
	}
	if !strings.Contains(view, "NoDesc") {
		t.Fatal("should contain item without description")
	}
}

func TestRenderSingleSelect_CursorHighlight(t *testing.T) {
	items := []SingleSelectItem{
		{Label: "A"},
		{Label: "B"},
	}
	view0 := RenderSingleSelect(items, 0)
	view1 := RenderSingleSelect(items, 1)
	if view0 == view1 {
		t.Fatal("cursor should affect rendering")
	}
}

func TestRenderSingleSelect_Empty(t *testing.T) {
	view := RenderSingleSelect(nil, 0)
	if view != "" {
		t.Fatalf("expected empty for nil items, got %q", view)
	}
}

// --- Review screen tests ---

func TestRenderReview_ShowsSelections(t *testing.T) {
	agents := []string{"Claude Code", "OpenCode"}
	components := []string{"Engram", "Context7"}
	view := RenderReview(model.ProfileMCPOnly, agents, components)
	if !strings.Contains(view, "Claude Code") {
		t.Fatal("should contain agent name")
	}
	if !strings.Contains(view, "Engram") {
		t.Fatal("should contain component name")
	}
	if !strings.Contains(view, "backed up") {
		t.Fatal("should mention backup")
	}
	if !strings.Contains(view, "MCP Servers Only") {
		t.Fatal("should contain profile name")
	}
}

func TestRenderReview_DotfilesOnly(t *testing.T) {
	view := RenderReview(model.ProfileDotfilesOnly, nil, nil)
	if !strings.Contains(view, "Dotfiles Only") {
		t.Fatal("should contain profile name")
	}
	if !strings.Contains(view, "deploy") {
		t.Fatal("should show dotfiles-specific info")
	}
}

func TestRenderReview_Full(t *testing.T) {
	agents := []string{"Claude Code"}
	components := []string{"Engram"}
	view := RenderReview(model.ProfileFull, agents, components)
	if !strings.Contains(view, "Full Setup") {
		t.Fatal("should contain profile name")
	}
	if !strings.Contains(view, "AI Tools") {
		t.Fatal("should contain AI Tools section")
	}
	if !strings.Contains(view, "MCP Servers") {
		t.Fatal("should contain MCP Servers section")
	}
}

// --- Installing screen tests ---

func TestRenderInstalling_NoEvents(t *testing.T) {
	view := RenderInstalling(model.ProfileMCPOnly, nil, 0)
	if !strings.Contains(view, "Installing") {
		t.Fatal("should contain header")
	}
	if !strings.Contains(view, "Configuring MCP servers") {
		t.Fatal("should contain MCP-specific subtitle")
	}
}

func TestRenderInstalling_WithEvents(t *testing.T) {
	events := []pipeline.ProgressEvent{
		{Stage: pipeline.StageApply, StepID: "inject-claude-engram", Status: pipeline.StatusSucceeded},
		{Stage: pipeline.StageApply, StepID: "inject-claude-context7", Status: pipeline.StatusRunning},
		{Stage: pipeline.StageApply, StepID: "inject-opencode-engram", Status: pipeline.StatusFailed, Err: errors.New("write error")},
		{Stage: pipeline.StageRollback, StepID: "rollback-claude", Status: pipeline.StatusRolledBack},
	}
	view := RenderInstalling(model.ProfileMCPOnly, events, 4)
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
	view := RenderInstalling(model.ProfileMCPOnly, events, 4)
	if !strings.Contains(view, "2/4") {
		t.Fatal("should show progress count")
	}
}

func TestRenderInstalling_ProfileSubtitles(t *testing.T) {
	tests := []struct {
		profile  model.SetupProfile
		contains string
	}{
		{model.ProfileFull, "Installing tools"},
		{model.ProfileMCPOnly, "Configuring MCP servers"},
		{model.ProfileDotfilesOnly, "Deploying dotfiles"},
		{model.SetupProfile("unknown"), "Setting up"},
	}
	for _, tt := range tests {
		view := RenderInstalling(tt.profile, nil, 0)
		if !strings.Contains(view, tt.contains) {
			t.Fatalf("profile %q: expected %q in view", tt.profile, tt.contains)
		}
	}
}

// --- Complete screen tests ---

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
				{StepID: "rollback-2", Status: pipeline.StatusFailed, Err: errors.New("restore error")},
			},
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Rollback") {
		t.Fatal("should mention rollback")
	}
	if !strings.Contains(view, "rollback-1") {
		t.Fatal("should show individual rollback step IDs")
	}
	if !strings.Contains(view, "rollback-2") {
		t.Fatal("should show failed rollback step")
	}
	if !strings.Contains(view, "restore error") {
		t.Fatal("should surface individual rollback errors")
	}
}

func TestRenderComplete_WithRollbackNoSteps(t *testing.T) {
	result := pipeline.ExecutionResult{
		Err:      errors.New("failed"),
		Rollback: &pipeline.StageResult{},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Rollback") {
		t.Fatal("should mention rollback")
	}
	if !strings.Contains(view, "Original configs restored") {
		t.Fatal("should show generic rollback message when no steps")
	}
}

func TestRenderComplete_EmptySteps(t *testing.T) {
	result := pipeline.ExecutionResult{}
	view := RenderComplete(result)
	if view == "" {
		t.Fatal("should render even with empty result")
	}
}

func TestRenderComplete_WithInstallSteps(t *testing.T) {
	result := pipeline.ExecutionResult{
		Install: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "install-opencode", Status: pipeline.StatusSucceeded},
				{StepID: "install-engram", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Tools Installed") {
		t.Fatal("should show install section header")
	}
	if !strings.Contains(view, "install-opencode") {
		t.Fatal("should show install step results")
	}
}

func TestRenderComplete_WithDeploySteps(t *testing.T) {
	result := pipeline.ExecutionResult{
		Deploy: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "deploy-dotfiles", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Dotfiles Deployed") {
		t.Fatal("should show deploy section header")
	}
	if !strings.Contains(view, "deploy-dotfiles") {
		t.Fatal("should show deploy step results")
	}
}

func TestRenderComplete_AllStages(t *testing.T) {
	result := pipeline.ExecutionResult{
		Install: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "install-opencode", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
		Deploy: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "deploy-dotfiles", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
		Apply: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "inject-claude-engram", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
	}
	view := RenderComplete(result)
	if !strings.Contains(view, "Tools Installed") {
		t.Fatal("should show install section")
	}
	if !strings.Contains(view, "Dotfiles Deployed") {
		t.Fatal("should show deploy section")
	}
	if !strings.Contains(view, "MCP Servers Configured") {
		t.Fatal("should show apply section")
	}
}

// --- Common helpers ---

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
