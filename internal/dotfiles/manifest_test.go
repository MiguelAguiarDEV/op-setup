package dotfiles

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildManifest(t *testing.T) {
	mappings, err := BuildManifest(EmbeddedFS, "/home/test/.config")
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}

	if len(mappings) < 17 {
		t.Errorf("BuildManifest() returned %d mappings, want >= 17", len(mappings))
	}

	// Verify a few specific mappings.
	tests := []struct {
		embedSuffix  string
		targetSuffix string
	}{
		{"opencode/AGENTS.md", "opencode/AGENTS.md"},
		{"opencode/agents/planner.md", "opencode/agents/planner.md"},
		{"opencode/skills/op-guardrails/SKILL.md", "opencode/skills/op-guardrails/SKILL.md"},
		{"nvim/init.lua", "nvim/init.lua"},
		{"nvim/lua/plugins/opencode.lua", "nvim/lua/plugins/opencode.lua"},
	}

	for _, tt := range tests {
		t.Run(tt.embedSuffix, func(t *testing.T) {
			found := false
			for _, m := range mappings {
				if strings.HasSuffix(m.EmbedPath, tt.embedSuffix) {
					found = true
					wantTarget := filepath.Join("/home/test/.config", tt.targetSuffix)
					if m.TargetPath != wantTarget {
						t.Errorf("TargetPath = %q, want %q", m.TargetPath, wantTarget)
					}
					break
				}
			}
			if !found {
				t.Errorf("mapping with embed suffix %q not found", tt.embedSuffix)
			}
		})
	}
}

func TestBuildManifest_AllTargetsAbsolute(t *testing.T) {
	mappings, err := BuildManifest(EmbeddedFS, "/home/test/.config")
	if err != nil {
		t.Fatalf("BuildManifest() error = %v", err)
	}

	for _, m := range mappings {
		if !filepath.IsAbs(m.TargetPath) {
			t.Errorf("TargetPath %q is not absolute", m.TargetPath)
		}
	}
}
