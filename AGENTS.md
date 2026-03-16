# op-setup — Agent Instructions

## Project Overview

Go CLI that sets up AI coding environments. Single binary: detect tools → install dependencies → deploy dotfiles → configure MCP servers. TUI (Bubbletea) or headless mode.

**Module:** `github.com/MiguelAguiarDEV/op-setup`
**Go version:** 1.24
**Entry point:** `cmd/op-setup/main.go`

## Architecture

```
cmd/op-setup/main.go           Flag parsing → app.Run(cfg)
internal/
├── app/                        Wiring: TUI vs headless (config.go, app.go, headless.go)
├── model/                      Domain types: AgentID, ComponentID, SetupProfile, MCPServerConfig
├── adapter/                    AI tool adapters (interface + registry + factory)
│   ├── claude/                 Claude Code — JSON mcpServers
│   ├── opencode/               OpenCode — JSON mcp + plugin PostInject
│   ├── codex/                  Codex — TOML [mcp_servers.X]
│   └── gemini/                 Gemini CLI — JSON mcpServers
├── component/                  MCP server catalog + resolver
├── installer/                  Tool installers (detect/install/rollback)
├── dotfiles/                   go:embed deployer (embed/ directory)
├── config/                     Atomic writes, JSON merger, TOML merger
├── backup/                     Timestamped snapshots + restore
├── pipeline/                   4-stage orchestrator: Prepare → Install → Deploy → Apply
│   └── steps/                  BackupStep, ValidateStep, InjectStep
├── tui/                        Bubbletea TUI model + 8 screens
│   ├── screens/                Screen renderers (Kanagawa theme)
│   └── styles/                 Shared lipgloss styles
└── xdg/                        XDG_CONFIG_HOME helper
```

## Key Patterns

- **Adapter pattern** — each AI tool is a sub-package implementing `adapter.Adapter` interface
- **Registry pattern** — adapters and installers registered in typed registries
- **DI via struct fields** — all external deps (LookPath, StatPath, ReadFile, WriteFile) injectable
- **Pipeline with cross-stage rollback** — Install fail → rollback Install. Deploy fail → rollback Deploy+Install
- **Functional options** — `tui.ModelOption` for backward-compatible TUI config
- **Atomic writes** — temp file + rename via `config.WriteFileAtomic`
- **`//go:embed all:embed`** — dotfiles embedded at compile time (the `all:` prefix includes dotfiles like `.gitignore`)

## How to Add a New MCP Server

**2 files. That's it.**

### Step 1: `internal/model/types.go`

Add the ComponentID constant and include it in `AllComponents()`:

```go
const (
    ComponentEngram      ComponentID = "engram"
    ComponentContextMode ComponentID = "context-mode"
    // ...
    ComponentMyServer    ComponentID = "my-server"  // ← add
)

func AllComponents() []ComponentID {
    return []ComponentID{
        // ...existing...
        ComponentMyServer,  // ← add
    }
}
```

### Step 2: `internal/component/catalog.go`

Add an entry to the `catalog` slice:

**Local server** (runs a command):
```go
{
    ID:          model.ComponentMyServer,
    Name:        "My Server",
    Description: "What it does",
    Config: model.MCPServerConfig{
        Type:    model.MCPTypeLocal,
        Command: []string{"npx", "-y", "@scope/package"},
        Enabled: true,
    },
},
```

**Remote server** (URL endpoint):
```go
{
    ID:          model.ComponentMyServer,
    Name:        "My Server",
    Description: "What it does",
    Config: model.MCPServerConfig{
        Type:    model.MCPTypeRemote,
        URL:     "https://api.example.com/mcp",
        Enabled: true,
    },
},
```

**Remote server with auth** (requires env var):
```go
{
    ID:          model.ComponentMyServer,
    Name:        "My Server",
    Description: "What it does",
    Config: model.MCPServerConfig{
        Type:    model.MCPTypeRemote,
        URL:     "https://api.example.com/mcp",
        Headers: map[string]string{"Authorization": "Bearer {env:MY_TOKEN}"},
        Enabled: true,
    },
    EnvVars: []string{"MY_TOKEN"},
},
```

### What happens automatically

- TUI shows it in component selection screen
- `{env:X}` resolved at build-time for non-Claude adapters (Claude keeps it as-is)
- Components with missing env vars are deselected by default and can't be toggled
- Headless mode auto-selects it if env vars are satisfied
- Dry-run lists it in the plan
- Injected into all selected adapters' config files (JSON or TOML)

### Tests to update

- `internal/component/catalog_test.go` — update `TestAll_Returns5Components` count
- `internal/model/types_test.go` — update `TestAllComponents` if it checks count
- `internal/tui/model_test.go` — update `TestNewModel_InitialState` component count

## How to Add a New AI Tool Adapter

### Step 1: `internal/model/types.go`

Add AgentID constant and include in `AllAgents()`:

```go
const (
    AgentClaudeCode AgentID = "claude-code"
    // ...
    AgentMyTool     AgentID = "my-tool"  // ← add
)
```

### Step 2: Create `internal/adapter/mytool/adapter.go`

