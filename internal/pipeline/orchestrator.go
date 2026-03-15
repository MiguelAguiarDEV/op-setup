package pipeline

import "fmt"

// Orchestrator executes a StagePlan: Prepare → Install → Deploy → Apply,
// with rollback on failure.
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
//  2. Runs all Install steps with StopOnError policy.
//     If any Install step fails, rolls back succeeded Install steps.
//  3. Runs all Deploy steps with StopOnError policy.
//     If any Deploy step fails, rolls back succeeded Deploy + Install steps.
//  4. Runs all Apply steps with StopOnError policy.
//     If any Apply step fails, rolls back succeeded Apply + Deploy + Install steps.
func (o *Orchestrator) Execute(plan StagePlan) ExecutionResult {
	result := ExecutionResult{}

	// Stage 1: Prepare
	if len(plan.Prepare) > 0 {
		prepareRunner := NewRunner(StopOnError, o.progress)
		result.Prepare = prepareRunner.Run(StagePrepare, plan.Prepare)
		if !result.Prepare.Success {
			result.Err = fmt.Errorf("prepare stage failed")
			return result
		}
	}

	// Track all succeeded steps across stages for rollback.
	var allSucceeded []Step

	// Stage 2: Install
	if len(plan.Install) > 0 {
		installRunner := NewRunner(StopOnError, o.progress)
		result.Install = installRunner.Run(StageInstall, plan.Install)
		if !result.Install.Success {
			allSucceeded = collectSucceeded(result.Install, plan.Install)
			rollbackResult := ExecuteRollback(allSucceeded, o.progress)
			result.Rollback = &rollbackResult
			result.Err = fmt.Errorf("install stage failed, rollback executed")
			return result
		}
		allSucceeded = append(allSucceeded, collectSucceeded(result.Install, plan.Install)...)
	}

	// Stage 3: Deploy
	if len(plan.Deploy) > 0 {
		deployRunner := NewRunner(StopOnError, o.progress)
		result.Deploy = deployRunner.Run(StageDeploy, plan.Deploy)
		if !result.Deploy.Success {
			allSucceeded = append(allSucceeded, collectSucceeded(result.Deploy, plan.Deploy)...)
			rollbackResult := ExecuteRollback(allSucceeded, o.progress)
			result.Rollback = &rollbackResult
			result.Err = fmt.Errorf("deploy stage failed, rollback executed")
			return result
		}
		allSucceeded = append(allSucceeded, collectSucceeded(result.Deploy, plan.Deploy)...)
	}

	// Stage 4: Apply
	if len(plan.Apply) > 0 {
		applyRunner := NewRunner(StopOnError, o.progress)
		result.Apply = applyRunner.Run(StageApply, plan.Apply)
		if !result.Apply.Success {
			allSucceeded = append(allSucceeded, collectSucceeded(result.Apply, plan.Apply)...)
			rollbackResult := ExecuteRollback(allSucceeded, o.progress)
			result.Rollback = &rollbackResult
			result.Err = fmt.Errorf("apply stage failed, rollback executed")
			return result
		}
	}

	return result
}

// collectSucceeded returns the steps that succeeded from a stage result.
func collectSucceeded(stageResult StageResult, steps []Step) []Step {
	var succeeded []Step
	for i, sr := range stageResult.Steps {
		if sr.Status == StatusSucceeded && i < len(steps) {
			succeeded = append(succeeded, steps[i])
		}
	}
	return succeeded
}
