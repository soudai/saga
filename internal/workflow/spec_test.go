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

	plan := workflow.Stages[0]
	if plan.Name != "plan" {
		t.Fatalf("plan.Name = %q, want %q", plan.Name, "plan")
	}
	if plan.Role != "planner" {
		t.Fatalf("plan.Role = %q, want %q", plan.Role, "planner")
	}
	if plan.Sandbox != "read-only" {
		t.Fatalf("plan.Sandbox = %q, want %q", plan.Sandbox, "read-only")
	}
	if plan.Network {
		t.Fatalf("plan.Network = %v, want false", plan.Network)
	}
	if plan.Timeout.String() != "30s" {
		t.Fatalf("plan.Timeout = %s, want 30s", plan.Timeout)
	}
	if plan.Retry != 1 {
		t.Fatalf("plan.Retry = %d, want 1", plan.Retry)
	}
	if plan.WorktreeMode != "none" {
		t.Fatalf("plan.WorktreeMode = %q, want %q", plan.WorktreeMode, "none")
	}
	if plan.Transition.Success != "implement" {
		t.Fatalf("plan.Transition.Success = %q, want %q", plan.Transition.Success, "implement")
	}

	implement := workflow.Stages[1]
	if implement.Name != "implement" {
		t.Fatalf("implement.Name = %q, want %q", implement.Name, "implement")
	}
	if implement.Role != "implementer" {
		t.Fatalf("implement.Role = %q, want %q", implement.Role, "implementer")
	}
	if implement.Sandbox != "workspace-write" {
		t.Fatalf("implement.Sandbox = %q, want %q", implement.Sandbox, "workspace-write")
	}
	if implement.Network {
		t.Fatalf("implement.Network = %v, want false", implement.Network)
	}
	if implement.Timeout.String() != "5m0s" && implement.Timeout.String() != "5m" {
		t.Fatalf("implement.Timeout = %s, want 5m", implement.Timeout)
	}
}

func TestParseWorkflowInvalidTimeout(t *testing.T) {
	t.Parallel()

	data := []byte(`
stages:
  - name: bad
    role: tester
    sandbox: read-only
    network: false
    timeout: not-a-duration
`)

	if _, err := Parse(data); err == nil {
		t.Fatal("Parse() error = nil, want invalid timeout error")
	}
}

func TestParseWorkflowRejectsInvalidTransition(t *testing.T) {
	t.Parallel()

	data := []byte(`
stages:
  - name: plan
    role: planner
    retry: -1
    transition:
      success: missing
`)

	if _, err := Parse(data); err == nil {
		t.Fatal("Parse() error = nil, want validation error")
	}
}
