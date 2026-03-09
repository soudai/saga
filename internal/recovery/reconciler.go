package recovery

import "github.com/soudai/saga/internal/store"

type Decision string

const (
	DecisionKeepRunning Decision = "keep_running"
	DecisionRetry       Decision = "retry"
	DecisionComplete    Decision = "complete"
	DecisionCancel      Decision = "cancel"
)

type Snapshot struct {
	Task         store.Task
	PROpen       bool
	BranchExists bool
}

func Reconcile(snapshot Snapshot) Decision {
	switch snapshot.Task.State {
	case store.TaskStateCancelled:
		return DecisionCancel
	case store.TaskStateQueued:
		return DecisionRetry
	case store.TaskStateRunning:
		if snapshot.PROpen || snapshot.BranchExists {
			return DecisionKeepRunning
		}
		return DecisionRetry
	case store.TaskStateCompleted:
		return DecisionComplete
	case store.TaskStateFailed:
		return DecisionRetry
	default:
		if snapshot.PROpen {
			return DecisionComplete
		}
		return DecisionRetry
	}
}
