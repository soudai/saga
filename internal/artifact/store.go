package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Store struct {
	Root string
}

func New(root string) Store {
	return Store{Root: root}
}

func (s Store) StageDir(runID, stageName string) string {
	return filepath.Join(s.Root, runID, stageName)
}

func (s Store) EnsureStageDir(runID, stageName string) (string, error) {
	dir := s.StageDir(runID, stageName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("mkdir artifact dir: %w", err)
	}
	return dir, nil
}

func (s Store) WriteFile(runID, stageName, name string, data []byte) (string, error) {
	dir, err := s.EnsureStageDir(runID, stageName)
	if err != nil {
		return "", err
	}

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write artifact: %w", err)
	}
	return path, nil
}

func (s Store) WriteJSON(runID, stageName, name string, value any) (string, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}
	data = append(data, '\n')
	return s.WriteFile(runID, stageName, name, data)
}
