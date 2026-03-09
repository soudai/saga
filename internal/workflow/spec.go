package workflow

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/soudai/saga/internal/codex"
)

type Workflow struct {
	Stages []Stage `yaml:"stages"`
}

type Stage struct {
	Name         string            `yaml:"name"`
	Role         string            `yaml:"role"`
	Sandbox      codex.SandboxMode `yaml:"sandbox"`
	Network      bool              `yaml:"network"`
	Timeout      time.Duration     `yaml:"-"`
	TimeoutRaw   string            `yaml:"timeout"`
	Retry        int               `yaml:"retry"`
	WorktreeMode string            `yaml:"worktree_mode"`
	Transition   Transition        `yaml:"transition"`
}

type Transition struct {
	Success string `yaml:"success"`
	Failure string `yaml:"failure"`
}

func Parse(data []byte) (Workflow, error) {
	var workflow Workflow
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return Workflow{}, fmt.Errorf("parse workflow: %w", err)
	}

	for i := range workflow.Stages {
		stage := &workflow.Stages[i]
		if stage.TimeoutRaw == "" {
			continue
		}
		timeout, err := time.ParseDuration(stage.TimeoutRaw)
		if err != nil {
			return Workflow{}, fmt.Errorf("parse timeout for stage %s: %w", stage.Name, err)
		}
		stage.Timeout = timeout
	}
	return workflow, nil
}
