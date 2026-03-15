package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline"
	"github.com/MiguelAguiarDEV/op-setup/internal/tui/screens"
)

// Messages
type detectDoneMsg struct {
	results map[model.AgentID]model.DetectResult
}

type installDoneMsg struct {
	result pipeline.ExecutionResult
}

type progressMsg struct {
	event pipeline.ProgressEvent
}

// Model is the main bubbletea model.
type Model struct {
	screen  Screen
	version string
	homeDir string

	// Dependencies
	registry *adapter.Registry

	// Detection
	detected map[model.AgentID]model.DetectResult

	// Selections
	agents     []agentEntry
	components []componentEntry
	cursor     int

	// Installation
	progressEvents []pipeline.ProgressEvent
	totalSteps     int
	result         *pipeline.ExecutionResult

	// Error
	err error
}

type agentEntry struct {
	adapter  adapter.Adapter
	detected bool
	selected bool
}

type componentEntry struct {
	component component.Component
	selected  bool
}

// NewModel creates the TUI model.
func NewModel(registry *adapter.Registry, version string, homeDir string) Model {
	// Build agent entries.
	allAdapters := registry.All()
	agents := make([]agentEntry, len(allAdapters))
	for i, a := range allAdapters {
		agents[i] = agentEntry{adapter: a}
	}

	// Build component entries — all selected by default.
	allComps := component.All()
	comps := make([]componentEntry, len(allComps))
	for i, c := range allComps {
		comps[i] = componentEntry{component: c, selected: true}
	}

	return Model{
		screen:     ScreenWelcome,
		version:    version,
		homeDir:    homeDir,
		registry:   registry,
		detected:   make(map[model.AgentID]model.DetectResult),
		agents:     agents,
		components: comps,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case detectDoneMsg:
		m.detected = msg.results
		// Update agent entries with detection results and auto-select detected.
		for i := range m.agents {
			r, ok := m.detected[m.agents[i].adapter.Agent()]
			if ok {
				m.agents[i].detected = r.Installed
				m.agents[i].selected = r.Installed
			}
		}
		return m, nil
	case progressMsg:
		m.progressEvents = append(m.progressEvents, msg.event)
		return m, nil
	case installDoneMsg:
		m.result = &msg.result
		m.screen = ScreenComplete
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		if m.screen == ScreenInstalling {
			return m, nil // Don't quit during installation.
		}
		return m, tea.Quit

	case "enter":
		return m.handleEnter()

	case "esc":
		if prev, ok := PreviousScreen(m.screen); ok {
			m.screen = prev
			m.cursor = 0
		}
		return m, nil

	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil

	case "down", "j":
		max := m.maxCursor()
		if m.cursor < max {
			m.cursor++
		}
		return m, nil

	case " ":
		return m.handleSpace()
	}

	return m, nil
}

func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenWelcome:
		m.screen = ScreenDetection
		m.cursor = 0
		return m, m.detectCmd()

	case ScreenDetection:
		m.screen = ScreenAgents
		m.cursor = 0
		return m, nil

	case ScreenAgents:
		// Check at least one agent selected.
		hasSelection := false
		for _, a := range m.agents {
			if a.selected {
				hasSelection = true
				break
			}
		}
		if !hasSelection {
			return m, nil
		}
		m.screen = ScreenComponents
		m.cursor = 0
		return m, nil

	case ScreenComponents:
		hasSelection := false
		for _, c := range m.components {
			if c.selected {
				hasSelection = true
				break
			}
		}
		if !hasSelection {
			return m, nil
		}
		m.screen = ScreenReview
		m.cursor = 0
		return m, nil

	case ScreenReview:
		m.screen = ScreenInstalling
		m.progressEvents = nil
		return m, m.installCmd()

	case ScreenComplete:
		return m, tea.Quit
	}

	return m, nil
}

