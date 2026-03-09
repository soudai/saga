package doctor

import (
	"testing"

	"github.com/soudai/saga/internal/config"
)

func TestRunIncludesOperationalChecks(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	checks := Run(cfg)

	wantNames := map[string]bool{
		"config validation":               false,
		"socket path":                     false,
		"runtime state dir":               false,
		"runtime run dir":                 false,
		"runtime log dir":                 false,
		"systemd available":               false,
		"wsl2 environment":                false,
		"linux filesystem recommendation": false,
	}

	for _, check := range checks {
		if _, ok := wantNames[check.Name]; ok {
			wantNames[check.Name] = true
		}
	}

	for name, seen := range wantNames {
		if !seen {
			t.Fatalf("missing check %q", name)
		}
	}
}
