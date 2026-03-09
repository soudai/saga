# Saga

Saga は、Linux / WSL2 上で常駐 daemon として動作し、将来的には Codex を使ったサブエージェント実行と GitHub 連携の自動開発フローを実現することを目指した Go 製の AI agent framework です。

## 現在の状態

現在の `main` に入っているのは Phase 0-1 の基盤部分です。

- Go module、基本ディレクトリ構成、build/test エントリポイント
- CLI コマンド: `saga version`、`saga doctor`、`saga serve`
- YAML config 読み込みと環境変数 override
- runtime path / socket path の絶対パス validation
- `slog` による構造化 logging
- systemd readiness notify に対応した daemon skeleton

[`docs/execution/`](./docs/execution/) に書かれている後続 phase は、現時点では実装ではなく計画ドキュメントです。GitHub 自動化、workflow 実行、SQLite による状態管理、worktree orchestration はまだ `main` には入っていません。

## 現在できること

### CLI

- `saga version` で build metadata を表示する
- `saga doctor` で読み込んだ設定と基本 runtime 条件を確認する
- `saga serve` で config を読み込み、runtime directory を作成し、長時間動作する daemon process として待機する

### 設定

Saga は YAML ファイルを読み込み、その後で次の環境変数 override を適用します。

- `SAGA_STATE_DIR`
- `SAGA_RUN_DIR`
- `SAGA_LOG_DIR`
- `SAGA_SOCKET_PATH`
- `SAGA_LOG_LEVEL`

runtime path はすべて絶対パスである必要があります。

### 実行時挙動

- state、run file、log、socket 親ディレクトリを作成する
- 起動時と終了時に log を出力する
- 環境が対応していれば systemd notify で `READY=1` を送る
- systemd notify が使えなくても継続動作する

## 必要要件

- Go `1.26`
- Linux または WSL2 推奨
- 現状実装では `systemd` は必須ではないが、運用ターゲットは `systemd` 前提

## クイックスタート

まず binary を build します。

```bash
make build
```

root 権限なしで試すためのローカル config 例です。

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

利用可能なコマンドは次のとおりです。

```bash
./bin/saga version
./bin/saga doctor --config ./saga.local.yaml
./bin/saga serve --config ./saga.local.yaml
```

`saga serve` は `SIGINT` または `SIGTERM` を受けるまで foreground で動作します。

## 開発

```bash
make fmt
make test
```

CI 用の補助スクリプトは [`ci/test.sh`](./ci/test.sh) にあります。

## ロードマップ

実装計画は次の phase に分かれています。

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

設計と実装計画は次のドキュメントを参照してください。

- [`docs/README.md`](./docs/README.md)
- [`docs/requirements.md`](./docs/requirements.md)
- [`docs/architecture.md`](./docs/architecture.md)
- [`docs/implementation-plan.md`](./docs/implementation-plan.md)
- [`docs/systemd-wsl2.md`](./docs/systemd-wsl2.md)
