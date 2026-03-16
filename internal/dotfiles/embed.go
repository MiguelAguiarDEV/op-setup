// Package dotfiles embeds configuration files and deploys them to the user's machine.
package dotfiles

import "embed"

// EmbeddedFS contains all dotfiles embedded at compile time.
// The directory structure mirrors the deployment targets:
//
//	embed/opencode/  → $XDG_CONFIG_HOME/opencode/ (default: ~/.config/opencode/)
//	embed/nvim/      → $XDG_CONFIG_HOME/nvim/ (default: ~/.config/nvim/)
//
//go:embed all:embed
var EmbeddedFS embed.FS
