package pipeline

import "testing"

func TestStagePlan_TotalSteps_Empty(t *testing.T) {
	plan := StagePlan{}
	if got := plan.TotalSteps(); got != 0 {
		t.Fatalf("TotalSteps() = %d, want 0", got)
	}
}

func TestStagePlan_TotalSteps_AllStages(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "p1"}, &fakeStep{id: "p2"}},
		Install: []Step{&fakeStep{id: "i1"}},
		Deploy:  []Step{&fakeStep{id: "d1"}},
		Apply:   []Step{&fakeStep{id: "a1"}, &fakeStep{id: "a2"}, &fakeStep{id: "a3"}},
	}
	if got := plan.TotalSteps(); got != 7 {
		t.Fatalf("TotalSteps() = %d, want 7", got)
	}
}

func TestStagePlan_TotalSteps_PartialStages(t *testing.T) {
	plan := StagePlan{
		Prepare: []Step{&fakeStep{id: "p1"}},
		Apply:   []Step{&fakeStep{id: "a1"}},
	}
	if got := plan.TotalSteps(); got != 2 {
		t.Fatalf("TotalSteps() = %d, want 2", got)
	}
}
