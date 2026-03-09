package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "modernc.org/sqlite"

	"github.com/soudai/saga/internal/store"
)

type Store struct {
	db *sql.DB
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	s := &Store{db: db}
	if err := s.initSchema(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS tasks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  repository TEXT NOT NULL,
  issue_number INTEGER NOT NULL,
  state TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  task_id INTEGER NOT NULL,
  state TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS leases (
  scope TEXT PRIMARY KEY,
  holder TEXT NOT NULL,
  expires_at TEXT NOT NULL
);
`
	_, err := s.db.Exec(schema)
	if err != nil {
		return fmt.Errorf("init schema: %w", err)
	}
	return nil
}

func (s *Store) CreateTask(ctx context.Context, repository string, issueNumber int64, state store.TaskState) (store.Task, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO tasks (repository, issue_number, state, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		repository,
		issueNumber,
		string(state),
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return store.Task{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return store.Task{}, err
	}

	return store.Task{
		ID:          id,
		Repository:  repository,
		IssueNumber: issueNumber,
		State:       state,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (s *Store) ListTasks(ctx context.Context) ([]store.Task, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, repository, issue_number, state, created_at, updated_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []store.Task
	for rows.Next() {
		var (
			task                 store.Task
			state                string
			createdAt, updatedAt string
		)
		if err := rows.Scan(&task.ID, &task.Repository, &task.IssueNumber, &state, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		task.State = store.TaskState(state)
		task.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		task.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}
	return tasks, rows.Err()
}

func (s *Store) UpdateTaskState(ctx context.Context, id int64, state store.TaskState) error {
	result, err := s.db.ExecContext(
		ctx,
		`UPDATE tasks SET state = ?, updated_at = ? WHERE id = ?`,
		string(state),
		time.Now().UTC().Format(time.RFC3339Nano),
		id,
	)
	if err != nil {
		return err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return store.ErrTaskNotFound
	}
	return nil
}

func (s *Store) CreateRun(ctx context.Context, taskID int64, state string) (store.Run, error) {
	now := time.Now().UTC()
	result, err := s.db.ExecContext(
		ctx,
		`INSERT INTO runs (task_id, state, created_at, updated_at) VALUES (?, ?, ?, ?)`,
		taskID,
		state,
		now.Format(time.RFC3339Nano),
		now.Format(time.RFC3339Nano),
	)
	if err != nil {
		return store.Run{}, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return store.Run{}, err
	}

	return store.Run{ID: id, TaskID: taskID, State: state, CreatedAt: now, UpdatedAt: now}, nil
}

func (s *Store) ListActiveRuns(ctx context.Context) ([]store.Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, task_id, state, created_at, updated_at FROM runs WHERE state = 'running' ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []store.Run
	for rows.Next() {
		var (
			run                  store.Run
			createdAt, updatedAt string
		)
		if err := rows.Scan(&run.ID, &run.TaskID, &run.State, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		run.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		run.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (s *Store) AcquireLease(ctx context.Context, scope, holder string, expiresAt time.Time) error {
	_, err := s.db.ExecContext(
		ctx,
		`INSERT INTO leases (scope, holder, expires_at) VALUES (?, ?, ?)
		 ON CONFLICT(scope) DO UPDATE SET holder = excluded.holder, expires_at = excluded.expires_at`,
		scope,
		holder,
		expiresAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) Status(ctx context.Context) (store.Status, error) {
	tasks, err := s.ListTasks(ctx)
	if err != nil {
		return store.Status{}, err
	}

	runs, err := s.ListActiveRuns(ctx)
	if err != nil {
		return store.Status{}, err
	}

	return store.Status{
		Tasks:      tasks,
		ActiveRuns: len(runs),
	}, nil
}

var _ store.Store = (*Store)(nil)

func IsTaskNotFound(err error) bool {
	return errors.Is(err, store.ErrTaskNotFound)
}
