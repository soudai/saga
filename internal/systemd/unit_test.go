package systemd

import (
	"strings"
	"testing"
)

func TestDefaultUnit(t *testing.T) {
	t.Parallel()

	unit, err := DefaultUnit("/usr/local/bin/sg serve --config /etc/sg/config.yaml")
	if err != nil {
		t.Fatalf("DefaultUnit() error = %v", err)
	}
	for _, want := range []string{
		"Type=notify",
		"Restart=on-failure",
		"StateDirectory=sg",
		"ExecStart=/usr/local/bin/sg serve --config /etc/sg/config.yaml",
	} {
		if !strings.Contains(unit, want) {
			t.Fatalf("unit does not contain %q", want)
		}
	}
}

func TestDefaultUnitRejectsNewline(t *testing.T) {
	t.Parallel()

	if _, err := DefaultUnit("/usr/local/bin/sg\nExecStart=/bin/evil"); err == nil {
		t.Fatal("DefaultUnit() error = nil, want validation error")
	}
}
