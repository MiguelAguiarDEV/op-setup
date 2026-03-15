# op-setup

TUI installer that configures [MCP](https://modelcontextprotocol.io/) servers for AI coding tools.

```
                           _
  ___  _ __    ___  ___| |_ _   _ _ __
 / _ \| '_ \  / __|/ _ \ __| | | | '_ \
| (_) | |_) | \__ \  __/ |_| |_| | |_) |
 \___/| .__/  |___/\___|\__|\__,_| .__/
      |_|                         |_|
```

## What it does

op-setup detects which AI coding tools you have installed, lets you pick which MCP servers to configure, then writes the correct config entries into each tool's config file — idempotently, with backup and rollback.

**Supported AI tools:**

| Tool | Config file | Format |
|------|------------|--------|
| Claude Code | `~/.claude/settings.json` | JSON (`mcpServers`) |
| OpenCode | `~/.config/opencode/opencode.json` | JSON (`mcp`) |
| Codex | `~/.codex/config.toml` | TOML (`[mcp_servers.X]`) |
| Gemini CLI | `~/.gemini/settings.json` | JSON (`mcpServers`) |

**Available MCP servers:**

| Component | Type | Description |
|-----------|------|-------------|
| Engram | Local | Persistent cross-session memory |
| Context Mode | Local | Context optimization (~98% usage reduction) |
| Playwright | Local | Browser automation via MCP |
| GitHub MCP | Remote | GitHub API access via Copilot MCP |
| Context7 | Remote | Up-to-date framework documentation |

## Install

### From source

```bash
go install github.com/MiguelAguiarDEV/op-setup/cmd/op-setup@latest
```

### Build locally

```bash
git clone https://github.com/MiguelAguiarDEV/op-setup.git
cd op-setup
make build
./op-setup
```

## Usage

```bash
op-setup
```

The TUI guides you through:

1. **Detection** — scans for installed AI tools
2. **Agent selection** — pick which tools to configure
3. **Component selection** — pick which MCP servers to install
4. **Review** — confirm selections before applying
5. **Installation** — writes configs with backup + rollback on failure

### Idempotent

Running op-setup multiple times is safe. It merges new entries without overwriting existing config keys. If an MCP server entry already exists with identical config, it's skipped.

### Backups

Before modifying any config file, op-setup creates a timestamped backup snapshot in `~/.op-setup/backups/`. Each snapshot includes:

- Copies of all original config files
- A `manifest.json` describing what was backed up
- Original file permissions preserved

If installation fails mid-way, the pipeline automatically rolls back all changes using the backup.

## Prerequisites

MCP servers must be installed separately. op-setup only writes the configuration — it does not install the servers themselves.

| Component | Install |
|-----------|---------|
| Engram | `go install github.com/nicholasgasior/engram@latest` |
| Context Mode | `npm install -g context-mode` |
| Playwright | Auto-installed via `npx` (no pre-install needed) |
| GitHub MCP | Set `GITHUB_MCP_PAT` environment variable |
| Context7 | No install needed (remote server) |

## Architecture

```
cmd/op-setup/          Entry point
internal/
  model/               Core domain types (AgentID, ComponentID, MCPServerConfig)
  adapter/             Adapter interface + registry
    claude/            Claude Code adapter
    opencode/          OpenCode adapter (+ PostInject for plugin array)
    codex/             Codex adapter
    gemini/            Gemini CLI adapter
  component/           MCP server catalog + resolver
  config/              Atomic file writes, JSON merger, TOML merger
  backup/              Timestamped snapshots + restore from manifest
  pipeline/            Prepare→Apply orchestration with rollback
    steps/             BackupStep, ValidateStep, InjectStep
  tui/                 Bubbletea TUI
    screens/           7 screen renderers (Kanagawa theme)
    styles/            Shared lipgloss styles
  app/                 Wires registry + TUI
```

**Design principles:**

- **Adapter pattern** — adding a new AI tool requires one sub-package + registry entry
- **DI via struct fields** — all external dependencies injectable for testing
- **Idempotent merges** — JSON and TOML mergers check before writing
- **Atomic writes** — temp file + rename prevents partial writes on crash
- **Backup before modify** — every config change is reversible
- **Pipeline with rollback** — if any step fails, all previous steps are undone

## Development

```bash
make test          # Run all tests with -race
make test-cover    # Generate coverage report
make lint          # go vet
make build         # Build binary
make clean         # Remove artifacts
```

### Test approach

- stdlib `testing` only (no testify)
- Table-driven tests with `t.Run`
- `t.TempDir()` for filesystem tests (no mocks)
- DI via struct fields for dependency injection
- Race detector enabled by default

## License

MIT
