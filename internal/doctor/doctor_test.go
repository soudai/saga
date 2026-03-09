package doctor

import (
	"testing"

	"github.com/soudai/saga/internal/config"
)

func TestRunIncludesOperationalChecks(t *testing.T) {
	t.Parallel()

	cfg := config.Default()
	checks := Run(cfg)
	if len(checks) < 6 {
		t.Fatalf("len(checks) = %d, want >= 6", len(checks))
	}
}
