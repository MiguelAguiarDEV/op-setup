package pipeline

import (
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/installer"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
	"github.com/MiguelAguiarDEV/op-setup/internal/pipeline/steps"
)

func TestPlanner_PlanMCP_TwoAgentsThreeComponents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	plan, err := planner.PlanMCP(
		[]model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
		[]model.ComponentID{model.ComponentEngram, model.ComponentContext7, model.ComponentPlaywright},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Prepare: validate + backup = 2 steps
	if len(plan.Prepare) != 2 {
		t.Fatalf("expected 2 prepare steps, got %d", len(plan.Prepare))
	}
	if plan.Prepare[0].ID() != "validate-deps" {
		t.Fatalf("first prepare step = %q, want %q", plan.Prepare[0].ID(), "validate-deps")
	}
	if plan.Prepare[1].ID() != "backup-configs" {
		t.Fatalf("second prepare step = %q, want %q", plan.Prepare[1].ID(), "backup-configs")
	}

	// Apply: 1 inject per agent = 2 steps
	if len(plan.Apply) != 2 {
		t.Fatalf("expected 2 apply steps, got %d", len(plan.Apply))
	}
	if plan.Apply[0].ID() != "inject-claude-code" {
		t.Fatalf("first apply step = %q, want %q", plan.Apply[0].ID(), "inject-claude-code")
	}
	if plan.Apply[1].ID() != "inject-opencode" {
		t.Fatalf("second apply step = %q, want %q", plan.Apply[1].ID(), "inject-opencode")
	}

	// MCP-only: no install or deploy steps.
	if len(plan.Install) != 0 {
		t.Fatalf("expected 0 install steps, got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 0 {
		t.Fatalf("expected 0 deploy steps, got %d", len(plan.Deploy))
	}
}

func TestPlanner_PlanMCP_NoAgents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.PlanMCP(nil, []model.ComponentID{model.ComponentEngram})
	if err == nil {
		t.Fatal("expected error for no agents")
	}
}

func TestPlanner_PlanMCP_NoComponents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.PlanMCP([]model.AgentID{model.AgentClaudeCode}, nil)
	if err == nil {
		t.Fatal("expected error for no components")
	}
}

func TestPlanner_Plan_UnknownAgent(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.Plan(model.ProfileMCPOnly,
		[]model.AgentID{"unknown"},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err == nil {
		t.Fatal("expected error for unknown agent")
	}
}

func TestPlanner_Plan_UnknownComponent(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.Plan(model.ProfileMCPOnly,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{"unknown"},
	)
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
}

func TestPlanner_PlanMCP_AllAgentsAllComponents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	plan, err := planner.PlanMCP(model.AllAgents(), model.AllComponents())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Prepare) != 2 {
		t.Fatalf("expected 2 prepare steps, got %d", len(plan.Prepare))
	}
	if len(plan.Apply) != 4 {
		t.Fatalf("expected 4 apply steps (one per agent), got %d", len(plan.Apply))
	}
}

