package installer

import (
	"sort"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Registry holds all registered installers indexed by InstallerID.
type Registry struct {
	installers map[model.InstallerID]Installer
}

// NewRegistry creates a Registry pre-populated with the given installers.
// Returns an error if any installer has a duplicate ID.
func NewRegistry(installers ...Installer) (*Registry, error) {
	r := &Registry{
		installers: make(map[model.InstallerID]Installer, len(installers)),
	}
	for _, inst := range installers {
		if err := r.Register(inst); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Register adds an installer to the registry.
// Returns DuplicateInstallerError if the ID is already registered.
func (r *Registry) Register(inst Installer) error {
	id := inst.ID()
	if _, exists := r.installers[id]; exists {
		return &DuplicateInstallerError{ID: id}
	}
	r.installers[id] = inst
	return nil
}

// Get returns the installer for the given ID, or false if not found.
func (r *Registry) Get(id model.InstallerID) (Installer, bool) {
	inst, ok := r.installers[id]
	return inst, ok
}

// All returns every registered installer sorted by ID.
func (r *Registry) All() []Installer {
	result := make([]Installer, 0, len(r.installers))
	for _, inst := range r.installers {
		result = append(result, inst)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID() < result[j].ID()
	})
	return result
}

// Len returns the number of registered installers.
func (r *Registry) Len() int {
	return len(r.installers)
}
