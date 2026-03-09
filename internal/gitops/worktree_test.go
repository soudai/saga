package gitops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestManagerCreateAndCleanup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	repoPath := filepath.Join(dir, "repo")
	worktreesPath := filepath.Join(dir, "worktrees")

	mustRun(t, dir, "git", "init", repoPath)
	mustRun(t, repoPath, "git", "config", "user.name", "Saga Test")
	mustRun(t, repoPath, "git", "config", "user.email", "saga@example.com")
	if err := os.WriteFile(filepath.Join(repoPath, "README.md"), []byte("# repo\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	mustRun(t, repoPath, "git", "add", "README.md")
	mustRun(t, repoPath, "git", "commit", "-m", "initial commit")

	manager := NewManager(repoPath, worktreesPath)
	ctx := context.Background()

	primary, err := manager.CreatePrimary(ctx, manager.BranchName(12, "task-1"))
	if err != nil {
		t.Fatalf("CreatePrimary() error = %v", err)
	}
	if _, err := os.Stat(primary); err != nil {
		t.Fatalf("primary worktree error = %v", err)
	}

	reusedPrimary, err := manager.CreatePrimary(ctx, manager.BranchName(12, "task-1"))
	if err != nil {
		t.Fatalf("CreatePrimary() reuse error = %v", err)
	}
	if reusedPrimary != primary {
		t.Fatalf("reused primary = %q, want %q", reusedPrimary, primary)
	}

	shadow, err := manager.CreateShadow(ctx, "HEAD", "shadow-task-1")
	if err != nil {
		t.Fatalf("CreateShadow() error = %v", err)
	}
	if _, err := os.Stat(shadow); err != nil {
		t.Fatalf("shadow worktree error = %v", err)
	}

	if err := manager.Cleanup(ctx, shadow); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
	if _, err := os.Stat(shadow); !os.IsNotExist(err) {
		t.Fatalf("shadow stat error = %v, want not exist", err)
	}
	if err := manager.Cleanup(ctx, primary); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}
	if _, err := os.Stat(primary); !os.IsNotExist(err) {
		t.Fatalf("primary stat error = %v, want not exist", err)
	}
	if err := manager.CleanupOrphans(ctx); err != nil {
		t.Fatalf("CleanupOrphans() error = %v", err)
	}

	worktreeList := mustRunOutput(t, repoPath, "git", "worktree", "list")
	if strings.Contains(worktreeList, shadow) {
		t.Fatalf("worktree list still contains shadow path: %s", worktreeList)
	}
	if strings.Contains(worktreeList, primary) {
		t.Fatalf("worktree list still contains primary path: %s", worktreeList)
	}
}

func mustRun(t *testing.T, dir string, name string, args ...string) {
	t.Helper()

	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v: %s", name, args, err, string(output))
	}
}

func mustRunOutput(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()

	command := exec.Command(name, args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s %v: %v: %s", name, args, err, string(output))
	}
	return string(output)
}