func TestPlanner_Plan_Full(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	plan, err := planner.Plan(model.ProfileFull,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Full: Prepare + Deploy + Apply (no Install without InstallerRegistry).
	if len(plan.Prepare) != 2 {
		t.Fatalf("expected 2 prepare steps, got %d", len(plan.Prepare))
	}
	if len(plan.Install) != 0 {
		t.Fatalf("expected 0 install steps (no InstallerRegistry), got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 1 {
		t.Fatalf("expected 1 deploy step, got %d", len(plan.Deploy))
	}
	if plan.Deploy[0].ID() != "deploy-dotfiles" {
		t.Fatalf("deploy step = %q, want %q", plan.Deploy[0].ID(), "deploy-dotfiles")
	}
	if len(plan.Apply) != 1 {
		t.Fatalf("expected 1 apply step, got %d", len(plan.Apply))
	}
}

func TestPlanner_Plan_FullWithInstallers(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	installerReg, _ := installer.NewDefaultRegistry("/home/test")
	planner := NewPlanner(registry, "/home/test")
	planner.InstallerRegistry = installerReg

	plan, err := planner.Plan(model.ProfileFull,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Install) != 4 {
		t.Fatalf("expected 4 install steps, got %d", len(plan.Install))
	}
}

func TestPlanner_Plan_DotfilesOnly(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	plan, err := planner.Plan(model.ProfileDotfilesOnly, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Prepare) != 0 {
		t.Fatalf("expected 0 prepare steps, got %d", len(plan.Prepare))
	}
	if len(plan.Install) != 0 {
		t.Fatalf("expected 0 install steps, got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 1 {
		t.Fatalf("expected 1 deploy step, got %d", len(plan.Deploy))
	}
	if len(plan.Apply) != 0 {
		t.Fatalf("expected 0 apply steps, got %d", len(plan.Apply))
	}
}

func TestPlanner_Plan_UnsupportedProfile(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.Plan("invalid", nil, nil)
	if err == nil {
		t.Fatal("expected error for unsupported profile")
	}
}

func TestPlanner_Plan_Full_IncludesInstallerPrereqs(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	installerReg, _ := installer.NewDefaultRegistry("/home/test")
	planner := NewPlanner(registry, "/home/test")
	planner.InstallerRegistry = installerReg

	plan, err := planner.Plan(model.ProfileFull,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// First prepare step should be ValidateStep.
	validateStep, ok := plan.Prepare[0].(*steps.ValidateStep)
	if !ok {
		t.Fatal("first prepare step should be ValidateStep")
	}

	// Installer prerequisites: npm (opencode, context-mode), go (engram), npx (playwright).
	prereqBinaries := map[string]bool{"npm": false, "go": false, "npx": false}
	for _, check := range validateStep.Checks {
		if _, exists := prereqBinaries[check.Binary]; exists {
			prereqBinaries[check.Binary] = true
			if !check.Required {
				t.Fatalf("installer prereq %q should be Required=true", check.Binary)
			}
		}
	}
	for binary, found := range prereqBinaries {
		if !found {
			t.Fatalf("expected installer prereq %q in validation checks", binary)
		}
	}
}

func TestBuildPlan_Full_WithInstallerRegistry(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	installerReg, _ := installer.NewDefaultRegistry("/home/test")

	plan, err := BuildPlan(registry, installerReg, "/home/test", model.ProfileFull,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Install) != 4 {
		t.Fatalf("expected 4 install steps, got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 1 {
		t.Fatalf("expected 1 deploy step, got %d", len(plan.Deploy))
	}
	if len(plan.Apply) != 1 {
		t.Fatalf("expected 1 apply step, got %d", len(plan.Apply))
	}
}

func TestBuildPlan_Full_NilInstallerRegistry(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()

	plan, err := BuildPlan(registry, nil, "/home/test", model.ProfileFull,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Install) != 0 {
		t.Fatalf("expected 0 install steps (nil registry), got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 1 {
		t.Fatalf("expected 1 deploy step, got %d", len(plan.Deploy))
	}
}

func TestBuildPlan_MCPOnly(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	installerReg, _ := installer.NewDefaultRegistry("/home/test")

	plan, err := BuildPlan(registry, installerReg, "/home/test", model.ProfileMCPOnly,
		[]model.AgentID{model.AgentClaudeCode, model.AgentOpenCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Install) != 0 {
		t.Fatalf("expected 0 install steps for MCPOnly, got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 0 {
		t.Fatalf("expected 0 deploy steps for MCPOnly, got %d", len(plan.Deploy))
	}
	if len(plan.Apply) != 2 {
		t.Fatalf("expected 2 apply steps, got %d", len(plan.Apply))
	}
}

func TestBuildPlan_DotfilesOnly(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()

	plan, err := BuildPlan(registry, nil, "/home/test", model.ProfileDotfilesOnly,
		nil, nil,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(plan.Prepare) != 0 {
		t.Fatalf("expected 0 prepare steps, got %d", len(plan.Prepare))
	}
	if len(plan.Install) != 0 {
		t.Fatalf("expected 0 install steps, got %d", len(plan.Install))
	}
	if len(plan.Deploy) != 1 {
		t.Fatalf("expected 1 deploy step, got %d", len(plan.Deploy))
	}
	if len(plan.Apply) != 0 {
		t.Fatalf("expected 0 apply steps, got %d", len(plan.Apply))
	}
}

func TestPlanner_Plan_MCPOnly_NoInstallerPrereqs(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	installerReg, _ := installer.NewDefaultRegistry("/home/test")
	planner := NewPlanner(registry, "/home/test")
	planner.InstallerRegistry = installerReg

	plan, err := planner.Plan(model.ProfileMCPOnly,
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{model.ComponentEngram},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	validateStep, ok := plan.Prepare[0].(*steps.ValidateStep)
	if !ok {
		t.Fatal("first prepare step should be ValidateStep")
	}

	// MCPOnly should NOT include installer prerequisites.
	installerPrereqs := map[string]bool{"npm": true, "go": true, "npx": true}
	for _, check := range validateStep.Checks {
		if installerPrereqs[check.Binary] && check.Required {
			t.Fatalf("MCPOnly should not include required installer prereq %q", check.Binary)
		}
	}
}
