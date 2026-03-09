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

type ExecutionRecord struct {
	StageName string
	Attempt   int
	Result    StageResult
}

type StageRunner func(context.Context, Stage, map[string]string) (StageResult, error)

type Engine struct {
	runner StageRunner
}

func NewEngine(runner StageRunner) Engine {
	return Engine{runner: runner}
}

func (e Engine) Execute(ctx context.Context, workflow Workflow) ([]ExecutionRecord, error) {
	if err := Validate(workflow); err != nil {
		return nil, err
	}

	stageIndex := make(map[string]int, len(workflow.Stages))
	for i, stage := range workflow.Stages {
		stageIndex[stage.Name] = i
	}

	records := make([]ExecutionRecord, 0, len(workflow.Stages))
	artifacts := make(map[string]string)
	maxTransitions := len(workflow.Stages) * 16
	transitions := 0

	for index := 0; index < len(workflow.Stages); {
		if transitions >= maxTransitions {
			return records, fmt.Errorf("workflow exceeded max transitions: %d", maxTransitions)
		}

		stage := workflow.Stages[index]

		var (
			result StageResult
			err    error
		)
		attempts := stage.Retry + 1
		for attempt := 1; attempt <= attempts; attempt++ {
			result, err = e.runner(ctx, stage, cloneArtifacts(artifacts))
			records = append(records, ExecutionRecord{
				StageName: stage.Name,
				Attempt:   attempt,
				Result:    result,
			})
			if err == nil && result.Status == StageStatusSucceeded {
				break
			}
		}
		if err != nil {
			return records, fmt.Errorf("run stage %s: %w", stage.Name, err)
		}

		for key, value := range result.Artifacts {
			artifacts[key] = value
		}
		transitions++

		switch result.Status {
		case StageStatusSucceeded:
			if stage.Transition.Success == "" {
				index++
				continue
			}
			nextIndex, ok := stageIndex[stage.Transition.Success]
			if !ok {
				return records, fmt.Errorf("unknown success transition: %s", stage.Transition.Success)
			}
			index = nextIndex
		case StageStatusFailed:
			if stage.Transition.Failure == "" {
				return records, fmt.Errorf("stage %s failed without failure transition", stage.Name)
			}
			nextIndex, ok := stageIndex[stage.Transition.Failure]
			if !ok {
				return records, fmt.Errorf("unknown failure transition: %s", stage.Transition.Failure)
			}
			index = nextIndex
		default:
			return records, fmt.Errorf("unsupported stage status: %s", result.Status)
		}
	}

	return records, nil
}

func cloneArtifacts(artifacts map[string]string) map[string]string {
	if len(artifacts) == 0 {
		return map[string]string{}
	}

	cloned := make(map[string]string, len(artifacts))
	for key, value := range artifacts {
		cloned[key] = value
	}
	return cloned
}
