package steps

import (
	"fmt"

	"github.com/MiguelAguiarDEV/op-setup/internal/adapter"
	"github.com/MiguelAguiarDEV/op-setup/internal/backup"
	"github.com/MiguelAguiarDEV/op-setup/internal/component"
	"github.com/MiguelAguiarDEV/op-setup/internal/config"
	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// InjectStep merges MCP server configs into one agent's config file.
type InjectStep struct {
	adapter    adapter.Adapter
	components []component.Component
	homeDir    string
	resolver   *component.Resolver
	manifest   *backup.Manifest // set externally for rollback
	changed    bool
}

// NewInjectStep creates an InjectStep.
func NewInjectStep(
	a adapter.Adapter,
	components []component.Component,
	homeDir string,
	resolver *component.Resolver,
) *InjectStep {
	return &InjectStep{
		adapter:    a,
		components: components,
		homeDir:    homeDir,
		resolver:   resolver,
	}
}

// SetManifest sets the backup manifest for rollback support.
func (s *InjectStep) SetManifest(m *backup.Manifest) {
	s.manifest = m
}

func (s *InjectStep) ID() string {
	return fmt.Sprintf("inject-%s", s.adapter.Agent())
}

// Run merges MCP configs into the agent's config file, then runs PostInject.
func (s *InjectStep) Run() error {
	// Build the servers map.
	servers := make(map[string]model.MCPServerConfig, len(s.components))
	var componentIDs []model.ComponentID

	for _, comp := range s.components {
		cfg := s.resolver.Resolve(comp, s.adapter.Agent())
		key := s.resolver.ConfigKey(comp)
		servers[key] = cfg
		componentIDs = append(componentIDs, comp.ID)
	}

	cfgPath := s.adapter.ConfigPath(s.homeDir)

	// Merge based on strategy.
	var changed bool
	var err error

	switch s.adapter.MCPStrategy() {
	case model.StrategyMergeIntoJSON:
		merger := config.NewJSONMerger()
		changed, err = merger.Merge(cfgPath, s.adapter.MCPConfigKey(), servers)
	case model.StrategyMergeIntoTOML:
		merger := config.NewTOMLMerger()
		changed, err = merger.Merge(cfgPath, servers)
	default:
		return fmt.Errorf("unsupported MCP strategy: %d", s.adapter.MCPStrategy())
	}

	if err != nil {
		return fmt.Errorf("merge config for %s: %w", s.adapter.Name(), err)
	}

	s.changed = changed

	// Run post-inject actions (e.g., OpenCode plugin entry).
	if err := s.adapter.PostInject(s.homeDir, componentIDs); err != nil {
		return fmt.Errorf("post-inject for %s: %w", s.adapter.Name(), err)
	}

	return nil
}

// Rollback restores the original config from the backup manifest.
func (s *InjectStep) Rollback() error {
	if s.manifest == nil {
		return fmt.Errorf("no backup manifest available for rollback")
	}
	rs := backup.NewRestoreService()
	return rs.Restore(*s.manifest)
}

// Changed returns true if the config was actually modified.
func (s *InjectStep) Changed() bool {
	return s.changed
}
