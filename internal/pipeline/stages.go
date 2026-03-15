// Package pipeline orchestrates the Prepare → Install → Deploy → Apply workflow.
package pipeline

// Stage identifies a phase of the pipeline.
type Stage string

const (
	StagePrepare  Stage = "prepare"
	StageInstall  Stage = "install"
	StageDeploy   Stage = "deploy"
	StageApply    Stage = "apply"
	StageRollback Stage = "rollback"
)

// Step is a single unit of work in the pipeline.
type Step interface {
	// ID returns a unique identifier for this step.
	ID() string

	// Run executes the step. Returns nil on success.
	Run() error
}

// RollbackStep is a Step that can undo its changes.
type RollbackStep interface {
	Step
	Rollback() error
}

// FailurePolicy determines how the runner handles step failures.
type FailurePolicy int

const (
	// StopOnError stops the pipeline on the first error.
	StopOnError FailurePolicy = iota

	// ContinueOnError continues executing remaining steps after an error.
	ContinueOnError
)

// ProgressEvent reports the status of a step execution.
type ProgressEvent struct {
	Stage  Stage
	StepID string
	Status StepStatus
	Err    error
}

// ProgressFunc is called for each step status change.
type ProgressFunc func(ProgressEvent)

// StagePlan defines the steps for each stage of the pipeline.
//
// Execution order: Prepare → Install → Deploy → Apply.
// Each stage is optional — empty slices are skipped.
type StagePlan struct {
	Prepare []Step
	Install []Step // Tool installation (Phase 1 installers).
	Deploy  []Step // Dotfile deployment (Phase 2 deployer).
	Apply   []Step // MCP config injection (original v1 behavior).
}

// TotalSteps returns the total number of steps across all stages.
func (p StagePlan) TotalSteps() int {
	return len(p.Prepare) + len(p.Install) + len(p.Deploy) + len(p.Apply)
}
