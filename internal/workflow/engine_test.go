package workflow

import (
	"context"
	"testing"
)

func TestEngineExecute(t *testing.T) {
	t.Parallel()

	workflow := Workflow{
		Stages: []Stage{
			{
				Name: "plan",
				Transition: Transition{
					Success: "implement",
				},
			},
			{
				Name: "implement",
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
	if results["implement"].Artifacts["implement"] != "implement-done" {
		t.Fatalf("unexpected implement artifact: %+v", results["implement"].Artifacts)
	}
}
