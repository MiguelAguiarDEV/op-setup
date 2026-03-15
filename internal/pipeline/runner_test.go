package pipeline

import (
	"errors"
	"testing"
)

type fakeStep struct {
	id  string
	err error
	ran bool
}

func (s *fakeStep) ID() string { return s.id }
func (s *fakeStep) Run() error { s.ran = true; return s.err }

type fakeRollbackStep struct {
	fakeStep
	rolledBack  bool
	rollbackErr error
}

func (s *fakeRollbackStep) Rollback() error {
	s.rolledBack = true
	return s.rollbackErr
}

func TestRunner_AllSucceed(t *testing.T) {
	s1 := &fakeStep{id: "s1"}
	s2 := &fakeStep{id: "s2"}

	r := NewRunner(StopOnError, nil)
	result := r.Run(StagePrepare, []Step{s1, s2})

	if !result.Success {
		t.Fatal("expected success")
	}
	if len(result.Steps) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(result.Steps))
	}
	if !s1.ran || !s2.ran {
		t.Fatal("all steps should have run")
	}
}

func TestRunner_StopOnError(t *testing.T) {
	s1 := &fakeStep{id: "s1", err: errors.New("fail")}
	s2 := &fakeStep{id: "s2"}

	r := NewRunner(StopOnError, nil)
	result := r.Run(StageApply, []Step{s1, s2})

	if result.Success {
		t.Fatal("expected failure")
	}
	if len(result.Steps) != 1 {
		t.Fatalf("expected 1 step result (stopped), got %d", len(result.Steps))
	}
	if s2.ran {
		t.Fatal("s2 should not have run after s1 failed")
	}
}

func TestRunner_ContinueOnError(t *testing.T) {
	s1 := &fakeStep{id: "s1", err: errors.New("fail")}
	s2 := &fakeStep{id: "s2"}

	r := NewRunner(ContinueOnError, nil)
	result := r.Run(StageApply, []Step{s1, s2})

	if result.Success {
		t.Fatal("expected failure")
	}
	if len(result.Steps) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(result.Steps))
	}
	if !s2.ran {
		t.Fatal("s2 should have run despite s1 failure")
	}
}

func TestRunner_ProgressEvents(t *testing.T) {
	s1 := &fakeStep{id: "s1"}

	var events []ProgressEvent
	progress := func(e ProgressEvent) {
		events = append(events, e)
	}

	r := NewRunner(StopOnError, progress)
	r.Run(StagePrepare, []Step{s1})

	if len(events) != 2 {
		t.Fatalf("expected 2 events (running + succeeded), got %d", len(events))
	}
	if events[0].Status != StatusRunning {
		t.Fatalf("first event status = %q, want %q", events[0].Status, StatusRunning)
	}
	if events[1].Status != StatusSucceeded {
		t.Fatalf("second event status = %q, want %q", events[1].Status, StatusSucceeded)
	}
}

func TestRunner_ProgressEvents_Failure(t *testing.T) {
	s1 := &fakeStep{id: "s1", err: errors.New("fail")}

	var events []ProgressEvent
	progress := func(e ProgressEvent) {
		events = append(events, e)
	}

	r := NewRunner(StopOnError, progress)
	r.Run(StageApply, []Step{s1})

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[1].Status != StatusFailed {
		t.Fatalf("second event status = %q, want %q", events[1].Status, StatusFailed)
	}
	if events[1].Err == nil {
		t.Fatal("failed event should have error")
	}
}

func TestRunner_EmptySteps(t *testing.T) {
	r := NewRunner(StopOnError, nil)
	result := r.Run(StagePrepare, []Step{})

	if !result.Success {
		t.Fatal("empty steps should succeed")
	}
	if len(result.Steps) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result.Steps))
	}
}