Implement the `adapter.Adapter` interface:

```go
package mytool

type Adapter struct {
    LookPath  func(string) (string, error)
    StatPath  func(string) (os.FileInfo, error)
    ReadFile  func(string) ([]byte, error)
    WriteFile func(string, []byte, os.FileMode) error
}

func NewAdapter() *Adapter { /* wire defaults */ }
func (a *Adapter) Name() string                   { return "My Tool" }
func (a *Adapter) Agent() model.AgentID           { return model.AgentMyTool }
func (a *Adapter) Detect(homeDir string) (model.DetectResult, error) { /* ... */ }
func (a *Adapter) ConfigPath(homeDir string) string { /* ... */ }
func (a *Adapter) MCPStrategy() model.MCPStrategy { return model.StrategyMergeIntoJSON }
func (a *Adapter) MCPConfigKey() string           { return "mcpServers" }
func (a *Adapter) PostInject(homeDir string, components []model.ComponentID) error { return nil }
```

### Step 3: `internal/adapter/factory.go`

Register in `NewAdapter()` switch and `NewDefaultRegistry()`.

### Step 4: Tests

Add `internal/adapter/mytool/adapter_test.go` with identity, detect, config path, and post-inject tests.

## How to Add a New Tool Installer

### Step 1: `internal/model/types.go`

Add InstallerID constant.

### Step 2: Create `internal/installer/mytool.go`

Implement `installer.Installer` interface (ID, Name, Detect, Install, Rollback, Prerequisites).

### Step 3: `internal/installer/factory.go`

Add to `NewDefaultRegistry()` and compile-time interface check.

## Conventions

### Code

- **stdlib testing only** — no testify. Use `t.Fatal`/`t.Fatalf` for assertions.
- **Table-driven tests** with `t.Run` for multiple cases.
- **`t.TempDir()`** for filesystem tests — no mock filesystems (no afero).
- **DI via struct fields** — never call `os.ReadFile` directly in production code; use injected functions.
- **`t.Setenv("XDG_CONFIG_HOME", "")`** in any test that constructs `.config` paths.
- **Target ≥85% coverage** per package.
- **Race detector** enabled by default (`go test -race`).

### Git

- **Conventional commits** — `feat:`, `fix:`, `refactor:`, `ci:`, `docs:`, `test:`
- **No AI attribution** — never add "Co-Authored-By" or similar.
- **No build after changes** — CI handles it.

### File permissions

- All config file writes use `0o600`.
- `config.WriteFileAtomic` for all file writes (temp + rename).

### Env var handling

- `{env:X}` syntax only works natively in Claude Code.
- All other adapters get env vars resolved at build-time by `component.Resolver`.
- `component.EnvSatisfied()` is the single source of truth for env var checks.

## CLI Flags

```
op-setup                                            # TUI (default)
op-setup --dry-run                                  # TUI, no side effects
op-setup --profile mcp-only                         # TUI, pre-select profile
op-setup --no-interactive --profile full             # Headless execution
op-setup --no-interactive --profile full --dry-run   # Headless, print plan only
```

`--no-interactive` requires `--profile`. Flag parsing in `app.BuildConfig()`.

## Pipeline Stages

| Stage | Profile Full | Profile MCP-Only | Profile Dotfiles-Only |
|-------|:-----------:|:----------------:|:--------------------:|
| Prepare (validate + backup) | yes | yes | yes |
| Install (tool installers) | yes | — | — |
| Deploy (dotfiles) | yes | — | yes |
| Apply (MCP config injection) | yes | yes | — |

## Release Process

```bash
git tag v0.2.0
git push origin v0.2.0
# GitHub Actions runs goreleaser → builds linux/darwin x amd64/arm64 → publishes to Releases
```

Install script: `curl -sSL https://raw.githubusercontent.com/MiguelAguiarDEV/op-setup/main/install.sh | sh`

## Key Files Quick Reference

| What | Where |
|------|-------|
| MCP server catalog | `internal/component/catalog.go` |
| Domain types (IDs, profiles) | `internal/model/types.go` |
| Adapter interface | `internal/adapter/interface.go` |
| Adapter registry + factory | `internal/adapter/registry.go`, `factory.go` |
| Installer interface | `internal/installer/installer.go` |
| Installer registry + factory | `internal/installer/registry.go`, `factory.go` |
| Pipeline orchestrator | `internal/pipeline/orchestrator.go` |
| Pipeline planner | `internal/pipeline/planner.go` |
| TUI model | `internal/tui/model.go` |
| CLI config + flag validation | `internal/app/config.go` |
| Headless runner | `internal/app/headless.go` |
| XDG helper | `internal/xdg/xdg.go` |
| Env var resolver | `internal/component/resolver.go` |
| JSON merger | `internal/config/json_merger.go` |
| TOML merger | `internal/config/toml_merger.go` |
| Embedded dotfiles | `internal/dotfiles/embed/` |
| goreleaser config | `.goreleaser.yaml` |
| CI workflow | `.github/workflows/ci.yml` |
| Release workflow | `.github/workflows/release.yml` |