func (m Model) handleSpace() (tea.Model, tea.Cmd) {
	switch m.screen {
	case ScreenAgents:
		if m.cursor < len(m.agents) && m.agents[m.cursor].detected {
			m.agents[m.cursor].selected = !m.agents[m.cursor].selected
		}
	case ScreenComponents:
		if m.cursor < len(m.components) {
			m.components[m.cursor].selected = !m.components[m.cursor].selected
		}
	}
	return m, nil
}

func (m Model) maxCursor() int {
	switch m.screen {
	case ScreenAgents:
		return len(m.agents) - 1
	case ScreenComponents:
		return len(m.components) - 1
	}
	return 0
}

func (m Model) detectCmd() tea.Cmd {
	return func() tea.Msg {
		results := make(map[model.AgentID]model.DetectResult)
		for _, a := range m.registry.All() {
			r, _ := a.Detect(m.homeDir)
			results[a.Agent()] = r
		}
		return detectDoneMsg{results: results}
	}
}

func (m Model) installCmd() tea.Cmd {
	// Collect selections.
	var selectedAgents []model.AgentID
	for _, a := range m.agents {
		if a.selected {
			selectedAgents = append(selectedAgents, a.adapter.Agent())
		}
	}
	var selectedComponents []model.ComponentID
	for _, c := range m.components {
		if c.selected {
			selectedComponents = append(selectedComponents, c.component.ID)
		}
	}

	homeDir := m.homeDir
	registry := m.registry

	return func() tea.Msg {
		planner := pipeline.NewPlanner(registry, homeDir)
		plan, err := planner.Plan(selectedAgents, selectedComponents)
		if err != nil {
			return installDoneMsg{result: pipeline.ExecutionResult{Err: err}}
		}

		// Count total steps for progress.
		totalSteps := len(plan.Prepare) + len(plan.Apply)
		_ = totalSteps

		orchestrator := pipeline.NewOrchestrator(func(e pipeline.ProgressEvent) {
			// Note: In a real TUI, we'd send these via p.Send().
			// For now, they're collected in the result.
		})

		result := orchestrator.Execute(plan)
		return installDoneMsg{result: result}
	}
}

func (m Model) View() string {
	switch m.screen {
	case ScreenWelcome:
		return screens.RenderWelcome(m.version)

	case ScreenDetection:
		items := make([]screens.DetectionItem, len(m.agents))
		for i, a := range m.agents {
			r := m.detected[a.adapter.Agent()]
			items[i] = screens.DetectionItem{
				Name:        a.adapter.Name(),
				Agent:       a.adapter.Agent(),
				Installed:   r.Installed,
				ConfigFound: r.ConfigFound,
			}
		}
		return screens.RenderDetection(items)

	case ScreenAgents:
		items := make([]screens.AgentItem, len(m.agents))
		for i, a := range m.agents {
			items[i] = screens.AgentItem{
				Name:     a.adapter.Name(),
				Detected: a.detected,
				Selected: a.selected,
			}
		}
		return screens.RenderAgents(items, m.cursor)

	case ScreenComponents:
		items := make([]screens.ComponentItem, len(m.components))
		for i, c := range m.components {
			items[i] = screens.ComponentItem{
				Name:     c.component.Name,
				Desc:     c.component.Description,
				Selected: c.selected,
				EnvVars:  c.component.EnvVars,
			}
		}
		return screens.RenderComponents(items, m.cursor)

	case ScreenReview:
		var agentNames, compNames []string
		for _, a := range m.agents {
			if a.selected {
				agentNames = append(agentNames, a.adapter.Name())
			}
		}
		for _, c := range m.components {
			if c.selected {
				compNames = append(compNames, c.component.Name)
			}
		}
		return screens.RenderReview(agentNames, compNames)

	case ScreenInstalling:
		return screens.RenderInstalling(m.progressEvents, m.totalSteps)

	case ScreenComplete:
		if m.result != nil {
			return screens.RenderComplete(*m.result)
		}
		return "Completing..."
	}

	return ""
}

// HomeDir returns the home directory (for testing).
func HomeDir() string {
	h, _ := os.UserHomeDir()
	return h
}
