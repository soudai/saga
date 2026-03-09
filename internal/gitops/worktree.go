package gitops

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	RepoPath string
	RootDir  string
}

func NewManager(repoPath, rootDir string) Manager {
	return Manager{
		RepoPath: repoPath,
		RootDir:  rootDir,
	}
}

func (m Manager) BranchName(issueNumber int64, taskID string) string {
	return fmt.Sprintf("saga/issue-%d/%s", issueNumber, sanitize(taskID))
}

func (m Manager) CreatePrimary(ctx context.Context, branch string) (string, error) {
	path := filepath.Join(m.RootDir, sanitize(branch))
	if err := m.git(ctx, "worktree", "add", "-b", branch, path, "HEAD"); err != nil {
		return "", err
	}
	return path, nil
}

func (m Manager) CreateShadow(ctx context.Context, sourceRef, name string) (string, error) {
	path := filepath.Join(m.RootDir, sanitize(name))
	if err := m.git(ctx, "worktree", "add", path, sourceRef); err != nil {
		return "", err
	}
	return path, nil
}

func (m Manager) Cleanup(ctx context.Context, path string) error {
	return m.git(ctx, "worktree", "remove", "--force", path)
}

func (m Manager) Prune(ctx context.Context) error {
	return m.git(ctx, "worktree", "prune")
}

func (m Manager) git(ctx context.Context, args ...string) error {
	command := exec.CommandContext(ctx, "git", append([]string{"-C", m.RepoPath}, args...)...)
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return nil
}

func sanitize(value string) string {
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.ReplaceAll(value, " ", "-")
	return value
}
