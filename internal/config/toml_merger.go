package config

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// TOMLMerger handles read-modify-write for TOML config files (Codex).
type TOMLMerger struct {
	ReadFile  func(string) ([]byte, error)
	WriteFile func(string, []byte, os.FileMode) error
}

// NewTOMLMerger creates a TOMLMerger with default dependencies.
func NewTOMLMerger() *TOMLMerger {
	return &TOMLMerger{
		ReadFile:  os.ReadFile,
		WriteFile: WriteFileAtomic,
	}
}

// Merge reads the TOML file at path, merges MCP server entries under
// [mcp_servers.X] table sections, and writes back.
// Returns (changed bool, err error).
//
// Idempotent: if all entries already exist with identical config, changed=false.
// Creates the file and parent directories if they don't exist.
//
// Uses a simple line-based approach to preserve existing TOML content
// while appending new [mcp_servers.X] sections.
func (m *TOMLMerger) Merge(path string, servers map[string]model.MCPServerConfig) (bool, error) {
	var existingContent []byte

	data, err := m.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return false, err
		}
		existingContent = nil
	} else {
		existingContent = data
	}

	// Parse existing sections to check for idempotency.
	existingSections := parseTOMLSections(existingContent)

	// Build new sections for servers that don't already exist or differ.
	changed := false
	var newSections []string

	// Sort server names for deterministic output.
	names := make([]string, 0, len(servers))
	for name := range servers {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		cfg := servers[name]
		sectionHeader := fmt.Sprintf("[mcp_servers.%s]", name)
		newSection := renderTOMLSection(sectionHeader, cfg)

		if existing, ok := existingSections[sectionHeader]; ok {
			if normalizeTOML(existing) == normalizeTOML(newSection) {
				continue // Already identical.
			}
			// Section exists but differs — we need to replace it.
			// For simplicity, we'll rebuild the entire file.
			changed = true
		} else {
			changed = true
		}
		newSections = append(newSections, newSection)
	}

	if !changed {
		return false, nil
	}

	// Rebuild: keep existing content, replace or append sections.
	var result bytes.Buffer

	if len(existingContent) > 0 {
		// Remove existing mcp_servers sections that we're replacing.
		cleaned := removeTOMLSections(string(existingContent), names)
		cleaned = strings.TrimRight(cleaned, "\n")
		if cleaned != "" {
			result.WriteString(cleaned)
			result.WriteString("\n")
		}
	}

	for _, section := range newSections {
		result.WriteString("\n")
		result.WriteString(section)
		result.WriteString("\n")
	}

	if err := m.WriteFile(path, result.Bytes(), 0o644); err != nil {
		return false, err
	}

	return true, nil
}

// renderTOMLSection renders a single [mcp_servers.X] section.
func renderTOMLSection(header string, cfg model.MCPServerConfig) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf("type = %q\n", string(cfg.Type)))
	b.WriteString(fmt.Sprintf("enabled = %t\n", cfg.Enabled))

	if len(cfg.Command) > 0 {
		parts := make([]string, len(cfg.Command))
		for i, c := range cfg.Command {
			parts[i] = fmt.Sprintf("%q", c)
		}
		b.WriteString(fmt.Sprintf("command = [%s]\n", strings.Join(parts, ", ")))
	}

	if cfg.URL != "" {
		b.WriteString(fmt.Sprintf("url = %q\n", cfg.URL))
	}

	if len(cfg.Headers) > 0 {
		// Sort header keys for deterministic output.
		keys := make([]string, 0, len(cfg.Headers))
		for k := range cfg.Headers {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("headers.%s = %q\n", k, cfg.Headers[k]))
		}
	}

	return b.String()
}

// parseTOMLSections extracts [mcp_servers.X] sections from TOML content.
func parseTOMLSections(data []byte) map[string]string {
	sections := make(map[string]string)
	if len(data) == 0 {
		return sections
	}

	lines := strings.Split(string(data), "\n")
	var currentHeader string
	var currentBody strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[mcp_servers.") {
			// Save previous section if any.
			if currentHeader != "" {
				sections[currentHeader] = currentHeader + "\n" + currentBody.String()
			}
			currentHeader = trimmed
			currentBody.Reset()
		} else if strings.HasPrefix(trimmed, "[") && currentHeader != "" {
			// New non-mcp_servers section — save current and reset.
			sections[currentHeader] = currentHeader + "\n" + currentBody.String()
			currentHeader = ""
			currentBody.Reset()
		} else if currentHeader != "" {
			if trimmed != "" {
				currentBody.WriteString(line)
				currentBody.WriteString("\n")
			}
		}
	}

	// Save last section.
	if currentHeader != "" {
		sections[currentHeader] = currentHeader + "\n" + currentBody.String()
	}

	return sections
}

// removeTOMLSections removes [mcp_servers.X] sections for the given names.
func removeTOMLSections(content string, names []string) string {
	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[fmt.Sprintf("[mcp_servers.%s]", n)] = true
	}

	lines := strings.Split(content, "\n")
	var result []string
	skip := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "[mcp_servers.") {
			if nameSet[trimmed] {
				skip = true
				continue
			}
			skip = false
		} else if strings.HasPrefix(trimmed, "[") {
			skip = false
		}

		if !skip {
			result = append(result, line)
		}
	}

	return strings.Join(result, "\n")
}

// normalizeTOML normalizes whitespace for comparison.
func normalizeTOML(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var normalized []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}
	return strings.Join(normalized, "\n")
}
