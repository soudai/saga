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
EnvironmentFile=/etc/saga/saga.env
StateDirectory=saga
RuntimeDirectory=saga
LogsDirectory=saga
KillMode=control-group
TimeoutStopSec=180

[Install]
WantedBy=multi-user.target
`, execStart), nil
}
