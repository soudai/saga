package store

import (
	"context"
	"errors"
	"time"
)

var ErrTaskNotFound = errors.New("task not found")
var ErrLeaseHeld = errors.New("lease held by another worker")

type TaskState string

const (
	TaskStateQueued    TaskState = "queued"
	TaskStateRunning   TaskState = "running"
	TaskStateCancelled TaskState = "cancelled"
)

type Task struct {
	ID          int64
	Repository  string
	IssueNumber int64
	State       TaskState
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Run struct {
	ID        int64
	TaskID    int64
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Lease struct {
	Scope     string
	Holder    string
	ExpiresAt time.Time
}

type Status struct {
	Tasks      []Task
	ActiveRuns int
}

type Store interface {
	Close() error
	CreateTask(ctx context.Context, repository string, issueNumber int64, state TaskState) (Task, error)
	ListTasks(ctx context.Context) ([]Task, error)
	UpdateTaskState(ctx context.Context, id int64, state TaskState) error
	CreateRun(ctx context.Context, taskID int64, state string) (Run, error)
	ListActiveRuns(ctx context.Context) ([]Run, error)
	AcquireLease(ctx context.Context, scope, holder string, expiresAt time.Time) error
	Status(ctx context.Context) (Status, error)
}
