package systemd

import (
	"strings"
	"testing"
)

func TestDefaultUnit(t *testing.T) {
	t.Parallel()

	unit := DefaultUnit("/usr/local/bin/saga serve --config /etc/saga/config.yaml")
	for _, want := range []string{
		"Type=notify",
		"Restart=on-failure",
		"StateDirectory=saga",
		"ExecStart=/usr/local/bin/saga serve --config /etc/saga/config.yaml",
	} {
		if !strings.Contains(unit, want) {
			t.Fatalf("unit does not contain %q", want)
		}
	}
}
