# Saga

Saga は、Linux / WSL2 上でローカル常駐 orchestration を行う Go 製 AI agent framework です。最終的には Codex を使ったサブエージェントと GitHub 連携により、Issue 起点の自動開発フローを実現することを目指しています。現在の `main` には、ローカル daemon、SQLite ベースの control plane、workflow 関連ライブラリ、GitHub 連携ヘルパー、systemd 用 assets、release packaging 補助まで入っています。

## 現在のスコープ

`main` で end-to-end に動くもの:

- `init` / `enqueue` / `version` / `doctor` / `serve` / `status` / `cancel` / `retry` / `resume` の CLI
- Unix socket 上で動くローカル daemon
- SQLite ベースの task / run / lease 保存
- control API 経由の status 取得と task state 更新
- config 読み込み、validation、構造化 logging、systemd readiness notify
- systemd unit 生成ヘルパー、sample config、smoke test ドキュメント、release packaging script

テスト済みの package としては存在するが、まだ完全自律 daemon loop には接続されていないもの:

- Codex runner と artifact 保存
- Git worktree manager
- Workflow parser / execution engine
- GitHub issue client / selector
- Recovery policy / reconciliation helper

## 利用可能なコマンド

```bash
sg version
sg init
sg issue draft <repository> --from-file task.md
sg issue create <repository> --from-file task.md
sg enqueue <repository> <issue-number> --config /path/to/config.yaml
sg doctor --config /path/to/config.yaml
sg serve --config /path/to/config.yaml
sg status --config /path/to/config.yaml
sg cancel <task-id> --config /path/to/config.yaml
sg retry <task-id> --config /path/to/config.yaml
sg resume <task-id> --config /path/to/config.yaml
```

補足:

- `sg init` は対話形式で config file を生成し、project-local / system-wide の初期値を選べます
- `sg issue draft/create` は markdown の指示書ソースから task 用 GitHub Issue を下書きまたは作成します
- `sg enqueue <repository> <issue-number>` は daemon の control API 経由で `queued` task を登録します
- `sg serve` は `SIGINT` または `SIGTERM` を受けるまで foreground で動作します
- `sg status` と task action は設定された Unix socket 経由で daemon に接続します
- `sg issue create` を使うには GitHub CLI (`gh`) の install と認証が必要です

## ランタイム構成

現在の runtime 挙動:

- state、run file、log、socket 親ディレクトリを作成する
- 設定された state directory 配下に SQLite database を作成して利用する
- Unix domain socket 上でローカル HTTP control API を公開する
- socket file はローカル利用向けの権限に制限する
- `slog` で起動と停止を記録する
- `NOTIFY_SOCKET` があれば systemd に `READY=1` を送る

現在 daemon が公開している control plane endpoint:

- `GET /status`
- `POST /tasks`
- `POST /tasks/{id}/cancel`
- `POST /tasks/{id}/retry`
- `POST /tasks/{id}/resume`

## 設定

Saga は YAML config を読み込み、その後で環境変数 override を適用します。

設定項目:

- `runtime.state_dir`
- `runtime.run_dir`
- `runtime.log_dir`
- `server.socket_path`
- `log.level`

環境変数 override:

- `SAGA_STATE_DIR`
- `SAGA_RUN_DIR`
- `SAGA_LOG_DIR`
- `SAGA_SOCKET_PATH`
- `SAGA_LOG_LEVEL`

runtime path はすべて絶対パスである必要があります。

sample config:

- [`config/samples/minimal.config.yaml`](./config/samples/minimal.config.yaml)
- [`config/samples/production.config.yaml`](./config/samples/production.config.yaml)

## クイックスタート

binary を build します。

```bash
make build
```

まず対話形式で config を生成できます。

```bash
./bin/sg init
```

出力先を明示することもできます。

```bash
./bin/sg init ./sg.local.yaml
```

このコマンドは profile を選ばせたあと、各 path や log level を確認・編集できます。

手動で書く場合は、`./sg.local.yaml` のような config でも動きます。

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

1 つ目のターミナルで daemon を起動します。

```bash
./bin/sg serve --config ./sg.local.yaml
```

別ターミナルから確認します。

```bash
./bin/sg enqueue soudai/saga 123 --config ./sg.local.yaml
./bin/sg doctor --config ./sg.local.yaml
./bin/sg status --config ./sg.local.yaml
```

ローカルの markdown から task 指示書 Issue を GitHub に作る例:

```bash
cat > task.md <<'EOF'
GitHub Issue を enqueue できる flow を実装する。
EOF

gh auth login
./bin/sg issue draft soudai/saga --from-file task.md
./bin/sg issue create soudai/saga --from-file task.md
./bin/sg issue create soudai/saga --from-file task.md --enqueue --config ./sg.local.yaml
```

## 実装済み building block

`main` には次の package 実装が入り、テストもあります。

- SQLite store と lease 制御: [`internal/store/`](./internal/store)
- Unix socket control client/server: [`internal/control/`](./internal/control)
- Codex 実行と artifact 永続化: [`internal/codex/`](./internal/codex), [`internal/artifact/`](./internal/artifact)
- Git worktree 管理: [`internal/gitops/`](./internal/gitops)
- Workflow parse / execution: [`internal/workflow/`](./internal/workflow)
- GitHub issue 一覧取得と selector: [`internal/github/`](./internal/github)
- Recovery policy: [`internal/recovery/`](./internal/recovery)
- systemd unit helper: [`internal/systemd/`](./internal/systemd)

## 開発

必要要件:

- Go `1.26`
- Linux または WSL2 推奨

主なコマンド:

```bash
make fmt
make test
./ci/test.sh
./ci/release.sh v0.1.0
```

release script は `dist/` 配下に配布用 tarball を生成します。

## 運用関連

関連ファイル:

- systemd service template: [`contrib/systemd/sg.service`](./contrib/systemd/sg.service)
- systemd / WSL2 メモ: [`docs/systemd-wsl2.md`](./docs/systemd-wsl2.md)
- smoke test: [`docs/testing/smoke-test.md`](./docs/testing/smoke-test.md)

## ロードマップと設計資料

フェーズごとの実装計画は [`docs/execution/`](./docs/execution/) にあります。

主要ドキュメント:

- [`docs/README.md`](./docs/README.md)
- [`docs/requirements.md`](./docs/requirements.md)
- [`docs/architecture.md`](./docs/architecture.md)
- [`docs/github-integration.md`](./docs/github-integration.md)
- [`docs/implementation-plan.md`](./docs/implementation-plan.md)
