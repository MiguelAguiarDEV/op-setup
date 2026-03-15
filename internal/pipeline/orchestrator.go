package pipeline

import "fmt"

// Orchestrator executes a StagePlan: Prepare → Apply, with rollback on failure.
type Orchestrator struct {
	progress ProgressFunc
}

// NewOrchestrator creates an Orchestrator with the given progress callback.
func NewOrchestrator(progress ProgressFunc) *Orchestrator {
	if progress == nil {
		progress = func(ProgressEvent) {}
	}
	return &Orchestrator{progress: progress}
}

// Execute runs the StagePlan.
//
//  1. Runs all Prepare steps with StopOnError policy.
//     If any Prepare step fails, returns immediately (no rollback for prepare).
//  2. Runs all Apply steps with StopOnError policy.
//     If any Apply step fails, rolls back all succeeded Apply steps in reverse.
func (o *Orchestrator) Execute(plan StagePlan) ExecutionResult {
	result := ExecutionResult{}

	// Stage 1: Prepare
	prepareRunner := NewRunner(StopOnError, o.progress)
	result.Prepare = prepareRunner.Run(StagePrepare, plan.Prepare)

	if !result.Prepare.Success {
		result.Err = fmt.Errorf("prepare stage failed")
		return result
	}

	// Stage 2: Apply
	applyRunner := NewRunner(StopOnError, o.progress)
	result.Apply = applyRunner.Run(StageApply, plan.Apply)

	if !result.Apply.Success {
		// Collect succeeded apply steps for rollback.
		var succeeded []Step
		for i, sr := range result.Apply.Steps {
			if sr.Status == StatusSucceeded {
				succeeded = append(succeeded, plan.Apply[i])
			}
		}

		// Rollback succeeded steps.
		rollbackResult := ExecuteRollback(succeeded, o.progress)
		result.Rollback = &rollbackResult
		result.Err = fmt.Errorf("apply stage failed, rollback executed")
	}

	return result
}
