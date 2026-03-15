package pipeline

// ExecuteRollback iterates the succeeded steps in reverse order and calls
// Rollback() on each that implements RollbackStep.
// Returns a StageResult describing the rollback outcomes.
func ExecuteRollback(succeededSteps []Step, progress ProgressFunc) StageResult {
	if progress == nil {
		progress = func(ProgressEvent) {}
	}

	result := StageResult{
		Stage:   StageRollback,
		Steps:   make([]StepResult, 0),
		Success: true,
	}

	// Iterate in reverse order.
	for i := len(succeededSteps) - 1; i >= 0; i-- {
		step := succeededSteps[i]
		rs, ok := step.(RollbackStep)
		if !ok {
			continue // Step doesn't support rollback — skip.
		}

		progress(ProgressEvent{
			Stage:  StageRollback,
			StepID: rs.ID(),
			Status: StatusRunning,
		})

		err := rs.Rollback()
		status := StatusRolledBack
		if err != nil {
			status = StatusFailed
			result.Success = false
		}

		progress(ProgressEvent{
			Stage:  StageRollback,
			StepID: rs.ID(),
			Status: status,
			Err:    err,
		})

		result.Steps = append(result.Steps, StepResult{
			StepID: rs.ID(),
			Status: status,
			Err:    err,
		})
	}

	return result
}
