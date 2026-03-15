package pipeline

import (
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

func TestPlanner_Plan_TwoAgentsThreeComponents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	plan, err := planner.Plan(
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
}

func TestPlanner_Plan_NoAgents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.Plan(nil, []model.ComponentID{model.ComponentEngram})
	if err == nil {
		t.Fatal("expected error for no agents")
	}
}

func TestPlanner_Plan_NoComponents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.Plan([]model.AgentID{model.AgentClaudeCode}, nil)
	if err == nil {
		t.Fatal("expected error for no components")
	}
}

func TestPlanner_Plan_UnknownAgent(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	_, err := planner.Plan(
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

	_, err := planner.Plan(
		[]model.AgentID{model.AgentClaudeCode},
		[]model.ComponentID{"unknown"},
	)
	if err == nil {
		t.Fatal("expected error for unknown component")
	}
}

func TestPlanner_Plan_AllAgentsAllComponents(t *testing.T) {
	registry, _ := adapter.NewDefaultRegistry()
	planner := NewPlanner(registry, "/home/test")

	plan, err := planner.Plan(model.AllAgents(), model.AllComponents())
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
