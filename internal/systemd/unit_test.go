package systemd

import (
	"strings"
	"testing"
)

func TestDefaultUnit(t *testing.T) {
	t.Parallel()

	unit, err := DefaultUnit("/usr/local/bin/saga serve --config /etc/saga/config.yaml")
	if err != nil {
		t.Fatalf("DefaultUnit() error = %v", err)
	}
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

func TestDefaultUnitRejectsNewline(t *testing.T) {
	t.Parallel()

	if _, err := DefaultUnit("/usr/local/bin/saga\nExecStart=/bin/evil"); err == nil {
		t.Fatal("DefaultUnit() error = nil, want validation error")
	}
}
