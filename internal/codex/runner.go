package codex

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
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

	artifactStore := artifact.New(req.ArtifactRoot)
	stdoutFile, stdoutPath, err := artifactStore.CreateFile(req.RunID, req.StageName, "stdout.log")
	if err != nil {
		return Result{}, err
	}
	defer stdoutFile.Close()

	stderrFile, stderrPath, err := artifactStore.CreateFile(req.RunID, req.StageName, "stderr.log")
	if err != nil {
		return Result{}, err
	}
	defer stderrFile.Close()

	cmd := exec.Command(req.CommandPath, req.Args...)
	cmd.Dir = req.WorkDir
	cmd.Env = os.Environ()
	for key, value := range req.Env {
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	cmd.Env = append(cmd.Env,
		"SAGA_SANDBOX="+string(req.Sandbox),
		fmt.Sprintf("SAGA_NETWORK=%t", req.Network),
		"SAGA_MODEL="+req.Model,
	)
	cmd.Stdout = stdoutFile
	cmd.Stderr = stderrFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	startedAt := time.Now().UTC()
	result := Result{
		Command:    append([]string{req.CommandPath}, req.Args...),
		Sandbox:    req.Sandbox,
		Network:    req.Network,
		Model:      req.Model,
		ExitCode:   -1,
		StdoutPath: stdoutPath,
		StderrPath: stderrPath,
		StartedAt:  startedAt,
	}

	if err := cmd.Start(); err != nil {
		result.FinishedAt = time.Now().UTC()
		if _, writeErr := artifactStore.WriteJSON(req.RunID, req.StageName, "result.json", result); writeErr != nil {
			return Result{}, writeErr
		}
		return result, fmt.Errorf("start command: %w", err)
	}

	done := make(chan struct{})
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			_ = killProcessGroup(cmd.Process.Pid)
		case <-done:
		}
	}()

	waitCh := make(chan error, 1)
	go func() {
		waitCh <- cmd.Wait()
	}()

	runErr := <-waitCh
	finishedAt := time.Now().UTC()

	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else if ctx.Err() != nil {
			exitCode = -1
		} else {
			result.FinishedAt = finishedAt
			if _, writeErr := artifactStore.WriteJSON(req.RunID, req.StageName, "result.json", result); writeErr != nil {
				return Result{}, writeErr
			}
			return result, fmt.Errorf("run command: %w", runErr)
		}
	}

	result.ExitCode = exitCode
	result.FinishedAt = finishedAt
	result.TimedOut = ctx.Err() == context.DeadlineExceeded

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

func killProcessGroup(pid int) error {
	if pid <= 0 {
		return nil
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
		return err
	}
	return nil
}
