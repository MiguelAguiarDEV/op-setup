package pipeline

import (
	"errors"
	"testing"
)

func TestOrchestrator_AllSucceed(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Apply:   []Step{&fakeStep{id: "apply1"}, &fakeStep{id: "apply2"}},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Prepare.Success {
		t.Fatal("prepare should succeed")
	}
	if !result.Apply.Success {
		t.Fatal("apply should succeed")
	}
	if result.Rollback != nil {
		t.Fatal("rollback should be nil when all succeed")
	}
}

func TestOrchestrator_PrepareFails_NoApply(t *testing.T) {
	applyStep := &fakeStep{id: "apply1"}
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1", err: errors.New("prep fail")}},
		Apply:   []Step{applyStep},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Prepare.Success {
		t.Fatal("prepare should fail")
	}
	if applyStep.ran {
		t.Fatal("apply should not run when prepare fails")
	}
	if len(result.Apply.Steps) != 0 {
		t.Fatal("apply results should be empty")
	}
}

func TestOrchestrator_ApplyFails_RollbackExecuted(t *testing.T) {
	s1 := &fakeRollbackStep{fakeStep: fakeStep{id: "apply1"}}
	s2 := &fakeRollbackStep{fakeStep: fakeStep{id: "apply2", err: errors.New("fail")}}

	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Apply:   []Step{s1, s2},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Rollback == nil {
		t.Fatal("rollback should have been executed")
	}
	if !s1.rolledBack {
		t.Fatal("s1 should have been rolled back")
	}
	if s2.rolledBack {
		t.Fatal("s2 should NOT have been rolled back (it failed)")
	}
}

func TestOrchestrator_ApplyFails_FirstStep_NoRollback(t *testing.T) {
	s1 := &fakeRollbackStep{fakeStep: fakeStep{id: "apply1", err: errors.New("fail")}}

	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Apply:   []Step{s1},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Rollback == nil {
		t.Fatal("rollback result should exist")
	}
	// No steps succeeded, so rollback should have 0 steps.
	if len(result.Rollback.Steps) != 0 {
		t.Fatalf("expected 0 rollback steps, got %d", len(result.Rollback.Steps))
	}
}

func TestOrchestrator_ProgressEvents(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Apply:   []Step{&fakeStep{id: "apply1"}},
	}

	var events []ProgressEvent
	progress := func(e ProgressEvent) {
		events = append(events, e)
	}

	o := NewOrchestrator(progress)
	o.Execute(plan)

	// 2 events per step (running + succeeded) × 2 steps = 4 events
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
}

func TestOrchestrator_EmptyPlan(t *testing.T) {
	plan := StagePlan{}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}
