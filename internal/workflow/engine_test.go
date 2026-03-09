package workflow

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestEngineExecute(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{
				Name: "plan",
				Role: "planner",
				Transition: Transition{
					Success: "implement",
				},
			},
			{
				Name: "implement",
				Role: "implementer",
			},
		},
	}

	engine := NewEngine(func(ctx context.Context, stage Stage, artifacts map[string]string) (StageResult, error) {
		return StageResult{
			Status: StageStatusSucceeded,
			Artifacts: map[string]string{
				stage.Name: stage.Name + "-done",
			},
		}, nil
	})

	results, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[1].StageName != "implement" {
		t.Fatalf("results[1].StageName = %q, want %q", results[1].StageName, "implement")
	}
	if results[1].Result.Artifacts["implement"] != "implement-done" {
		t.Fatalf("unexpected implement artifact: %+v", results[1].Result.Artifacts)
	}
}

func TestEngineExecuteFailureTransition(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{
				Name:  "test",
				Role:  "tester",
				Retry: 0,
				Transition: Transition{
					Failure: "fix",
				},
			},
			{
				Name: "fix",
				Role: "implementer",
			},
		},
	}

	engine := NewEngine(func(ctx context.Context, stage Stage, artifacts map[string]string) (StageResult, error) {
		if stage.Name == "test" {
			return StageResult{Status: StageStatusFailed}, nil
		}
		return StageResult{Status: StageStatusSucceeded}, nil
	})

	results, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[1].StageName != "fix" {
		t.Fatalf("results[1].StageName = %q, want %q", results[1].StageName, "fix")
	}
}

func TestEngineExecuteRetry(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{
				Name:  "plan",
				Role:  "planner",
				Retry: 1,
			},
		},
	}

	attempts := 0
	engine := NewEngine(func(ctx context.Context, stage Stage, artifacts map[string]string) (StageResult, error) {
		attempts++
		if attempts == 1 {
			return StageResult{Status: StageStatusFailed}, nil
		}
		return StageResult{Status: StageStatusSucceeded}, nil
	})

	results, err := engine.Execute(context.Background(), workflow)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if results[0].Attempt != 1 || results[1].Attempt != 2 {
		t.Fatalf("attempt sequence = %+v, want attempts 1 then 2", results)
	}
}

func TestEngineExecuteUnknownTransition(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{
				Name: "plan",
				Role: "planner",
				Transition: Transition{
					Success: "missing",
				},
			},
		},
	}

	engine := NewEngine(func(ctx context.Context, stage Stage, artifacts map[string]string) (StageResult, error) {
		return StageResult{Status: StageStatusSucceeded}, nil
	})

	_, err := engine.Execute(context.Background(), workflow)
	if err == nil || !strings.Contains(err.Error(), "references unknown success transition") {
		t.Fatalf("Execute() error = %v, want unknown success transition validation error", err)
	}
}

func TestEngineExecuteCycleProtection(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{
				Name: "loop",
				Role: "tester",
				Transition: Transition{
					Failure: "loop",
				},
			},
		},
	}

	engine := NewEngine(func(ctx context.Context, stage Stage, artifacts map[string]string) (StageResult, error) {
		return StageResult{Status: StageStatusFailed}, nil
	})

	_, err := engine.Execute(context.Background(), workflow)
	if err == nil || !strings.Contains(err.Error(), "max transitions") {
		t.Fatalf("Execute() error = %v, want max transitions error", err)
	}
}

func TestEngineExecuteRunnerError(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{Name: "plan", Role: "planner"},
		},
	}

	engine := NewEngine(func(ctx context.Context, stage Stage, artifacts map[string]string) (StageResult, error) {
		return StageResult{}, errors.New("boom")
	})

	_, err := engine.Execute(context.Background(), workflow)
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("Execute() error = %v, want runner error", err)
	}
}
