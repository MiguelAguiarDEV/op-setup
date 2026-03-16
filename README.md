# op-setup

[![CI](https://github.com/MiguelAguiarDEV/op-setup/actions/workflows/ci.yml/badge.svg)](https://github.com/MiguelAguiarDEV/op-setup/actions/workflows/ci.yml)
[![Release](https://github.com/MiguelAguiarDEV/op-setup/actions/workflows/release.yml/badge.svg)](https://github.com/MiguelAguiarDEV/op-setup/releases)

TUI installer that sets up complete AI coding environments — install tools, deploy dotfiles, and configure [MCP](https://modelcontextprotocol.io/) servers.

```
                           _
  ___  _ __    ___  ___| |_ _   _ _ __
 / _ \| '_ \  / __|/ _ \ __| | | | '_ \
| (_) | |_) | \__ \  __/ |_| |_| | |_) |
 \___/| .__/  |___/\___|\__|\__,_| .__/
      |_|                         |_|
```

## Quick Start

```bash
# Install
curl -sSL https://raw.githubusercontent.com/MiguelAguiarDEV/op-setup/main/install.sh | sh

# Run (interactive TUI)
op-setup

# Or preview what it would do first
op-setup --dry-run

# Or run headless (CI/scripts)
op-setup --no-interactive --profile full
```

## What it does

op-setup is a single Go binary that sets up your AI coding environment from scratch. Choose a setup profile and it handles the rest — idempotently, with backup and rollback.

### Setup Profiles

| Profile | What it does |
|---------|-------------|
| **Full Setup** | Install tools + deploy dotfiles + configure MCP servers |
| **MCP Servers Only** | Only configure MCP servers in AI tool config files |
| **Dotfiles Only** | Only deploy agents, skills, scripts, and nvim config |

### Supported AI tools

| Tool | Config file | Format |
|------|------------|--------|
| Claude Code | `~/.claude/settings.json` | JSON (`mcpServers`) |
| OpenCode | `~/.config/opencode/opencode.json` | JSON (`mcp`) |
| Codex | `~/.codex/config.toml` | TOML (`[mcp_servers.X]`) |
| Gemini CLI | `~/.gemini/settings.json` | JSON (`mcpServers`) |

### Available MCP servers

| Component | Type | Description |
|-----------|------|-------------|
| Engram | Local | Persistent cross-session memory |
| Context Mode | Local | Context optimization (~98% usage reduction) |
| Playwright | Local | Browser automation via MCP |
| GitHub MCP | Remote | GitHub API access via Copilot MCP |
| Context7 | Remote | Up-to-date framework documentation |

### Tool installers (Full Setup profile)

| Tool | Install method |
|------|---------------|
| OpenCode | `npm install -g opencode` |
| Engram | `go install github.com/Gentleman-Programming/engram/cmd/engram@latest` |
| Context Mode | `npm install -g context-mode` |
| Playwright | `npx playwright install --with-deps chromium` |

### Embedded dotfiles (Full Setup / Dotfiles Only profiles)

Deploys OpenCode agents, skills, scripts, plugins, and Neovim config to `~/.config/`.

## Install

### One-liner (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/MiguelAguiarDEV/op-setup/main/install.sh | sh
```

Detects OS/arch automatically. Installs to `/usr/local/bin/`. Override with `INSTALL_DIR`:

```bash
curl -sSL https://raw.githubusercontent.com/MiguelAguiarDEV/op-setup/main/install.sh | INSTALL_DIR=~/.local/bin sh
```

### From GitHub Releases

Download the binary for your platform from [Releases](https://github.com/MiguelAguiarDEV/op-setup/releases).

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

### Interactive (TUI)

```bash
op-setup                          # Full TUI flow
op-setup --dry-run                # Preview without changes
op-setup --profile mcp-only       # Pre-select profile
```

### Non-interactive (headless)

```bash
op-setup --no-interactive --profile full --dry-run   # Preview full setup
op-setup --no-interactive --profile full              # Execute full setup
op-setup --no-interactive --profile mcp-only          # Only configure MCP servers
op-setup --no-interactive --profile dotfiles-only     # Only deploy dotfiles
```

### CLI flags

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would happen without executing |
| `--profile` | Setup profile: `full`, `mcp-only`, `dotfiles-only` |
| `--no-interactive` | Run headless without TUI (requires `--profile`) |

### TUI flow

The TUI guides you through:

1. **Profile selection** — choose Full Setup, MCP Only, or Dotfiles Only
2. **Detection** — scans for installed AI tools (skipped for Dotfiles Only)
3. **Agent selection** — pick which tools to configure (skipped for Dotfiles Only)
4. **Component selection** — pick which MCP servers to install (skipped for Dotfiles Only)
5. **Review** — confirm selections before applying
6. **Installation** — executes the 4-stage pipeline with live progress

### Idempotent

Running op-setup multiple times is safe. It merges new entries without overwriting existing config keys. If an MCP server entry already exists with identical config, it's skipped. Dotfiles are backed up before overwriting.

### Backups

Before modifying any config file or dotfile, op-setup creates a timestamped backup snapshot in `~/.op-setup/backups/`. Each snapshot includes:

- Copies of all original config files
- A `manifest.json` describing what was backed up
- Original file permissions preserved

If any stage fails mid-way, the pipeline automatically rolls back all changes using the backup.

## Architecture

```
cmd/op-setup/          Entry point
internal/
  model/               Core domain types (AgentID, ComponentID, InstallerID, SetupProfile)
  adapter/             Adapter interface + registry
    claude/            Claude Code adapter
    opencode/          OpenCode adapter (+ PostInject for plugin array)
    codex/             Codex adapter
    gemini/            Gemini CLI adapter
  component/           MCP server catalog + resolver
  installer/           Tool installers (detect/install/rollback)
  dotfiles/            Embedded dotfiles (go:embed) + deployer
    embed/             18 embedded config files
  config/              Atomic file writes, JSON merger, TOML merger
  backup/              Timestamped snapshots + restore from manifest
  pipeline/            Prepare→Install→Deploy→Apply orchestration with rollback
    steps/             BackupStep, ValidateStep, InjectStep
  tui/                 Bubbletea TUI
    screens/           8 screen renderers (Kanagawa theme)
    styles/            Shared lipgloss styles
  app/                 Wires registry + installer registry + TUI
```

### Pipeline stages

| Stage | Description |
|-------|-------------|
| **Prepare** | Validate dependencies + backup existing configs |
| **Install** | Install tools (OpenCode, Engram, Context Mode, Playwright) |
| **Deploy** | Deploy embedded dotfiles to `~/.config/` |
| **Apply** | Inject MCP server configs into AI tool config files |

Each stage is optional — the active profile determines which stages run. If any stage fails, all succeeded steps across all stages are rolled back.

### Design principles

- **Adapter pattern** — adding a new AI tool requires one sub-package + registry entry
- **DI via struct fields** — all external dependencies injectable for testing
- **Idempotent merges** — JSON and TOML mergers check before writing
- **Atomic writes** — temp file + rename prevents partial writes on crash
- **Backup before modify** — every config/dotfile change is reversible
- **Pipeline with rollback** — if any step fails, all previous steps are undone
- **Profile-aware routing** — TUI flow adapts to selected profile

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
- Target ≥85% coverage per package

## License

MIT
