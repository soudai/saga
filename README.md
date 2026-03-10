# Saga

Saga is a Go-based AI agent framework for long-running local orchestration on Linux and WSL2. The long-term goal is fully automated GitHub-driven development with Codex-powered subagents. The current `main` branch already includes the local daemon, SQLite-backed control plane, workflow-related libraries, GitHub integration helpers, systemd assets, and release packaging helpers.

## Current scope

What is end-to-end on `main`:

- CLI commands for `init`, `enqueue`, `version`, `doctor`, `serve`, `status`, `cancel`, `retry`, and `resume`
- A local daemon that opens a Unix socket control plane
- SQLite-backed task/run/lease storage
- Local status reporting and task state updates through the control API
- Config loading, validation, structured logging, and systemd readiness notification
- systemd unit generation helpers, sample configs, smoke-test docs, and a release packaging script

What is implemented as tested packages but not yet wired into a full autonomous daemon loop:

- Codex runner and artifact storage
- Git worktree manager
- Workflow parser and execution engine
- GitHub issue client and selector
- Recovery policy and reconciliation helpers

## Available commands

```bash
sg version
sg init
sg enqueue <repository> <issue-number> --config /path/to/config.yaml
sg doctor --config /path/to/config.yaml
sg serve --config /path/to/config.yaml
sg status --config /path/to/config.yaml
sg cancel <task-id> --config /path/to/config.yaml
sg retry <task-id> --config /path/to/config.yaml
sg resume <task-id> --config /path/to/config.yaml
```

Notes:

- `sg init` interactively creates a config file with project-local or system-wide defaults
- `sg enqueue <repository> <issue-number>` registers a queued task through the daemon control API
- `sg serve` runs in the foreground until it receives `SIGINT` or `SIGTERM`
- `sg status` and task actions talk to the daemon over the configured Unix socket

## Runtime architecture

Current runtime behavior:

- Creates runtime directories for state, run files, logs, and the socket parent directory
- Creates and uses a SQLite database under the configured state directory
- Serves a local HTTP control API over a Unix domain socket
- Restricts the socket file to local access
- Logs startup and shutdown with `slog`
- Sends `READY=1` to systemd when `NOTIFY_SOCKET` is available

Control plane endpoints currently exposed by the daemon:

- `GET /status`
- `POST /tasks`
- `POST /tasks/{id}/cancel`
- `POST /tasks/{id}/retry`
- `POST /tasks/{id}/resume`

## Configuration

Saga loads a YAML config file and then applies environment variable overrides.

Config fields:

- `runtime.state_dir`
- `runtime.run_dir`
- `runtime.log_dir`
- `server.socket_path`
- `log.level`

Environment overrides:

- `SAGA_STATE_DIR`
- `SAGA_RUN_DIR`
- `SAGA_LOG_DIR`
- `SAGA_SOCKET_PATH`
- `SAGA_LOG_LEVEL`

All runtime paths must be absolute.

Sample configs are available in:

- [`config/samples/minimal.config.yaml`](./config/samples/minimal.config.yaml)
- [`config/samples/production.config.yaml`](./config/samples/production.config.yaml)

## Quick start

Build the binary:

```bash
make build
```

Generate a config interactively:

```bash
./bin/sg init
```

Or choose an explicit output path:

```bash
./bin/sg init ./sg.local.yaml
```

The command asks for a profile and then lets you confirm or edit each path.

If you prefer to create a file manually, a local config such as `./sg.local.yaml` also works:

```yaml
runtime:
  state_dir: /tmp/sg/state
  run_dir: /tmp/sg/run
  log_dir: /tmp/sg/log

server:
  socket_path: /tmp/sg/run/sg.sock

log:
  level: info
```

Run the daemon in one terminal:

```bash
./bin/sg serve --config ./sg.local.yaml
```

Query it from another terminal:

```bash
./bin/sg enqueue soudai/saga 123 --config ./sg.local.yaml
./bin/sg doctor --config ./sg.local.yaml
./bin/sg status --config ./sg.local.yaml
```

## Implemented building blocks

These packages are already present on `main` and covered by tests:

- SQLite store and lease control in [`internal/store/`](./internal/store)
- Unix socket control client/server in [`internal/control/`](./internal/control)
- Codex execution and artifact persistence in [`internal/codex/`](./internal/codex) and [`internal/artifact/`](./internal/artifact)
- Git worktree management in [`internal/gitops/`](./internal/gitops)
- Workflow parsing and execution in [`internal/workflow/`](./internal/workflow)
- GitHub issue listing and selector logic in [`internal/github/`](./internal/github)
- Recovery policies in [`internal/recovery/`](./internal/recovery)
- systemd unit helpers in [`internal/systemd/`](./internal/systemd)

## Development

Requirements:

- Go `1.26`
- Linux or WSL2 recommended

Useful commands:

```bash
make fmt
make test
./ci/test.sh
./ci/release.sh v0.1.0
```

The release script builds a distributable tarball under `dist/`.

## Operations

Relevant operational assets:

- systemd service template: [`contrib/systemd/sg.service`](./contrib/systemd/sg.service)
- systemd/WSL2 notes: [`docs/systemd-wsl2.md`](./docs/systemd-wsl2.md)
- smoke test: [`docs/testing/smoke-test.md`](./docs/testing/smoke-test.md)

## Roadmap and design docs

The implementation plan is documented phase by phase under [`docs/execution/`](./docs/execution/).

Key references:

- [`docs/README.md`](./docs/README.md)
- [`docs/requirements.md`](./docs/requirements.md)
- [`docs/architecture.md`](./docs/architecture.md)
- [`docs/github-integration.md`](./docs/github-integration.md)
- [`docs/implementation-plan.md`](./docs/implementation-plan.md)
