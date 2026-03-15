package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
)

func newTestModel(t *testing.T) Model {
	t.Helper()
	registry, err := adapter.NewDefaultRegistry()
	if err != nil {
		t.Fatal(err)
	}
	return NewModel(registry, nil, nil, "test", "/home/test")
}

func TestNewModel_InitialState(t *testing.T) {
	m := newTestModel(t)

	if m.screen != ScreenWelcome {
		t.Fatalf("initial screen = %d, want %d", m.screen, ScreenWelcome)
	}
	if m.version != "test" {
		t.Fatalf("version = %q, want %q", m.version, "test")
	}
	if len(m.agents) != 4 {
		t.Fatalf("expected 4 agents, got %d", len(m.agents))
	}
	if len(m.components) != 5 {
		t.Fatalf("expected 5 components, got %d", len(m.components))
	}
	// Components selected by default only if env vars are satisfied.
	for _, c := range m.components {
		envOK := componentEnvSatisfied(c.component)
		if envOK && !c.selected {
			t.Fatalf("component %q should be selected (env vars satisfied)", c.component.ID)
		}
		if !envOK && c.selected {
			t.Fatalf("component %q should NOT be selected (missing env vars)", c.component.ID)
		}
	}
	// Profile items populated.
	if len(m.profileItems) != 3 {
		t.Fatalf("expected 3 profile items, got %d", len(m.profileItems))
	}
}

func TestModel_Init_ReturnsNil(t *testing.T) {
	m := newTestModel(t)
	cmd := m.Init()
	if cmd != nil {
		t.Fatal("Init should return nil")
	}
}

func TestModel_View_Welcome(t *testing.T) {
	m := newTestModel(t)
	view := m.View()
	if view == "" {
		t.Fatal("welcome view should not be empty")
	}
}

func TestModel_Quit_OnQ(t *testing.T) {
	m := newTestModel(t)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	_ = updated
	if cmd == nil {
		t.Fatal("q should produce quit command")
	}
}

func TestModel_Quit_OnCtrlC(t *testing.T) {
	m := newTestModel(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Fatal("ctrl+c should produce quit command")
	}
}

// --- Profile selection flow ---

func TestModel_Enter_WelcomeToProfile(t *testing.T) {
	m := newTestModel(t)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenProfile {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenProfile)
	}
	if cmd != nil {
		t.Fatal("welcome→profile should not produce a command")
	}
}

func TestModel_Enter_ProfileToDetection_MCPOnly(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile
	m.cursor = 1 // MCPOnly is second item.

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenDetection {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenDetection)
	}
	if um.profile != model.ProfileMCPOnly {
		t.Fatalf("profile = %q, want %q", um.profile, model.ProfileMCPOnly)
	}
	if cmd == nil {
		t.Fatal("profile→detection should produce detect command")
	}
}

func TestModel_Enter_ProfileToDetection_Full(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile
	m.cursor = 0 // Full is first item.

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenDetection {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenDetection)
	}
	if um.profile != model.ProfileFull {
		t.Fatalf("profile = %q, want %q", um.profile, model.ProfileFull)
	}
	if cmd == nil {
		t.Fatal("profile→detection should produce detect command")
	}
}

func TestModel_Enter_ProfileToReview_DotfilesOnly(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile
	m.cursor = 2 // DotfilesOnly is third item.

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenReview {
		t.Fatalf("screen = %d, want %d (should skip to review)", um.screen, ScreenReview)
	}
	if um.profile != model.ProfileDotfilesOnly {
		t.Fatalf("profile = %q, want %q", um.profile, model.ProfileDotfilesOnly)
	}
	if cmd != nil {
		t.Fatal("profile→review (dotfiles) should not produce a command")
	}
}

func TestModel_DetectDone_UpdatesAgents(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenDetection

	results := map[model.AgentID]model.DetectResult{
		model.AgentClaudeCode: {Installed: true, ConfigFound: true},
		model.AgentOpenCode:   {Installed: true, ConfigFound: false},
		model.AgentCodex:      {Installed: false},
		model.AgentGeminiCLI:  {Installed: false},
	}

	updated, _ := m.Update(detectDoneMsg{results: results})
	um := updated.(Model)

	// Check that detected agents are auto-selected.
	for _, a := range um.agents {
		switch a.adapter.Agent() {
		case model.AgentClaudeCode, model.AgentOpenCode:
			if !a.detected || !a.selected {
				t.Fatalf("agent %q should be detected and selected", a.adapter.Agent())
			}
		case model.AgentCodex, model.AgentGeminiCLI:
			if a.detected || a.selected {
				t.Fatalf("agent %q should not be detected or selected", a.adapter.Agent())
			}
		}
	}
}

func TestModel_Navigation_UpDown(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents

	// Move down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", um.cursor)
	}

	// Move up.
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um = updated.(Model)
	if um.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", um.cursor)
	}

	// Can't go below 0.
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	um = updated.(Model)
	if um.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (can't go below)", um.cursor)
	}
}

