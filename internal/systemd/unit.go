package systemd

import (
	"fmt"
	"strings"
)

func DefaultUnit(execStart string) (string, error) {
	if strings.ContainsAny(execStart, "\r\n") {
		return "", fmt.Errorf("execStart must not contain newlines")
	}

	return fmt.Sprintf(`[Unit]
Description=Saga Orchestrator
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=%s
Restart=on-failure
RestartSec=5
EnvironmentFile=/etc/sg/sg.env
StateDirectory=sg
RuntimeDirectory=sg
LogsDirectory=sg
KillMode=control-group
TimeoutStopSec=180

[Install]
WantedBy=multi-user.target
`, execStart), nil
}
