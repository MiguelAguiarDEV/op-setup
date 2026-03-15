package pipeline

// StepStatus represents the current state of a step.
type StepStatus string

const (
	StatusPending    StepStatus = "pending"
	StatusRunning    StepStatus = "running"
	StatusSucceeded  StepStatus = "succeeded"
	StatusFailed     StepStatus = "failed"
	StatusRolledBack StepStatus = "rolled-back"
)

// StepResult holds the outcome of a single step execution.
type StepResult struct {
	StepID string
	Status StepStatus
	Err    error
}

// StageResult holds the outcome of all steps in a stage.
type StageResult struct {
	Stage   Stage
	Steps   []StepResult
	Success bool
}

// ExecutionResult holds the outcome of the entire pipeline.
type ExecutionResult struct {
	Prepare  StageResult
	Install  StageResult
	Deploy   StageResult
	Apply    StageResult
	Rollback *StageResult // nil if no rollback was needed.
	Err      error
}
