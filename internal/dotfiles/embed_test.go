package dotfiles

import (
	"io/fs"
	"testing"
)

func TestEmbeddedFS_ContainsExpectedFiles(t *testing.T) {
	expectedFiles := []string{
		"embed/opencode/.gitignore",
		"embed/opencode/AGENTS.md",
		"embed/opencode/agents/ACTIVE_AGENTS.txt",
		"embed/opencode/agents/planner.md",
		"embed/opencode/agents/code-reviewer.md",
		"embed/opencode/agents/security-reviewer.md",
		"embed/opencode/agents/release-manager.md",
		"embed/opencode/agents/baymax.md",
		"embed/opencode/skills/ACTIVE_SKILLS.txt",
		"embed/opencode/skills/op-guardrails/SKILL.md",
		"embed/opencode/skills/op-skill-creator/SKILL.md",
		"embed/opencode/skills/qa-debugging/SKILL.md",
		"embed/opencode/scripts/generate-skills-table.py",
		"embed/opencode/scripts/validate-skill-tags.py",
		"embed/opencode/plugins/engram.ts",
		"embed/opencode/package.json",
		"embed/nvim/init.lua",
		"embed/nvim/lua/plugins/opencode.lua",
	}

	for _, path := range expectedFiles {
		t.Run(path, func(t *testing.T) {
			data, err := fs.ReadFile(EmbeddedFS, path)
			if err != nil {
				t.Fatalf("ReadFile(%q) error = %v", path, err)
			}
			if len(data) == 0 {
				t.Errorf("ReadFile(%q) returned empty content", path)
			}
		})
	}
}

func TestEmbeddedFS_FileCount(t *testing.T) {
	var count int
	err := fs.WalkDir(EmbeddedFS, "embed", func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir error = %v", err)
	}

	// With "all:embed" directive, dotfiles like .gitignore are included.
	if count != 18 {
		t.Errorf("embedded file count = %d, want 18", count)
	}
}
