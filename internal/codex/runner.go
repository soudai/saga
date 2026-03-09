package codex

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/soudai/saga/internal/artifact"
)

type SandboxMode string

const (
	SandboxReadOnly       SandboxMode = "read-only"
	SandboxWorkspaceWrite SandboxMode = "workspace-write"
	SandboxFull           SandboxMode = "full"
)

type RunnerRequest struct {
	RunID        string
	StageName    string
	CommandPath  string
	Args         []string
	Env          map[string]string
	WorkDir      string
	Timeout      time.Duration
	Sandbox      SandboxMode
	Network      bool
	Model        string
	ArtifactRoot string
}

type Result struct {
	Command    []string    `json:"command"`
	Sandbox    SandboxMode `json:"sandbox"`
	Network    bool        `json:"network"`
	Model      string      `json:"model,omitempty"`
	ExitCode   int         `json:"exit_code"`
	StdoutPath string      `json:"stdout_path"`
	StderrPath string      `json:"stderr_path"`
	StartedAt  time.Time   `json:"started_at"`
	FinishedAt time.Time   `json:"finished_at"`
	TimedOut   bool        `json:"timed_out"`
}

type Runner struct{}

func NewRunner() Runner {
	return Runner{}
}

func (Runner) Run(ctx context.Context, req RunnerRequest) (Result, error) {
	if req.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	}

	startedAt := time.Now().UTC()

	cmd := exec.CommandContext(ctx, req.CommandPath, req.Args...)
	cmd.Dir = req.WorkDir
	cmd.Env = os.Environ()
	for key, value := range req.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	finishedAt := time.Now().UTC()

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			exitCode = -1
		} else {
			return Result{}, fmt.Errorf("run command: %w", runErr)
		}
	}

	artifactStore := artifact.New(req.ArtifactRoot)
	stdoutPath, err := artifactStore.WriteFile(req.RunID, req.StageName, "stdout.log", stdout.Bytes())
	if err != nil {
		return Result{}, err
	}
	stderrPath, err := artifactStore.WriteFile(req.RunID, req.StageName, "stderr.log", stderr.Bytes())
	if err != nil {
		return Result{}, err
	}

	result := Result{
		Command:    append([]string{req.CommandPath}, req.Args...),
		Sandbox:    req.Sandbox,
		Network:    req.Network,
		Model:      req.Model,
		ExitCode:   exitCode,
		StdoutPath: stdoutPath,
		StderrPath: stderrPath,
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
		TimedOut:   ctx.Err() == context.DeadlineExceeded,
	}

	if _, err := artifactStore.WriteJSON(req.RunID, req.StageName, "result.json", result); err != nil {
		return Result{}, err
	}

	if runErr != nil && ctx.Err() != nil {
		return result, ctx.Err()
	}
	if runErr != nil {
		return result, runErr
	}
	return result, nil
}
