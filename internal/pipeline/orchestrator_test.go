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

func TestOrchestrator_AllFourStages_Success(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Install: []Step{&fakeStep{id: "inst1"}, &fakeStep{id: "inst2"}},
		Deploy:  []Step{&fakeStep{id: "dep1"}},
		Apply:   []Step{&fakeStep{id: "apply1"}},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Prepare.Success {
		t.Fatal("prepare should succeed")
	}
	if !result.Install.Success {
		t.Fatal("install should succeed")
	}
	if !result.Deploy.Success {
		t.Fatal("deploy should succeed")
	}
	if !result.Apply.Success {
		t.Fatal("apply should succeed")
	}
	if result.Rollback != nil {
		t.Fatal("rollback should be nil when all succeed")
	}
}

func TestOrchestrator_InstallFails_RollbackInstallOnly(t *testing.T) {
	i1 := &fakeRollbackStep{fakeStep: fakeStep{id: "inst1"}}
	i2 := &fakeRollbackStep{fakeStep: fakeStep{id: "inst2", err: errors.New("install fail")}}
	d1 := &fakeStep{id: "dep1"}
	a1 := &fakeStep{id: "apply1"}

	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Install: []Step{i1, i2},
		Deploy:  []Step{d1},
		Apply:   []Step{a1},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Install.Success {
		t.Fatal("install should fail")
	}
	if result.Rollback == nil {
		t.Fatal("rollback should have been executed")
	}
	if !i1.rolledBack {
		t.Fatal("i1 should have been rolled back")
	}
	if i2.rolledBack {
		t.Fatal("i2 should NOT have been rolled back (it failed)")
	}
	if d1.ran {
		t.Fatal("deploy should not run when install fails")
	}
	if a1.ran {
		t.Fatal("apply should not run when install fails")
	}
}

func TestOrchestrator_DeployFails_RollbackDeployAndInstall(t *testing.T) {
	i1 := &fakeRollbackStep{fakeStep: fakeStep{id: "inst1"}}
	d1 := &fakeRollbackStep{fakeStep: fakeStep{id: "dep1", err: errors.New("deploy fail")}}
	a1 := &fakeStep{id: "apply1"}

	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Install: []Step{i1},
		Deploy:  []Step{d1},
		Apply:   []Step{a1},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Rollback == nil {
		t.Fatal("rollback should have been executed")
	}
	// i1 succeeded and should be rolled back (cross-stage rollback).
	if !i1.rolledBack {
		t.Fatal("i1 should have been rolled back (cross-stage)")
	}
	// d1 failed, should NOT be rolled back.
	if d1.rolledBack {
		t.Fatal("d1 should NOT have been rolled back (it failed)")
	}
	if a1.ran {
		t.Fatal("apply should not run when deploy fails")
	}
}

func TestOrchestrator_ApplyFails_RollbackAllThreeStages(t *testing.T) {
	i1 := &fakeRollbackStep{fakeStep: fakeStep{id: "inst1"}}
	d1 := &fakeRollbackStep{fakeStep: fakeStep{id: "dep1"}}
	a1 := &fakeRollbackStep{fakeStep: fakeStep{id: "apply1"}}
	a2 := &fakeRollbackStep{fakeStep: fakeStep{id: "apply2", err: errors.New("apply fail")}}

	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Install: []Step{i1},
		Deploy:  []Step{d1},
		Apply:   []Step{a1, a2},
	}

	o := NewOrchestrator(nil)
	result := o.Execute(plan)

	if result.Err == nil {
		t.Fatal("expected error")
	}
	if result.Rollback == nil {
		t.Fatal("rollback should have been executed")
	}
	// All succeeded steps across all stages should be rolled back.
	if !i1.rolledBack {
		t.Fatal("i1 should have been rolled back (cross-stage)")
	}
	if !d1.rolledBack {
		t.Fatal("d1 should have been rolled back (cross-stage)")
	}
	if !a1.rolledBack {
		t.Fatal("a1 should have been rolled back")
	}
	if a2.rolledBack {
		t.Fatal("a2 should NOT have been rolled back (it failed)")
	}
}

