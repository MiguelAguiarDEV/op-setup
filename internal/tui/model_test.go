package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
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
