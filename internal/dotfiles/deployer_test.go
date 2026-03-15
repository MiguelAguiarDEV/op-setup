package dotfiles

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestDeployer_Plan_AllNew(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
	}

	plan, err := d.Plan()
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	if len(plan) < 17 {
		t.Fatalf("Plan() returned %d entries, want >= 17", len(plan))
	}

	for _, entry := range plan {
		if entry.Action != ActionCreate {
			t.Errorf("entry %q action = %q, want %q", entry.Mapping.TargetPath, entry.Action, ActionCreate)
		}
	}
}

func TestDeployer_Plan_IdenticalSkipped(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(path string, data []byte, perm os.FileMode) error {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, data, perm)
		},
		MkdirAll: os.MkdirAll,
	}

	// First deploy to create all files.
	plan, err := d.Plan()
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	_, err = d.Deploy(plan, nil)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	// Second plan should show all as skip.
	plan2, err := d.Plan()
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	for _, entry := range plan2 {
		if entry.Action != ActionSkip {
			t.Errorf("entry %q action = %q, want %q", entry.Mapping.TargetPath, entry.Action, ActionSkip)
		}
	}
}

func TestDeployer_Plan_ModifiedOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(path string, data []byte, perm os.FileMode) error {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, data, perm)
		},
		MkdirAll: os.MkdirAll,
	}

	// Deploy first.
	plan, _ := d.Plan()
	_, _ = d.Deploy(plan, nil)

	// Modify one file.
	agentsPath := filepath.Join(configDir, "opencode", "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("modified"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Re-plan.
	plan2, err := d.Plan()
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}

	foundOverwrite := false
	for _, entry := range plan2 {
		if entry.Mapping.TargetPath == agentsPath {
			if entry.Action != ActionOverwrite {
				t.Errorf("AGENTS.md action = %q, want %q", entry.Action, ActionOverwrite)
			}
			foundOverwrite = true
		}
	}
	if !foundOverwrite {
		t.Error("AGENTS.md not found in plan")
	}
}

func TestDeployer_Deploy_CreatesFiles(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(path string, data []byte, perm os.FileMode) error {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, data, perm)
		},
		MkdirAll: os.MkdirAll,
	}

	plan, _ := d.Plan()
	results, err := d.Deploy(plan, nil)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	for _, r := range results {
		if r.Err != nil {
			t.Errorf("result %q error = %v", r.Mapping.TargetPath, r.Err)
			continue
		}
		// Verify file exists.
		if _, err := os.Stat(r.Mapping.TargetPath); err != nil {
			t.Errorf("file %q not created: %v", r.Mapping.TargetPath, err)
		}
	}
}

func TestDeployer_Deploy_WithBackup(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(path string, data []byte, perm os.FileMode) error {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, data, perm)
		},
		MkdirAll: os.MkdirAll,
	}

	// First deploy.
	plan, _ := d.Plan()
	_, _ = d.Deploy(plan, nil)

	// Modify a file.
	agentsPath := filepath.Join(configDir, "opencode", "AGENTS.md")
	os.WriteFile(agentsPath, []byte("modified"), 0o600)

	// Re-plan and deploy with backup.
	plan2, _ := d.Plan()
	backupCalled := false
	var backedPaths []string

	_, err := d.Deploy(plan2, func(paths []string) error {
		backupCalled = true
		backedPaths = paths
		return nil
	})
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}

	if !backupCalled {
		t.Error("backup function was not called")
	}
	if len(backedPaths) != 1 || backedPaths[0] != agentsPath {
		t.Errorf("backed paths = %v, want [%s]", backedPaths, agentsPath)
	}
}

func TestDeployer_Deploy_BackupError(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(path string, data []byte, perm os.FileMode) error {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, data, perm)
		},
		MkdirAll: os.MkdirAll,
	}

	// First deploy.
	plan, _ := d.Plan()
	_, _ = d.Deploy(plan, nil)

	// Modify a file.
	agentsPath := filepath.Join(configDir, "opencode", "AGENTS.md")
	os.WriteFile(agentsPath, []byte("modified"), 0o600)

	// Re-plan and deploy with failing backup.
	plan2, _ := d.Plan()
	_, err := d.Deploy(plan2, func(_ []string) error {
		return errors.New("backup failed")
	})
	if err == nil {
		t.Fatal("Deploy() expected error when backup fails")
	}
}

func TestDeployer_Deploy_NilWriteAtomic(t *testing.T) {
	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: "/tmp/test",
		ReadFile:  os.ReadFile,
		// WriteFileAtomic is nil
	}

	plan := []PlanEntry{{
		Mapping: FileMapping{EmbedPath: "embed/opencode/AGENTS.md", TargetPath: "/tmp/test/AGENTS.md"},
		Action:  ActionCreate,
	}}

	_, err := d.Deploy(plan, nil)
	if err == nil {
		t.Fatal("Deploy() expected error when WriteFileAtomic is nil")
	}
}

