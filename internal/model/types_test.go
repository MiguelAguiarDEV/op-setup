package model

import "testing"

func TestAgentIDConstants_Unique(t *testing.T) {
	agents := AllAgents()
	seen := make(map[AgentID]bool, len(agents))
	for _, a := range agents {
		if seen[a] {
			t.Fatalf("duplicate AgentID: %q", a)
		}
		seen[a] = true
	}
	if len(agents) != 4 {
		t.Fatalf("expected 4 agents, got %d", len(agents))
	}
}

func TestComponentIDConstants_Unique(t *testing.T) {
	components := AllComponents()
	seen := make(map[ComponentID]bool, len(components))
	for _, c := range components {
		if seen[c] {
			t.Fatalf("duplicate ComponentID: %q", c)
		}
		seen[c] = true
	}
	if len(components) != 5 {
		t.Fatalf("expected 5 components, got %d", len(components))
	}
}

func TestMCPStrategy_Distinct(t *testing.T) {
	if StrategyMergeIntoJSON == StrategyMergeIntoTOML {
		t.Fatal("StrategyMergeIntoJSON and StrategyMergeIntoTOML must be distinct")
	}
}

func TestMCPType_Values(t *testing.T) {
	if MCPTypeLocal == MCPTypeRemote {
		t.Fatal("MCPTypeLocal and MCPTypeRemote must be distinct")
	}
	if MCPTypeLocal != "local" {
		t.Fatalf("MCPTypeLocal = %q, want %q", MCPTypeLocal, "local")
	}
	if MCPTypeRemote != "remote" {
		t.Fatalf("MCPTypeRemote = %q, want %q", MCPTypeRemote, "remote")
	}
}

func TestAllAgents_Order(t *testing.T) {
	agents := AllAgents()
	expected := []AgentID{AgentClaudeCode, AgentOpenCode, AgentCodex, AgentGeminiCLI}
	for i, a := range agents {
		if a != expected[i] {
			t.Fatalf("AllAgents()[%d] = %q, want %q", i, a, expected[i])
		}
	}
}

func TestAllComponents_Order(t *testing.T) {
	components := AllComponents()
	expected := []ComponentID{ComponentEngram, ComponentContextMode, ComponentPlaywright, ComponentGitHubMCP, ComponentContext7}
	for i, c := range components {
		if c != expected[i] {
			t.Fatalf("AllComponents()[%d] = %q, want %q", i, c, expected[i])
		}
	}
}
