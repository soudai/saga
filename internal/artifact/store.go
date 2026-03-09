package artifact

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

type Store struct {
	Root string
}

var safeNamePattern = regexp.MustCompile(`\A[A-Za-z0-9][A-Za-z0-9_.-]*\z`)

func New(root string) Store {
	return Store{Root: root}
}

func (s Store) StageDir(runID, stageName string) (string, error) {
	if err := validatePathSegment("run_id", runID); err != nil {
		return "", err
	}
	if err := validatePathSegment("stage_name", stageName); err != nil {
		return "", err
	}
	return filepath.Join(s.Root, runID, stageName), nil
}

func (s Store) EnsureStageDir(runID, stageName string) (string, error) {
	dir, err := s.StageDir(runID, stageName)
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("mkdir artifact dir: %w", err)
	}
	return dir, nil
}

func (s Store) CreateFile(runID, stageName, name string) (*os.File, string, error) {
	if err := validatePathSegment("artifact_name", name); err != nil {
		return nil, "", err
	}
	dir, err := s.EnsureStageDir(runID, stageName)
	if err != nil {
		return nil, "", err
	}

	path := filepath.Join(dir, name)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, "", fmt.Errorf("create artifact: %w", err)
	}
	return file, path, nil
}

func (s Store) WriteFile(runID, stageName, name string, data []byte) (string, error) {
	file, path, err := s.CreateFile(runID, stageName, name)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
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

func validatePathSegment(field, value string) error {
	if !safeNamePattern.MatchString(value) {
		return fmt.Errorf("%s contains unsupported characters: %s", field, value)
	}
	return nil
}