// --- DryExecute tests ---

func TestOrchestrator_DryExecute_EmitsEvents(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Install: []Step{&fakeStep{id: "inst1"}},
		Deploy:  []Step{&fakeStep{id: "dep1"}},
		Apply:   []Step{&fakeStep{id: "apply1"}},
	}

	var events []ProgressEvent
	progress := func(e ProgressEvent) {
		events = append(events, e)
	}

	o := NewOrchestrator(progress)
	result := o.DryExecute(plan)

	// 2 events per step (running + succeeded) × 4 steps = 8 events
	if len(events) != 8 {
		t.Fatalf("expected 8 events, got %d", len(events))
	}
	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func TestOrchestrator_DryExecute_NoSideEffects(t *testing.T) {
	s1 := &fakeStep{id: "prep1"}
	s2 := &fakeStep{id: "inst1"}
	s3 := &fakeStep{id: "dep1"}
	s4 := &fakeStep{id: "apply1"}

	plan := StagePlan{
		Prepare: []Step{s1},
		Install: []Step{s2},
		Deploy:  []Step{s3},
		Apply:   []Step{s4},
	}

	o := NewOrchestrator(nil)
	o.DryExecute(plan)

	for _, s := range []*fakeStep{s1, s2, s3, s4} {
		if s.ran {
			t.Fatalf("step %q should NOT have been run in dry-execute", s.id)
		}
	}
}

func TestOrchestrator_DryExecute_EmptyPlan(t *testing.T) {
	plan := StagePlan{}

	o := NewOrchestrator(nil)
	result := o.DryExecute(plan)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
}

func TestOrchestrator_DryExecute_StepResults(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Apply:   []Step{&fakeStep{id: "apply1"}, &fakeStep{id: "apply2"}},
	}

	o := NewOrchestrator(nil)
	result := o.DryExecute(plan)

	if len(result.Prepare.Steps) != 1 {
		t.Fatalf("expected 1 prepare step result, got %d", len(result.Prepare.Steps))
	}
	if len(result.Apply.Steps) != 2 {
		t.Fatalf("expected 2 apply step results, got %d", len(result.Apply.Steps))
	}
	for _, sr := range result.Apply.Steps {
		if sr.Status != StatusSucceeded {
			t.Fatalf("step %q status = %q, want %q", sr.StepID, sr.Status, StatusSucceeded)
		}
	}
	if result.Rollback != nil {
		t.Fatal("dry-execute should not produce rollback")
	}
}

func TestOrchestrator_DryExecute_FailingStepNotRun(t *testing.T) {
	// Even steps configured to fail should "succeed" in dry-execute (they're not run).
	s1 := &fakeStep{id: "fail-step", err: errors.New("would fail")}
	plan := StagePlan{
		Apply: []Step{s1},
	}

	o := NewOrchestrator(nil)
	result := o.DryExecute(plan)

	if result.Err != nil {
		t.Fatalf("dry-execute should not fail: %v", result.Err)
	}
	if s1.ran {
		t.Fatal("step should NOT have been run")
	}
	if result.Apply.Steps[0].Status != StatusSucceeded {
		t.Fatalf("status = %q, want %q", result.Apply.Steps[0].Status, StatusSucceeded)
	}
}

func TestOrchestrator_ProgressEvents_AllFourStages(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "prep1"}},
		Install: []Step{&fakeStep{id: "inst1"}},
		Deploy:  []Step{&fakeStep{id: "dep1"}},
		Apply:   []Step{&fakeStep{id: "apply1"}},
	}

	var events []ProgressEvent
	progress := func(e ProgressEvent) {
		events = append(events, e)
	}

	o := NewOrchestrator(progress)
	o.Execute(plan)

	// 2 events per step (running + succeeded) × 4 steps = 8 events
	if len(events) != 8 {
		t.Fatalf("expected 8 events, got %d", len(events))
	}
}
