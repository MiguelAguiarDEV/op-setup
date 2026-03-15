package installer

import (
	"context"
	"errors"
	"testing"

	"github.com/MiguelAguiarDEV/op-setup/internal/model"
)

// stubInstaller is a minimal Installer for registry tests.
type stubInstaller struct {
	id   model.InstallerID
	name string
}

func (s *stubInstaller) ID() model.InstallerID                  { return s.id }
func (s *stubInstaller) Name() string                           { return s.name }
func (s *stubInstaller) Detect(_ context.Context) (bool, error) { return false, nil }
func (s *stubInstaller) Install(_ context.Context) error        { return nil }
func (s *stubInstaller) Rollback(_ context.Context) error       { return nil }
func (s *stubInstaller) Prerequisites() []string                { return nil }

func TestRegistry_NewRegistry(t *testing.T) {
	a := &stubInstaller{id: "a", name: "A"}
	b := &stubInstaller{id: "b", name: "B"}

	r, err := NewRegistry(a, b)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}
	if r.Len() != 2 {
		t.Errorf("Len() = %d, want 2", r.Len())
	}
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
	a1 := &stubInstaller{id: "a", name: "A1"}
	a2 := &stubInstaller{id: "a", name: "A2"}

	_, err := NewRegistry(a1, a2)
	if err == nil {
		t.Fatal("NewRegistry() expected error for duplicate")
	}
	if !errors.Is(err, ErrDuplicateInstaller) {
		t.Errorf("error = %v, want ErrDuplicateInstaller", err)
	}
}

func TestRegistry_Get(t *testing.T) {
	a := &stubInstaller{id: "a", name: "A"}
	r, _ := NewRegistry(a)

	got, ok := r.Get("a")
	if !ok {
		t.Fatal("Get(a) = false, want true")
	}
	if got.ID() != "a" {
		t.Errorf("Get(a).ID() = %q, want %q", got.ID(), "a")
	}

	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Get(nonexistent) = true, want false")
	}
}

func TestRegistry_All_Sorted(t *testing.T) {
	c := &stubInstaller{id: "c", name: "C"}
	a := &stubInstaller{id: "a", name: "A"}
	b := &stubInstaller{id: "b", name: "B"}

	r, _ := NewRegistry(c, a, b)
	all := r.All()

	if len(all) != 3 {
		t.Fatalf("All() len = %d, want 3", len(all))
	}

	wantOrder := []model.InstallerID{"a", "b", "c"}
	for i, want := range wantOrder {
		if all[i].ID() != want {
			t.Errorf("All()[%d].ID() = %q, want %q", i, all[i].ID(), want)
		}
	}
}

func TestRegistry_Register_AfterCreation(t *testing.T) {
	r, _ := NewRegistry()
	if r.Len() != 0 {
		t.Errorf("Len() = %d, want 0", r.Len())
	}

	a := &stubInstaller{id: "a", name: "A"}
	if err := r.Register(a); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if r.Len() != 1 {
		t.Errorf("Len() = %d, want 1", r.Len())
	}

	// Duplicate.
	if err := r.Register(a); err == nil {
		t.Error("Register() expected error for duplicate")
	}
}
