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
	return NewModel(registry, "test", "/home/test")
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
	// All components selected by default.
	for _, c := range m.components {
		if !c.selected {
			t.Fatalf("component %q should be selected by default", c.component.ID)
		}
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

func TestModel_Enter_WelcomeToDetection(t *testing.T) {
	m := newTestModel(t)
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	um := updated.(Model)

	if um.screen != ScreenDetection {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenDetection)
	}
	if cmd == nil {
		t.Fatal("should return detect command")
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

func TestModel_Esc_GoesBack(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenDetection {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenDetection)
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

// --- Esc from various screens ---

func TestModel_Esc_FromComponents(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenComponents

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenAgents {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenAgents)
	}
}

func TestModel_Esc_FromReview(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenReview

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.screen != ScreenComponents {
		t.Fatalf("screen = %d, want %d", um.screen, ScreenComponents)
	}
}

func TestModel_Esc_ResetssCursor(t *testing.T) {
	m := newTestModel(t)
	m.screen = ScreenAgents
	m.cursor = 3

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	um := updated.(Model)

	if um.cursor != 0 {
		t.Fatalf("cursor = %d, want 0 (should reset on esc)", um.cursor)
	}
}
