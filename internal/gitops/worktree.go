package gitops

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
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
	path, err := m.worktreePath(branch)
	if err != nil {
		return "", err
	}
	if err := m.git(ctx, "worktree", "add", "-b", branch, "--", path, "HEAD"); err != nil {
		if !isBranchExistsError(err) {
			return "", err
		}
		if err := m.git(ctx, "worktree", "add", "--", path, branch); err != nil {
			if statErr := ensureExistingPath(path); statErr == nil {
				return path, nil
			}
			return "", err
		}
	}
	return path, nil
}

func (m Manager) CreateShadow(ctx context.Context, sourceRef, name string) (string, error) {
	path, err := m.worktreePath(name)
	if err != nil {
		return "", err
	}
	if err := m.git(ctx, "worktree", "add", "--", path, sourceRef); err != nil {
		return "", err
	}
	return path, nil
}

func (m Manager) Cleanup(ctx context.Context, path string) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}
	return m.git(ctx, "worktree", "remove", "--force", "--", absPath)
}

func (m Manager) Prune(ctx context.Context) error {
	return m.git(ctx, "worktree", "prune")
}

func (m Manager) CleanupOrphans(ctx context.Context) error {
	return m.Prune(ctx)
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
	var builder strings.Builder
	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
		case r == '-', r == '_':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}

	sanitized := strings.Trim(builder.String(), "-_.")
	sanitized = collapseDashes(sanitized)
	if sanitized == "" {
		return "worktree"
	}
	if strings.HasPrefix(sanitized, "-") {
		sanitized = "wt" + sanitized
	}
	return sanitized
}

func collapseDashes(value string) string {
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	return value
}

func (m Manager) worktreePath(name string) (string, error) {
	return m.validatePath(filepath.Join(m.RootDir, sanitize(name)))
}

func (m Manager) validatePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	rootAbs, err := filepath.Abs(m.RootDir)
	if err != nil {
		return "", err
	}
	if absPath != rootAbs && !strings.HasPrefix(absPath, rootAbs+string(filepath.Separator)) {
		return "", fmt.Errorf("refusing to use worktree outside root dir: %s", absPath)
	}
	return absPath, nil
}

func isBranchExistsError(err error) bool {
	message := err.Error()
	return strings.Contains(message, "already exists") || strings.Contains(message, "already checked out")
}

func ensureExistingPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("worktree path is not a directory: %s", path)
	}
	return nil
}
