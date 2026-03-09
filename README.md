# Saga

Saga is a Go-based AI agent framework intended to run as a long-lived daemon on Linux or WSL2 and eventually automate GitHub-driven development workflows with Codex-powered subagents.

## Current status

The repository currently contains the Phase 0-1 baseline:

- Go module, project layout, and build/test entrypoints
- CLI commands: `saga version`, `saga doctor`, and `saga serve`
- YAML config loading with environment variable overrides
- Validation for absolute runtime paths and socket path
- Structured logging via `slog`
- Daemon skeleton with systemd readiness notification when available

The later phases described in [`docs/execution/`](./docs/execution/) are still documentation and execution plans. GitHub automation, workflow execution, SQLite state management, and worktree orchestration are not on `main` yet.

## What works today

### CLI

- `saga version` prints build metadata
- `saga doctor` validates the loaded configuration and reports basic runtime checks
- `saga serve` loads config, creates runtime directories, initializes logging, and waits as a long-running daemon process

### Configuration

Saga loads configuration from a YAML file and then applies environment overrides:

- `SAGA_STATE_DIR`
- `SAGA_RUN_DIR`
- `SAGA_LOG_DIR`
- `SAGA_SOCKET_PATH`
- `SAGA_LOG_LEVEL`

All runtime paths must be absolute.

### Runtime behavior

- Creates runtime directories for state, run files, logs, and the socket parent directory
- Logs startup and shutdown events
- Sends `READY=1` through systemd notify when the environment supports it
- Continues running even if systemd notification is unavailable

## Requirements

- Go `1.26`
- Linux or WSL2 recommended
- `systemd` is optional for the current implementation, but it is the target deployment model

## Quick start

Build the binary:

```bash
make build
```

Create a local config for non-root testing:

```yaml
runtime:
  state_dir: /tmp/saga/state
  run_dir: /tmp/saga/run
  log_dir: /tmp/saga/log
server:
  socket_path: /tmp/saga/run/saga.sock
log:
  level: info
```

Run the available commands:

```bash
./bin/saga version
./bin/saga doctor --config ./saga.local.yaml
./bin/saga serve --config ./saga.local.yaml
```

`saga serve` runs in the foreground until it receives `SIGINT` or `SIGTERM`.

## Development

```bash
make fmt
make test
```

The CI helper script is available at [`ci/test.sh`](./ci/test.sh).

## Roadmap

The planned implementation is organized into phases:

- Phase 0: Repository bootstrap
- Phase 1: Runtime foundation
- Phase 2: State and control plane
- Phase 3: Codex execution layer
- Phase 4: Git worktree layer
- Phase 5: Workflow engine
- Phase 6: GitHub automation
- Phase 7: Validation and recovery
- Phase 8: systemd hardening
- Phase 9: Release readiness

See the following documents for the design and execution plan:

- [`docs/README.md`](./docs/README.md)
- [`docs/requirements.md`](./docs/requirements.md)
- [`docs/architecture.md`](./docs/architecture.md)
- [`docs/implementation-plan.md`](./docs/implementation-plan.md)
- [`docs/systemd-wsl2.md`](./docs/systemd-wsl2.md)