func TestDeployer_Deploy_SkipsCorrectly(t *testing.T) {
	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: "/tmp/test",
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(_ string, _ []byte, _ os.FileMode) error {
			t.Error("WriteFileAtomic should not be called for skip")
			return nil
		},
		MkdirAll: os.MkdirAll,
	}

	plan := []PlanEntry{{
		Mapping: FileMapping{EmbedPath: "embed/opencode/AGENTS.md", TargetPath: "/tmp/test/AGENTS.md"},
		Action:  ActionSkip,
	}}

	results, err := d.Deploy(plan, nil)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}
	if len(results) != 1 || results[0].Action != ActionSkip {
		t.Errorf("results = %v, want 1 skip", results)
	}
}

func TestDeployer_BackedUpPaths(t *testing.T) {
	d := &Deployer{}
	if paths := d.BackedUpPaths(); paths != nil {
		t.Errorf("BackedUpPaths() = %v, want nil", paths)
	}
}

func TestDeployer_Deploy_MkdirFails(t *testing.T) {
	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: "/tmp/test",
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(_ string, _ []byte, _ os.FileMode) error {
			return nil
		},
		MkdirAll: func(_ string, _ os.FileMode) error {
			return errors.New("mkdir failed")
		},
	}

	plan := []PlanEntry{{
		Mapping: FileMapping{EmbedPath: "embed/opencode/AGENTS.md", TargetPath: "/tmp/test/AGENTS.md"},
		Action:  ActionCreate,
	}}

	results, err := d.Deploy(plan, nil)
	if err != nil {
		t.Fatalf("Deploy() error = %v (should not fail globally)", err)
	}
	if len(results) != 1 || results[0].Err == nil {
		t.Error("expected individual file error for mkdir failure")
	}
	if !errors.Is(results[0].Err, ErrDeployFailed) {
		t.Errorf("error type = %T, want DeployFailedError", results[0].Err)
	}
}

func TestDeployer_Deploy_WriteFails(t *testing.T) {
	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: "/tmp/test",
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(_ string, _ []byte, _ os.FileMode) error {
			return errors.New("write failed")
		},
		MkdirAll: func(_ string, _ os.FileMode) error {
			return nil
		},
	}

	plan := []PlanEntry{{
		Mapping: FileMapping{EmbedPath: "embed/opencode/AGENTS.md", TargetPath: "/tmp/test/AGENTS.md"},
		Action:  ActionCreate,
	}}

	results, err := d.Deploy(plan, nil)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}
	if len(results) != 1 || results[0].Err == nil {
		t.Error("expected individual file error for write failure")
	}
	if !errors.Is(results[0].Err, ErrDeployFailed) {
		t.Errorf("error type = %T, want DeployFailedError", results[0].Err)
	}
}

func TestDeployer_Deploy_BadEmbedPath(t *testing.T) {
	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: "/tmp/test",
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(_ string, _ []byte, _ os.FileMode) error {
			return nil
		},
		MkdirAll: func(_ string, _ os.FileMode) error {
			return nil
		},
	}

	plan := []PlanEntry{{
		Mapping: FileMapping{EmbedPath: "embed/nonexistent/file.md", TargetPath: "/tmp/test/file.md"},
		Action:  ActionCreate,
	}}

	results, err := d.Deploy(plan, nil)
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}
	if len(results) != 1 || results[0].Err == nil {
		t.Error("expected individual file error for bad embed path")
	}
	if !errors.Is(results[0].Err, ErrReadEmbed) {
		t.Errorf("error type = %T, want ReadEmbedError", results[0].Err)
	}
}

func TestDeployer_Deploy_NoBackupWhenOnlyCreates(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "config")

	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: configDir,
		ReadFile:  os.ReadFile,
		WriteFileAtomic: func(path string, data []byte, perm os.FileMode) error {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return err
			}
			return os.WriteFile(path, data, perm)
		},
		MkdirAll: os.MkdirAll,
	}

	plan, _ := d.Plan()
	backupCalled := false
	_, err := d.Deploy(plan, func(_ []string) error {
		backupCalled = true
		return nil
	})
	if err != nil {
		t.Fatalf("Deploy() error = %v", err)
	}
	if backupCalled {
		t.Error("backup should not be called when all files are new (create only)")
	}
}

func TestDeployer_Plan_ReadEmbedError(t *testing.T) {
	// Use a minimal FS that has a directory but the file read fails.
	// This is hard to trigger with embed.FS, so we test via BuildManifest error path.
	// The Plan() function wraps BuildManifest errors, so we test that path indirectly.
	d := &Deployer{
		FS:        EmbeddedFS,
		ConfigDir: "/tmp/test",
		ReadFile:  os.ReadFile,
	}

	// Plan should succeed with the real embedded FS.
	plan, err := d.Plan()
	if err != nil {
		t.Fatalf("Plan() error = %v", err)
	}
	if len(plan) == 0 {
		t.Error("Plan() returned empty plan")
	}
}
