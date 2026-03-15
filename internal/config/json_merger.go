package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"reflect"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// JSONMerger handles read-modify-write for JSON config files.
type JSONMerger struct {
	// ReadFile reads a file's contents. Defaults to os.ReadFile.
	ReadFile func(string) ([]byte, error)

	// WriteFile writes data to a file atomically. Defaults to WriteFileAtomic.
	WriteFile func(string, []byte, os.FileMode) error
}

// NewJSONMerger creates a JSONMerger with default dependencies.
func NewJSONMerger() *JSONMerger {
	return &JSONMerger{
		ReadFile:  os.ReadFile,
		WriteFile: WriteFileAtomic,
	}
}

// Merge reads the JSON file at path, merges MCP server entries under the given key,
// and writes back. Returns (changed bool, err error).
//
// Idempotent: if all entries already exist with identical config, changed=false
// and no write occurs.
//
// Creates the file and parent directories if they don't exist.
// Preserves all existing keys in the file.
func (m *JSONMerger) Merge(path string, key string, servers map[string]model.MCPServerConfig) (bool, error) {
	var cfg map[string]any

	data, err := m.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return false, err
		}
		// File doesn't exist — start with empty config.
		cfg = make(map[string]any)
	} else {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return false, fmt.Errorf("config file is not valid: %s (%w)", path, err)
		}
	}

	// Get or create the MCP servers map.
	var mcpMap map[string]any
	if raw, ok := cfg[key]; ok {
		if m, ok := raw.(map[string]any); ok {
			mcpMap = m
		} else {
			mcpMap = make(map[string]any)
		}
	} else {
		mcpMap = make(map[string]any)
	}

	// Check if any changes are needed.
	changed := false
	for name, serverCfg := range servers {
		// Convert the MCPServerConfig to a generic map for comparison.
		newEntry := serverConfigToMap(serverCfg)

		if existing, ok := mcpMap[name]; ok {
			if existingMap, ok := existing.(map[string]any); ok {
				if mapsEqual(existingMap, newEntry) {
					continue // Already identical, skip.
				}
			}
		}

		mcpMap[name] = newEntry
		changed = true
	}

	if !changed {
		return false, nil
	}

	cfg[key] = mcpMap

	out, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return false, fmt.Errorf("marshal config: %w", err)
	}
	out = append(out, '\n')

	if err := m.WriteFile(path, out, 0o644); err != nil {
		return false, err
	}

	return true, nil
}

// serverConfigToMap converts an MCPServerConfig to a map[string]any
// matching the JSON representation.
func serverConfigToMap(cfg model.MCPServerConfig) map[string]any {
	m := map[string]any{
		"type":    string(cfg.Type),
		"enabled": cfg.Enabled,
	}

	if len(cfg.Command) > 0 {
		// Convert []string to []any for JSON compatibility.
		cmd := make([]any, len(cfg.Command))
		for i, c := range cfg.Command {
			cmd[i] = c
		}
		m["command"] = cmd
	}

	if cfg.URL != "" {
		m["url"] = cfg.URL
	}

	if len(cfg.Headers) > 0 {
		headers := make(map[string]any, len(cfg.Headers))
		for k, v := range cfg.Headers {
			headers[k] = v
		}
		m["headers"] = headers
	}

	return m
}

// mapsEqual performs a deep comparison of two map[string]any values.
func mapsEqual(a, b map[string]any) bool {
	return reflect.DeepEqual(a, b)
}
