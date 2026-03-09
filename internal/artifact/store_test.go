package artifact

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFileUsesRestrictivePermissions(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store := New(root)

	path, err := store.WriteFile("run-1", "planner", "stdout.log", []byte("hello"))
	if err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("file mode = %o, want 600", got)
	}

	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("Stat(dir) error = %v", err)
	}
	if got := dirInfo.Mode().Perm(); got != 0o700 {
		t.Fatalf("dir mode = %o, want 700", got)
	}
}

func TestWriteFileRejectsPathTraversal(t *testing.T) {
	t.Parallel()

	store := New(t.TempDir())
	if _, err := store.WriteFile("../run", "planner", "stdout.log", []byte("hello")); err == nil {
		t.Fatal("WriteFile() error = nil, want invalid run id error")
	}
	if _, err := store.WriteFile("run-1", "../planner", "stdout.log", []byte("hello")); err == nil {
		t.Fatal("WriteFile() error = nil, want invalid stage name error")
	}
}
