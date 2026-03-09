package control

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/soudai/saga/internal/store"
)

type Server struct {
	store store.Store
	mux   *http.ServeMux
}

type StatusResponse struct {
	Tasks      []store.Task `json:"tasks"`
	ActiveRuns int          `json:"active_runs"`
}

type EnqueueRequest struct {
	Repository  string `json:"repository"`
	IssueNumber int64  `json:"issue_number"`
}

func NewServer(store store.Store) *Server {
	server := &Server{
		store: store,
		mux:   http.NewServeMux(),
	}
	server.routes()
	return server
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("/status", s.handleStatus)
	s.mux.HandleFunc("/tasks", s.handleTasks)
	s.mux.HandleFunc("/tasks/", s.handleTaskAction)
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := s.store.Status(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusOK, StatusResponse{
		Tasks:      status.Tasks,
		ActiveRuns: status.ActiveRuns,
	})
}

func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/tasks" {
		writeError(w, http.StatusNotFound, errors.New("unknown route"))
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	var req EnqueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("decode request: %w", err))
		return
	}
	if req.Repository == "" {
		writeError(w, http.StatusBadRequest, errors.New("repository is required"))
		return
	}
	if req.IssueNumber <= 0 {
		writeError(w, http.StatusBadRequest, errors.New("issue_number must be greater than zero"))
		return
	}

	task, err := s.store.CreateTask(r.Context(), req.Repository, req.IssueNumber, store.TaskStateQueued)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusCreated, task)
}

func (s *Server) handleTaskAction(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/tasks/"), "/"), "/")
	if len(parts) != 2 {
		writeError(w, http.StatusNotFound, errors.New("unknown route"))
		return
	}

	taskID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid task id: %w", err))
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
		return
	}

	state, err := actionToState(parts[1])
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	if err := s.store.UpdateTaskState(r.Context(), taskID, state); err != nil {
		if errors.Is(err, store.ErrTaskNotFound) {
			writeError(w, http.StatusNotFound, err)
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func actionToState(action string) (store.TaskState, error) {
	switch action {
	case "cancel":
		return store.TaskStateCancelled, nil
	case "retry":
		return store.TaskStateQueued, nil
	case "resume":
		return store.TaskStateRunning, nil
	default:
		return "", errors.New("unknown action")
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, statusCode int, err error) {
	writeJSON(w, statusCode, map[string]string{"error": err.Error()})
}

func WithContext(ctx context.Context, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
