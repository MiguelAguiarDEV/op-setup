package pipeline

import (
	"errors"
	"testing"
)

func TestExecuteRollback_ReverseOrder(t *testing.T) {
	s1 := &fakeRollbackStep{fakeStep: fakeStep{id: "s1"}}
	s2 := &fakeRollbackStep{fakeStep: fakeStep{id: "s2"}}
	s3 := &fakeRollbackStep{fakeStep: fakeStep{id: "s3"}}

	var order []string
	progress := func(e ProgressEvent) {
		if e.Status == StatusRolledBack {
			order = append(order, e.StepID)
		}
	}

	result := ExecuteRollback([]Step{s1, s2, s3}, progress)

	if !result.Success {
		t.Fatal("expected success")
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 rollbacks, got %d", len(order))
	}
	// Should be reversed: s3, s2, s1
	if order[0] != "s3" || order[1] != "s2" || order[2] != "s1" {
		t.Fatalf("rollback order = %v, want [s3 s2 s1]", order)
	}
}

func TestExecuteRollback_SkipsNonRollbackSteps(t *testing.T) {
	s1 := &fakeStep{id: "s1"} // Not a RollbackStep
	s2 := &fakeRollbackStep{fakeStep: fakeStep{id: "s2"}}

	result := ExecuteRollback([]Step{s1, s2}, nil)

	if !result.Success {
		t.Fatal("expected success")
	}
	if len(result.Steps) != 1 {
		t.Fatalf("expected 1 rollback result, got %d", len(result.Steps))
	}
	if result.Steps[0].StepID != "s2" {
		t.Fatalf("rolled back step = %q, want %q", result.Steps[0].StepID, "s2")
	}
}

func TestExecuteRollback_RollbackError(t *testing.T) {
	s1 := &fakeRollbackStep{
		fakeStep:    fakeStep{id: "s1"},
		rollbackErr: errors.New("rollback failed"),
	}

	result := ExecuteRollback([]Step{s1}, nil)

	if result.Success {
		t.Fatal("expected failure when rollback errors")
	}
	if result.Steps[0].Status != StatusFailed {
		t.Fatalf("status = %q, want %q", result.Steps[0].Status, StatusFailed)
	}
}

func TestExecuteRollback_Empty(t *testing.T) {
	result := ExecuteRollback([]Step{}, nil)

	if !result.Success {
		t.Fatal("empty rollback should succeed")
	}
	if len(result.Steps) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Steps))
	}
}
