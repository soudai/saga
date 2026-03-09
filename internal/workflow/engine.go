package workflow

import (
	"context"
	"fmt"
)

type StageStatus string

const (
	StageStatusSucceeded StageStatus = "succeeded"
	StageStatusFailed    StageStatus = "failed"
)

type StageResult struct {
	Status    StageStatus
	Artifacts map[string]string
}

type StageRunner func(context.Context, Stage, map[string]string) (StageResult, error)

type Engine struct {
	runner StageRunner
}

func NewEngine(runner StageRunner) Engine {
	return Engine{runner: runner}
}

func (e Engine) Execute(ctx context.Context, workflow Workflow) (map[string]StageResult, error) {
	stageIndex := make(map[string]int, len(workflow.Stages))
	for i, stage := range workflow.Stages {
		stageIndex[stage.Name] = i
	}

	results := make(map[string]StageResult, len(workflow.Stages))
	artifacts := make(map[string]string)

	for index := 0; index < len(workflow.Stages); {
		stage := workflow.Stages[index]

		var (
			result StageResult
			err    error
		)
		attempts := stage.Retry + 1
		for attempt := 0; attempt < attempts; attempt++ {
			result, err = e.runner(ctx, stage, artifacts)
			if err == nil && result.Status == StageStatusSucceeded {
				break
			}
		}
		if err != nil {
			return results, fmt.Errorf("run stage %s: %w", stage.Name, err)
		}

		results[stage.Name] = result
		for key, value := range result.Artifacts {
			artifacts[key] = value
		}

		switch result.Status {
		case StageStatusSucceeded:
			if stage.Transition.Success == "" {
				index++
				continue
			}
			nextIndex, ok := stageIndex[stage.Transition.Success]
			if !ok {
				return results, fmt.Errorf("unknown success transition: %s", stage.Transition.Success)
			}
			index = nextIndex
		case StageStatusFailed:
			if stage.Transition.Failure == "" {
				return results, fmt.Errorf("stage %s failed without failure transition", stage.Name)
			}
			nextIndex, ok := stageIndex[stage.Transition.Failure]
			if !ok {
				return results, fmt.Errorf("unknown failure transition: %s", stage.Transition.Failure)
			}
			index = nextIndex
		default:
			return results, fmt.Errorf("unsupported stage status: %s", result.Status)
		}
	}

	return results, nil
}
