package adapter

import (
	"sort"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// Registry holds all registered adapters indexed by AgentID.
type Registry struct {
	adapters map[model.AgentID]Adapter
}

// NewRegistry creates a Registry pre-populated with the given adapters.
// Returns an error if any adapter has a duplicate AgentID.
func NewRegistry(adapters ...Adapter) (*Registry, error) {
	r := &Registry{
		adapters: make(map[model.AgentID]Adapter, len(adapters)),
	}
	for _, a := range adapters {
		if err := r.Register(a); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// Register adds an adapter to the registry.
// Returns DuplicateAdapterError if the AgentID is already registered.
func (r *Registry) Register(a Adapter) error {
	id := a.Agent()
	if _, exists := r.adapters[id]; exists {
		return &DuplicateAdapterError{Agent: id}
	}
	r.adapters[id] = a
	return nil
}

// Get returns the adapter for the given AgentID, or false if not found.
func (r *Registry) Get(agent model.AgentID) (Adapter, bool) {
	a, ok := r.adapters[agent]
	return a, ok
}

// All returns every registered adapter sorted by AgentID (alphabetical).
func (r *Registry) All() []Adapter {
	result := make([]Adapter, 0, len(r.adapters))
	for _, a := range r.adapters {
		result = append(result, a)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Agent() < result[j].Agent()
	})
	return result
}

// Len returns the number of registered adapters.
func (r *Registry) Len() int {
	return len(r.adapters)
}