func TestModel_Space_TogglesSelection(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents

	initial := m.components[0].selected

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	um := updated.(Model)

	if um.components[0].selected == initial {
		t.Fatal("space should toggle selection")
	}
}

func TestModel_Space_BlocksMissingEnvVar(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents

	// Find the GitHub MCP component (has GITHUB_MCP_PAT env var).
	ghIdx := -1
	for i, c := range m.components {
		if c.component.ID == model.ComponentGitHubMCP {
			ghIdx = i
			break
		}
	}
	if ghIdx == -1 {
		t.Fatal("GitHub MCP component not found")
	}

	// Ensure env var is unset.
	t.Setenv("GITHUB_MCP_PAT", "")

	m.cursor = ghIdx
	initial := m.components[ghIdx].selected

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	um := updated.(Model)

	if um.components[ghIdx].selected != initial {
		t.Fatal("space should NOT toggle component with missing env var")
	}
}

func TestModel_Space_AllowsWithEnvVar(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents

	// Find the GitHub MCP component.
	ghIdx := -1
	for i, c := range m.components {
		if c.component.ID == model.ComponentGitHubMCP {
			ghIdx = i
			break
		}
	}
	if ghIdx == -1 {
		t.Fatal("GitHub MCP component not found")
	}

	// Set the env var.
	t.Setenv("GITHUB_MCP_PAT", "test-token")

	m.cursor = ghIdx
	initial := m.components[ghIdx].selected

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	um := updated.(Model)

	if um.components[ghIdx].selected == initial {
		t.Fatal("space should toggle component when env var is set")
	}
}

// --- Esc behavior ---

func TestModel_Esc_FromProfile(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenWelcome {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenWelcome)
	}
}

func TestModel_Esc_FromDetection(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenDetection

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenProfile {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenProfile)
	}
}

func TestModel_Esc_FromAgents(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenDetection {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenDetection)
	}
}

func TestModel_Esc_FromComponents(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenAgents {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenAgents)
	}
}

func TestModel_Esc_FromReview_MCPOnly(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenReview
	m.profile = model.ProfileMCPOnly

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenComponents {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenComponents)
	}
}

func TestModel_Esc_FromReview_DotfilesOnly(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenReview
	m.profile = model.ProfileDotfilesOnly

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenProfile {
		t.Fatalf("screen = %d, want %d (should go back to profile)", um.screen, ScreenProfile)
	}
}

func TestModel_Esc_WelcomeStays(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenWelcome

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenWelcome {
		t.Fatalf("screen = %d, want %d (should stay)", um.screen, ScreenWelcome)
	}
}

func TestModel_Esc_ResetsCursor(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	m.cursor = 3

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (should reset on esc)", um.cursor)
	}
}

func TestModel_Agents_CantSelectUndetected(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	// All agents undetected by default.

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	um := updated.(Model)

	if um.agents[0].selected {
		t.Fatal("should not be able to select undetected agent")
	}
}

func TestModel_View_AllScreens(t *testing.T) {
	m := newTestModel(t)

	allScreens := []Screen{
		ScreenWelcome,
		ScreenProfile,
		ScreenDetection,
		ScreenAgents,
		ScreenComponents,
		ScreenReview,
		ScreenInstalling,
		ScreenComplete,
	}

	for _, s := range allScreens {
		m.screen = s
		view := m.View()
		if view == "" && s != ScreenComplete {
			t.Fatalf("view for screen %d should not be empty", s)
		}
	}
}

func TestModel_View_Profile(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile

	view := m.View()
	if view == "" {
		t.Fatal("profile view should not be empty")
	}
}

// --- handleEnter flow tests ---

func TestModel_Enter_DetectionToAgents(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenDetection

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenAgents {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenAgents)
	}
	if cmd != nil {
		t.Fatal("detection→agents should not produce a command")
	}
}

func TestModel_Enter_AgentsToComponents_WithSelection(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	// Mark first agent as detected and selected.
	m.agents[0].detected = true
	m.agents[0].selected = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenComponents {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenComponents)
	}
	if um.cursor != 0 {
		t.Fatalf("cursor should reset to 0, got %d", um.cursor)
	}
}

func TestModel_Enter_AgentsBlocked_NoSelection(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	// No agents selected.

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenAgents {
		t.Fatalf("screen = %d, want %d (should stay)", um.screen, ScreenAgents)
	}
}

func TestModel_Enter_ComponentsToReview_WithSelection(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents
	// Components are selected by default.

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenReview {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenReview)
	}
}

func TestModel_Enter_ComponentsBlocked_NoSelection(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents
	// Deselect all components.
	for i := range m.components {
		m.components[i].selected = false
	}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenComponents {
		t.Fatalf("screen = %d, want %d (should stay)", um.screen, ScreenComponents)
	}
}

func TestModel_Enter_ReviewToInstalling(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenReview
	m.profile = model.ProfileMCPOnly
	// Need at least one agent selected for installCmd.
	m.agents[0].detected = true
	m.agents[0].selected = true

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenInstalling {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenInstalling)
	}
	if cmd == nil {
		t.Fatal("review→installing should produce install command")
	}
}

