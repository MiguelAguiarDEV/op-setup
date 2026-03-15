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

// --- InstallerID tests ---

func TestInstallerIDConstants_Unique(t *testing.T) {
	installers := AllInstallers()
	seen := make(map[InstallerID]bool, len(installers))
	for _, id := range installers {
		if seen[id] {
			t.Fatalf("duplicate InstallerID: %q", id)
		}
		seen[id] = true
	}
	if len(installers) != 4 {
		t.Fatalf("expected 4 installers, got %d", len(installers))
	}
}

func TestAllInstallers_Order(t *testing.T) {
	installers := AllInstallers()
	expected := []InstallerID{InstallerOpenCode, InstallerEngram, InstallerContextMode, InstallerPlaywright}
	for i, id := range installers {
		if id != expected[i] {
			t.Fatalf("AllInstallers()[%d] = %q, want %q", i, id, expected[i])
		}
	}
}

func TestInstallerIDConstants_Values(t *testing.T) {
	tests := []struct {
		id   InstallerID
		want string
	}{
		{InstallerOpenCode, "opencode"},
		{InstallerEngram, "engram"},
		{InstallerContextMode, "context-mode"},
		{InstallerPlaywright, "playwright"},
	}
	for _, tt := range tests {
		if string(tt.id) != tt.want {
			t.Fatalf("InstallerID %q != %q", tt.id, tt.want)
		}
	}
}

// --- SetupProfile tests ---

func TestAllProfiles_Count(t *testing.T) {
	profiles := AllProfiles()
	if len(profiles) != 3 {
		t.Fatalf("expected 3 profiles, got %d", len(profiles))
	}
}

func TestAllProfiles_Order(t *testing.T) {
	profiles := AllProfiles()
	expected := []SetupProfile{ProfileFull, ProfileMCPOnly, ProfileDotfilesOnly}
	for i, p := range profiles {
		if p != expected[i] {
			t.Fatalf("AllProfiles()[%d] = %q, want %q", i, p, expected[i])
		}
	}
}

func TestSetupProfile_Name(t *testing.T) {
	tests := []struct {
		profile SetupProfile
		want    string
	}{
		{ProfileFull, "Full Setup"},
		{ProfileMCPOnly, "MCP Servers Only"},
		{ProfileDotfilesOnly, "Dotfiles Only"},
		{SetupProfile("unknown"), "unknown"},
	}
	for _, tt := range tests {
		got := tt.profile.Name()
		if got != tt.want {
			t.Fatalf("SetupProfile(%q).Name() = %q, want %q", tt.profile, got, tt.want)
		}
	}
}

func TestSetupProfile_Description(t *testing.T) {
	tests := []struct {
		profile SetupProfile
		want    string
	}{
		{ProfileFull, "Install tools, deploy dotfiles, and configure MCP servers"},
		{ProfileMCPOnly, "Only configure MCP servers in AI tool config files"},
		{ProfileDotfilesOnly, "Only deploy agents, skills, scripts, and nvim config"},
		{SetupProfile("unknown"), ""},
	}
	for _, tt := range tests {
		got := tt.profile.Description()
		if got != tt.want {
			t.Fatalf("SetupProfile(%q).Description() = %q, want %q", tt.profile, got, tt.want)
		}
	}
}

func TestSetupProfile_StringValues(t *testing.T) {
	tests := []struct {
		profile SetupProfile
		want    string
	}{
		{ProfileFull, "full"},
		{ProfileMCPOnly, "mcp-only"},
		{ProfileDotfilesOnly, "dotfiles-only"},
	}
	for _, tt := range tests {
		if string(tt.profile) != tt.want {
			t.Fatalf("SetupProfile string = %q, want %q", tt.profile, tt.want)
		}
	}
}

// --- AgentID constant values ---

func TestAgentIDConstants_Values(t *testing.T) {
	tests := []struct {
		id   AgentID
		want string
	}{
		{AgentClaudeCode, "claude-code"},
		{AgentOpenCode, "opencode"},
		{AgentCodex, "codex"},
		{AgentGeminiCLI, "gemini-cli"},
	}
	for _, tt := range tests {
		if string(tt.id) != tt.want {
			t.Fatalf("AgentID %q != %q", tt.id, tt.want)
		}
	}
}

// --- ComponentID constant values ---

func TestComponentIDConstants_Values(t *testing.T) {
	tests := []struct {
		id   ComponentID
		want string
	}{
		{ComponentEngram, "engram"},
		{ComponentContextMode, "context-mode"},
		{ComponentPlaywright, "playwright"},
		{ComponentGitHubMCP, "github"},
		{ComponentContext7, "context7"},
	}
	for _, tt := range tests {
		if string(tt.id) != tt.want {
			t.Fatalf("ComponentID %q != %q", tt.id, tt.want)
		}
	}
}

// --- MCPServerConfig struct ---

func TestMCPServerConfig_Fields(t *testing.T) {
	cfg := MCPServerConfig{
		Type:    MCPTypeLocal,
		Command: []string{"npx", "server"},
		Enabled: true,
	}
	if cfg.Type != MCPTypeLocal {
		t.Fatal("Type mismatch")
	}
	if len(cfg.Command) != 2 {
		t.Fatal("Command length mismatch")
	}
	if !cfg.Enabled {
		t.Fatal("Enabled should be true")
	}
	if cfg.URL != "" {
		t.Fatal("URL should be empty for local type")
	}
}

func TestMCPServerConfig_Remote(t *testing.T) {
	cfg := MCPServerConfig{
		Type:    MCPTypeRemote,
		URL:     "https://example.com/mcp",
		Headers: map[string]string{"Authorization": "Bearer token"},
		Enabled: true,
	}
	if cfg.Type != MCPTypeRemote {
		t.Fatal("Type mismatch")
	}
	if cfg.URL != "https://example.com/mcp" {
		t.Fatal("URL mismatch")
	}
	if cfg.Headers["Authorization"] != "Bearer token" {
		t.Fatal("Headers mismatch")
	}
}

// --- DetectResult struct ---

func TestDetectResult_Installed(t *testing.T) {
	r := DetectResult{
		Installed:   true,
		BinaryPath:  "/usr/bin/claude",
		ConfigPath:  "/home/user/.claude/config.json",
		ConfigFound: true,
	}
	if !r.Installed {
		t.Fatal("should be installed")
	}
	if r.BinaryPath != "/usr/bin/claude" {
		t.Fatal("BinaryPath mismatch")
	}
	if !r.ConfigFound {
		t.Fatal("ConfigFound should be true")
	}
}

func TestDetectResult_NotInstalled(t *testing.T) {
	r := DetectResult{}
	if r.Installed {
		t.Fatal("should not be installed")
	}
	if r.BinaryPath != "" {
		t.Fatal("BinaryPath should be empty")
	}
	if r.ConfigFound {
		t.Fatal("ConfigFound should be false")
	}
}

// --- MCPStrategy values ---

func TestMCPStrategy_Values(t *testing.T) {
	if StrategyMergeIntoJSON != 0 {
		t.Fatalf("StrategyMergeIntoJSON = %d, want 0", StrategyMergeIntoJSON)
	}
	if StrategyMergeIntoTOML != 1 {
		t.Fatalf("StrategyMergeIntoTOML = %d, want 1", StrategyMergeIntoTOML)
	}
}
