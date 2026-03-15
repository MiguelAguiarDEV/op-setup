// Package dotfiles embeds configuration files and deploys them to the user's machine.
package dotfiles

import "embed"

// EmbeddedFS contains all dotfiles embedded at compile time.
// The directory structure mirrors the deployment targets:
//
//	embed/opencode/  → ~/.config/opencode/
//	embed/nvim/      → ~/.config/nvim/
//
//go:embed all:embed
var EmbeddedFS embed.FS