func TestModel_Enter_CompleteQuits(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComplete

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("enter on complete should produce quit command")
	}
}

func TestModel_NoQuit_DuringInstalling(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenInstalling

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd != nil {
		t.Fatal("q during installing should not quit")
	}

	_, cmd = m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd != nil {
		t.Fatal("ctrl+c during installing should not quit")
	}
}

// --- Navigation edge cases ---

func TestModel_Navigation_ArrowKeys(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents

	// Arrow down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	um := updated.(Model)
	if um.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", um.cursor)
	}

	// Arrow up.
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyUp})
	um = updated.(Model)
	if um.cursor != 0 {
		t.Fatalf("cursor = %d, want 0", um.cursor)
	}
}

func TestModel_Navigation_CantExceedMax(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	m.cursor = len(m.agents) - 1 // At last position.

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.cursor != len(m.agents)-1 {
		t.Fatalf("cursor = %d, want %d (should not exceed max)", um.cursor, len(m.agents)-1)
	}
}

func TestModel_MaxCursor_Profile(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile
	max := m.maxCursor()
	if max != 2 {
		t.Fatalf("maxCursor for profile = %d, want 2", max)
	}
}

func TestModel_MaxCursor_Components(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents
	m.cursor = len(m.components) - 1

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.cursor != len(m.components)-1 {
		t.Fatalf("cursor = %d, want %d", um.cursor, len(m.components)-1)
	}
}

func TestModel_MaxCursor_OtherScreens(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenWelcome
	max := m.maxCursor()
	if max != 0 {
		t.Fatalf("maxCursor for welcome = %d, want 0", max)
	}
}

// --- Space toggle on agents ---

func TestModel_Space_TogglesDetectedAgent(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	m.agents[0].detected = true
	m.agents[0].selected = true

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	um := updated.(Model)

	if um.agents[0].selected {
		t.Fatal("space should deselect detected agent")
	}
}

// --- Progress message ---

func TestModel_ProgressMsg_AppendsEvent(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenInstalling

	event := pipeline.ProgressEvent{
		Stage:  pipeline.StageApply,
		StepID: "test-step",
		Status: pipeline.StatusRunning,
	}

	updated, _ := m.Update(progressMsg{event: event})
	um := updated.(Model)

	if len(um.progressEvents) != 1 {
		t.Fatalf("expected 1 progress event, got %d", len(um.progressEvents))
	}
	if um.progressEvents[0].StepID != "test-step" {
		t.Fatalf("step ID = %q, want %q", um.progressEvents[0].StepID, "test-step")
	}
}

// --- Total steps message ---

func TestModel_TotalStepsMsg_SetsTotalSteps(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenInstalling

	updated, _ := m.Update(totalStepsMsg{count: 7})
	um := updated.(Model)

	if um.totalSteps != 7 {
		t.Fatalf("totalSteps = %d, want 7", um.totalSteps)
	}
}

// --- Install done message ---

func TestModel_InstallDone_GoesToComplete(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenInstalling

	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{Success: true},
	}

	updated, _ := m.Update(installDoneMsg{result: result})
	um := updated.(Model)

	if um.screen != ScreenComplete {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenComplete)
	}
	if um.result == nil {
		t.Fatal("result should be set")
	}
}

// --- View with result ---

func TestModel_View_CompleteWithResult(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComplete
	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "test", Status: pipeline.StatusSucceeded},
			},
			Success: true,
		},
	}
	m.result = &result

	view := m.View()
	if view == "" {
		t.Fatal("complete view with result should not be empty")
	}
}

// --- Unknown message type ---

func TestModel_UnknownMsg_NoOp(t *testing.T) {
	m := newTestModel(t)
	type unknownMsg struct{}

	updated, cmd := m.Update(unknownMsg{})
	um := updated.(Model)

	if um.screen != ScreenWelcome {
		t.Fatalf("screen should not change on unknown msg")
	}
	if cmd != nil {
		t.Fatal("unknown msg should not produce command")
	}
}

// --- Profile navigation ---

func TestModel_Navigation_ProfileUpDown(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenProfile

	// Move down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um := updated.(Model)
	if um.cursor != 1 {
		t.Fatalf("cursor = %d, want 1", um.cursor)
	}

	// Move down again.
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um = updated.(Model)
	if um.cursor != 2 {
		t.Fatalf("cursor = %d, want 2", um.cursor)
	}

	// Can't go past max.
	updated, _ = um.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	um = updated.(Model)
	if um.cursor != 2 {
		t.Fatalf("cursor = %d, want 2 (should not exceed max)", um.cursor)
	}
}

// --- Esc from installing/complete stays ---

func TestModel_Esc_FromInstalling_Stays(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenInstalling

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenInstalling {
		t.Fatalf("screen = %d, want %d (should stay)", um.screen, ScreenInstalling)
	}
}

func TestModel_Esc_FromComplete_Stays(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComplete

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenComplete {
		t.Fatalf("screen = %d, want %d (should stay)", um.screen, ScreenComplete)
	}
}
