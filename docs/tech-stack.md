# Saga Tech Stack

## 1. 技術選定方針

- Go 単一バイナリを最優先にする
- 依存は必要最小限に抑え、運用時の外部ランタイムを増やさない
- GitHub 連携は CLI ラッパーではなく API client を優先する
- git 操作は互換性を優先して `git` CLI を利用する
- SQLite は pure Go 実装を優先し、CGO 依存を避ける

## 2. 採用技術スタック

| 領域 | 採用候補 | 用途 | 採用理由 |
|---|---|---|---|
| 言語 | Go 1.24 系を推奨 | 本体実装 | 単一バイナリ、並行処理、配布容易性 |
| CLI | `cobra` | `saga` コマンド群 | サブコマンド設計がしやすい |
| 設定 | `koanf` + `yaml.v3` | config 読み込み | YAML + env + file merge がしやすい |
| ログ | `log/slog` | 構造化ログ | 標準ライブラリで十分 |
| systemd 通知 | `github.com/coreos/go-systemd/v22/daemon` | `sd_notify` | `Type=notify` 対応 |
| DB | `modernc.org/sqlite` | 状態永続化 | pure Go で単一バイナリ向き |
| GitHub API | `github.com/google/go-github` + `oauth2` | Issue/PR/Checks 操作 | `gh` 非依存で自動化しやすい |
| git 実行 | `os/exec` + `git` CLI | worktree/branch/push | 実運用での互換性が高い |
| HTTP/IPC | 標準 `net/http` + Unix socket | daemon/CLI 間制御 | 依存を増やさない |
| テンプレート | `text/template` | prompt / PR body / comments | 標準で十分 |
| テスト | `testing`, `httptest`, `testify` | unit/integration | 標準中心で必要箇所のみ補助 |
| lint/静的解析 | `gofmt`, `go test`, `govulncheck`, `golangci-lint` | 品質管理 | Go 標準運用に沿う |
| リリース | `goreleaser` | binary 配布 | Linux 向け配布を簡素化 |

## 3. 外部依存

### 3.1 必須

- `Codex CLI`
- `git`
- `systemd`
- GitHub 認証情報
- OpenAI/Codex 認証情報

### 3.2 任意

- `gh`
  - v1 の本体動作には必須としない
  - 運用者向け補助コマンドで使う余地はある

## 4. 開発要件

### 4.1 開発環境

- WSL2 Ubuntu
- Go toolchain
- `make` または `just`
- `git`
- `Codex CLI`

### 4.2 ビルド要件

- Linux amd64/arm64 向けビルド
- CGO 非依存ビルドを優先
- バージョン情報を埋め込めること

### 4.3 実行要件

- `systemd=true` な WSL2
- `/usr/local/bin/sg` に配置可能
- `/etc/sg/` に config と secrets を配置可能
- `/var/lib/sg`, `/run/sg`, `/var/log/sg` に書き込み可能

### 4.4 リポジトリ要件

- 対象 repository は `git worktree` を使えること
- CI が PR ベースで動作していること
- Saga 用 GitHub credential で push/PR/merge できること

## 5. 推奨設定ファイル

### 5.1 `/etc/sg/config.yaml`

- daemon 全体設定
- GitHub repositories
- workflow defaults
- concurrency
- poll intervals

### 5.2 `/etc/sg/sg.env`

- `GITHUB_APP_ID`
- `GITHUB_INSTALLATION_ID`
- `GITHUB_PRIVATE_KEY_FILE`
- `OPENAI_API_KEY`
- `CODEX_PATH`

### 5.3 `.sg/`

プロジェクトローカル override:

- `.sg/workflows/*.yaml`
- `.sg/prompts/*.md`
- `.sg/policies/*.md`
- `.sg/config.yaml`

## 6. 運用要件

### 6.1 systemd unit

最低要件:

- `Type=notify`
- `Restart=on-failure`
- `KillMode=control-group`
- `StateDirectory=sg`
- `RuntimeDirectory=sg`
- `LogsDirectory=sg`

### 6.2 ログ

- journald に構造化ログ出力
- run artifact に NDJSON trace を保存
- worker ごとの stdout/stderr を保存

### 6.3 セキュリティ

- `workspace-write` を既定とし、`full` は明示指定
- secrets は config 本文に直書きしない
- GitHub token の scope を最小化する

## 7. テスト戦略

### 7.1 Unit Test

- workflow parser
- state machine
- GitHub sync decision logic
- Codex adapter result parsing
- lease / lock logic

### 7.2 Integration Test

- mock GitHub API
- mock Codex binary
- real git worktree
- SQLite state recovery

### 7.3 Smoke Test

- WSL2 Ubuntu
- `systemctl start sg`
- Issue 検出から PR 作成まで

## 8. 採用しない技術

### 8.1 `go-git`

不採用理由:

- `git worktree` と実リポジトリの運用互換性で `git` CLI が有利

### 8.2 外部メッセージブローカー

不採用理由:

- v1 は単機能 daemon で十分
- SQLite + goroutine + local IPC で要件を満たせる

### 8.3 Webhook 前提設計

不採用理由:

- WSL2 ローカル常駐を前提とすると受信公開が複雑
- v1 は polling + reconciliation で実現する方が堅実
