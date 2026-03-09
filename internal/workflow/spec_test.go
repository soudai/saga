package workflow

import "testing"

func TestParseWorkflow(t *testing.T) {
	t.Parallel()

	data := []byte(`
stages:
  - name: plan
    role: planner
    sandbox: read-only
    network: false
    timeout: 30s
    retry: 1
    worktree_mode: none
    transition:
      success: implement
  - name: implement
    role: implementer
    sandbox: workspace-write
    network: false
    timeout: 5m
`)

	workflow, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(workflow.Stages) != 2 {
		t.Fatalf("len(stages) = %d, want 2", len(workflow.Stages))
	}
	if workflow.Stages[0].Timeout.String() != "30s" {
		t.Fatalf("timeout = %s, want 30s", workflow.Stages[0].Timeout)
	}
}
