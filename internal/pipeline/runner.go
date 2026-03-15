package pipeline

// Runner executes a list of steps with a given failure policy.
type Runner struct {
	Policy   FailurePolicy
	Progress ProgressFunc
}

// NewRunner creates a Runner with the given failure policy.
func NewRunner(policy FailurePolicy, progress ProgressFunc) *Runner {
	if progress == nil {
		progress = func(ProgressEvent) {}
	}
	return &Runner{Policy: policy, Progress: progress}
}

// Run executes all steps in order. Returns a StageResult.
// If Policy is StopOnError, stops on the first failure.
// If Policy is ContinueOnError, runs all steps and collects errors.
func (r *Runner) Run(stage Stage, steps []Step) StageResult {
	result := StageResult{
		Stage:   stage,
		Steps:   make([]StepResult, 0, len(steps)),
		Success: true,
	}

	for _, step := range steps {
		r.Progress(ProgressEvent{
			Stage:  stage,
			StepID: step.ID(),
			Status: StatusRunning,
		})

		err := step.Run()
		status := StatusSucceeded
		if err != nil {
			status = StatusFailed
			result.Success = false
		}

		r.Progress(ProgressEvent{
			Stage:  stage,
			StepID: step.ID(),
			Status: status,
			Err:    err,
		})

		result.Steps = append(result.Steps, StepResult{
			StepID: step.ID(),
			Status: status,
			Err:    err,
		})

		if err != nil && r.Policy == StopOnError {
			break
		}
	}

	return result
}
