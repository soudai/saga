# Saga Implementation Plan

## 1. 目的

本書は `saga` v1 を実装するための段階的な開発ステップを示す。  
ゴールは、WSL2 Ubuntu + systemd 上で常駐する Go 製 daemon が、GitHub Issue/PR と完全自動連携しつつ、Codex CLI を使ってサブエージェントを実行できる状態に到達することである。

## 2. フェーズ一覧

| Phase | 名称 | 主な成果物 |
|---|---|---|
| 0 | Repository Bootstrap | Go module, docs, basic layout |
| 1 | Runtime Foundation | CLI, config, logging, daemon skeleton |
| 2 | State & Control Plane | SQLite schema, Unix socket API, task lifecycle |
| 3 | Codex Execution Layer | Codex adapter, timeout, artifacts |
| 4 | Git Worktree Layer | worktree manager, cleanup, retry safety |
| 5 | Workflow Engine | YAML stages, transitions, parallel/fix loop |
| 6 | GitHub Automation | issue intake, comments, PR create/update/merge |
| 7 | Validation & Recovery | CI watch, reconcile, restart recovery |
| 8 | systemd Hardening | unit, doctor, shutdown, smoke tests |
| 9 | Release Readiness | packaging, docs sync, v1 acceptance |

## 3. フェーズ詳細

### Phase 0: Repository Bootstrap

目的:

- 実装の土台を作る

作業:

- Go module 初期化
- `cmd/sg`, `internal/`, `docs/`, `test/` 構成作成
- `Makefile` または `justfile` 追加
- CI のひな形追加

完了条件:

- `go test ./...` が空でも通る
- `sg version` が動く

### Phase 1: Runtime Foundation

目的:

- daemon と CLI の骨格を作る

作業:

- `cobra` で CLI 実装
- `sg serve`
- `sg doctor`
- config loader
- `slog` 初期化
- systemd readiness notify

完了条件:

- `systemd` なしでも `sg serve` が起動する
- config 読み込みと validation が動く

### Phase 2: State & Control Plane

目的:

- 永続状態と local control API を作る

作業:

- SQLite schema 定義
- repository/task/run/subagent テーブル作成
- lease テーブル作成
- Unix socket API 実装
- `status/cancel/retry/resume` API 実装

完了条件:

- task を DB に登録できる
- CLI から daemon の状態を取得できる

### Phase 3: Codex Execution Layer

目的:

- Codex CLI を安全に呼び出せるようにする

作業:

- Codex adapter 実装
- sandbox/network/model 設定適用
- stdout/stderr/result 保存
- timeout/cancel/kill 実装
- mock Codex で integration test

完了条件:

- 単一 stage を Codex で実行できる
- timeout と cancel が正しく動く

### Phase 4: Git Worktree Layer

目的:

- task ごとの隔離実行を成立させる

作業:

- primary worktree 作成
- branch 命名
- retry/recreate ロジック
- startup orphan cleanup
- shadow worktree 生成

完了条件:

- task ごとに独立 worktree を作れる
- cleanup の自動テストが通る

### Phase 5: Workflow Engine

目的:

- `takt` 的な宣言的 workflow を実行可能にする

作業:

- YAML schema 定義
- stage transition 実装
- parallel stage 実装
- artifact handoff 実装
- fix loop 実装
- 基本 workflow の builtin 追加

完了条件:

- `plan -> implement -> test/review -> verify -> complete` が動く
- validation failure で implement に戻れる

### Phase 6: GitHub Automation

目的:

- GitHub Issue/PR 完全自動連携を成立させる

作業:

- repository poller 実装
- issue selector 実装
- issue lease 実装
- plan/progress/result comment 実装
- PR create/update 実装
- check/status/review polling 実装
- merge 実装
- issue close/sync 実装

完了条件:

- open Issue を自動検出して PR まで作れる
- success 時に PR merge と Issue close まで進む

### Phase 7: Validation & Recovery

目的:

- 実運用の中断と再開に耐えるようにする

作業:

- startup reconciliation
- stale task detection
- open PR からの状態再構築
- failed CI からの fix loop 自動化
- resume/retry policy 実装

完了条件:

- daemon 再起動後に in-progress task を再同期できる
- CI failure からの自動再試行が動く

### Phase 8: systemd Hardening

目的:

- WSL2/systemd 常駐運用を固める

作業:

- unit file 作成
- install 手順作成
- `doctor` の systemd/WSL2 チェック追加
- graceful shutdown 実装
- journald 出力確認

完了条件:

- `systemctl enable --now sg` で常駐できる
- 再起動、停止、異常終了時の挙動が確認できる

### Phase 9: Release Readiness

目的:

- v1 を出荷可能にする

作業:

- smoke test
- docs 最終更新
- sample config
- release pipeline
- versioning and changelog

完了条件:

- v1 acceptance criteria を満たす
- binary を配布できる

## 4. 推奨ディレクトリ構成

```text
cmd/
  saga/
internal/
  app/
  daemon/
  config/
  control/
  store/
  workflow/
  codex/
  github/
  gitops/
  runtime/
  artifact/
  systemd/
pkg/
test/
docs/
```

## 5. 実装順の優先度

1. daemon skeleton
2. state store
3. Codex adapter
4. worktree manager
5. workflow engine
6. GitHub issue intake
7. PR automation
8. CI watch and merge
9. recovery/systemd hardening

## 6. Definition of Done

### コードレベル

- unit/integration test が追加されている
- lint と format が通る
- 主要構造体と interface に最低限のコメントがある

### 機能レベル

- task 起票から merge までの happy path が動く
- CI failure path が動く
- restart recovery が動く

### 運用レベル

- WSL2 Ubuntu で `systemd` 管理のサービスとして起動する
- ログ、状態、artifact を確認できる
- `doctor` で主要依存を検査できる

## 7. リスクと対策

| リスク | 内容 | 対策 |
|---|---|---|
| Codex CLI interface change | CLI flag や出力形式が変わる | adapter 層に閉じ込める |
| GitHub rate limit | polling が多すぎる | backoff と差分同期 |
| WSL2 file performance | `/mnt/c` 上で遅い | Linux 側 FS を必須推奨 |
| orphaned child process | daemon crash 時に残る | cgroup, timeout, startup cleanup |
| branch protection mismatch | auto-merge できない | doctor と preflight で検出 |
