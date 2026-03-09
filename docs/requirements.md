# Saga v1 Requirements

## 1. 文書の目的

本書は `saga` の v1 実装に必要な機能要件、非機能要件、制約、受け入れ条件を定義する。  
`saga` は AI Agent Framework であり、`Codex CLI` を用いて複数のサブエージェントを実行し、GitHub Issue/PR と完全自動連携しながら開発タスクを自律的に完了させる。

## 2. プロダクトゴール

- GitHub 上の Issue を起点に、実装から PR 作成、CI 監視、修正、マージ、Issue 完了までを自動化する
- Go 単一バイナリとして配布し、WSL2 Ubuntu 上で `systemd` 管理のサービスとして安定稼働させる
- `Codex CLI` をサブエージェントの実行基盤として利用する
- worktree 分離、段階的権限、ログ・状態永続化により、安全性と再現性を確保する

## 3. 前提条件と制約

- 実行環境は WSL2 上の Ubuntu とする
- `systemd` が有効化されていることを前提にする
- `saga` 自身は Go 単一バイナリで提供する
- `Codex CLI` は外部依存としてインストール済みであること
- タスク対象リポジトリは WSL2 の Linux 側ファイルシステムに置く
- v1 で GitHub Issue/PR の完全自動連携を必須とする
- v1 で GitHub Webhook 受信は必須にしない
  - WSL2 のネットワーク制約を考慮し、v1 の自動連携は GitHub API のポーリングとリコンシリエーションを主方式とする

## 4. v1 のスコープ

### 4.1 対象

- 常駐 Orchestrator サービス
- YAML ベースのワークフロー実行
- サブエージェント実行
- GitHub Issue/PR 自動連携
- git worktree を用いた隔離実行
- ローカル状態管理と復旧
- systemd 管理、監視、ログ収集

### 4.2 v1 で対象外

- 複数ホストにまたがる分散実行
- Kubernetes などのクラスタ運用
- GitHub Webhook を前提とする外部公開サーバ
- IDE プラグインや GUI
- 人間の承認なしには越えられない branch protection の回避

## 5. 主要ユースケース

1. GitHub Issue に `saga:ready` ラベルが付与される
2. `saga` が Issue を検知し、要件を取得してタスク化する
3. Planner が実装計画を作成し、Issue に計画コメントを投稿する
4. Implementer が worktree 上で実装する
5. Tester/Reviewer が並列にテスト・レビューを行う
6. 不備があれば Implementer に戻り、fix loop を回す
7. 条件を満たしたら PR を自動作成または更新する
8. CI を監視し、失敗時は自動修正して再 push する
9. すべての条件を満たしたら PR を自動マージする
10. Issue に結果コメントを投稿し、Issue を完了状態に同期する

## 6. 機能要件

### 6.1 コア実行要件

| ID | 要件 |
|---|---|
| FR-CORE-01 | `saga` は Go 製の単一バイナリとしてビルドできること |
| FR-CORE-02 | `saga serve` は常駐サービスとして実行できること |
| FR-CORE-03 | `saga enqueue`, `status`, `cancel`, `retry`, `resume`, `doctor` の CLI を持つこと |
| FR-CORE-04 | daemon は Unix domain socket でローカル CLI と通信できること |
| FR-CORE-05 | daemon は SQLite に状態を永続化し、再起動後に復旧できること |

### 6.2 ワークフロー要件

| ID | 要件 |
|---|---|
| FR-WF-01 | ワークフローは YAML で定義できること |
| FR-WF-02 | stage ごとに `role`, `sandbox`, `network`, `timeout`, `retry`, `transition` を定義できること |
| FR-WF-03 | `parallel` stage をサポートすること |
| FR-WF-04 | `plan -> implement -> test/review -> verify -> complete` の基本経路を提供すること |
| FR-WF-05 | `test` または `review` が失敗した場合、`implement` に戻る fix loop をサポートすること |
| FR-WF-06 | stage 出力は次 stage に artifact として引き渡せること |
| FR-WF-07 | stage 実行結果に応じて `complete`, `failed`, `blocked`, `cancelled` を判定できること |

### 6.3 サブエージェント要件

| ID | 要件 |
|---|---|
| FR-AGENT-01 | v1 は少なくとも `planner`, `implementer`, `tester`, `reviewer`, `verifier` を提供すること |
| FR-AGENT-02 | `implementer` は編集可能な worktree で動作すること |
| FR-AGENT-03 | `reviewer` は原則 read-only で動作すること |
| FR-AGENT-04 | `tester` は build/test を実行できるが primary worktree を破壊しないこと |
| FR-AGENT-05 | サブエージェントは独立した Codex CLI プロセスとして起動されること |
| FR-AGENT-06 | サブエージェントごとに専用ログと結果ファイルを持つこと |

### 6.4 Codex CLI 連携要件

| ID | 要件 |
|---|---|
| FR-CODEX-01 | `saga` は `Codex CLI` を外部プロセスとして起動できること |
| FR-CODEX-02 | stage ごとに `read-only`, `workspace-write`, `full` 相当の sandbox を切り替えられること |
| FR-CODEX-03 | stage ごとに network access の有無を切り替えられること |
| FR-CODEX-04 | Codex CLI の標準出力・標準エラー・終了コードを取得し、結果判定に用いること |
| FR-CODEX-05 | Codex CLI の実行結果を `result.json` などの構造化 artifact に変換すること |
| FR-CODEX-06 | Codex CLI の path, model, env, timeout を設定ファイルから変更できること |
| FR-CODEX-07 | Codex CLI の thread/resume 機能に依存しなくても stage を実行できること |

### 6.5 Git worktree 要件

| ID | 要件 |
|---|---|
| FR-GIT-01 | タスクごとに primary worktree を作成すること |
| FR-GIT-02 | branch 名は一意に生成されること |
| FR-GIT-03 | validate 用に shadow worktree を作成できること |
| FR-GIT-04 | failed/cancelled/crashed 後に orphaned worktree を回収できること |
| FR-GIT-05 | 再試行時は既存 branch/worktree を再利用または再作成できること |

### 6.6 GitHub Issue/PR 完全自動連携要件

| ID | 要件 |
|---|---|
| FR-GH-01 | `saga` は設定済みリポジトリの open Issue を定期ポーリングで検出できること |
| FR-GH-02 | 取り込み対象 Issue をラベル、assignee、repository、comment command で判定できること |
| FR-GH-03 | 同一 Issue の二重処理を防ぐ排他制御を持つこと |
| FR-GH-04 | Issue 本文、コメント、ラベル、関連 PR 情報を取得できること |
| FR-GH-05 | 実行計画を Issue コメントとして自動投稿できること |
| FR-GH-06 | 実装開始、進行中、blocked、CI failure、完了などの状態を Issue/PR コメントに反映できること |
| FR-GH-07 | 実装結果をもとに PR を自動作成または既存 PR を更新できること |
| FR-GH-08 | PR title/body をテンプレートで生成できること |
| FR-GH-09 | push 後に PR の CI 状態、check run、workflow run、review 状態を監視できること |
| FR-GH-10 | CI 失敗時に fix loop を再実行し、同じ PR に追記 push できること |
| FR-GH-11 | repo 設定に従って `merge`, `squash`, `rebase` を選択して自動マージできること |
| FR-GH-12 | マージ後に Issue を閉じる、または closed 状態に同期できること |
| FR-GH-13 | daemon 再起動後に open PR / in-progress run を再同期できること |
| FR-GH-14 | API rate limit と一時障害に対して backoff と再試行を持つこと |

### 6.7 systemd / WSL2 要件

| ID | 要件 |
|---|---|
| FR-OPS-01 | `systemd` unit により自動起動できること |
| FR-OPS-02 | `Restart=on-failure` に耐えること |
| FR-OPS-03 | 停止時に子プロセスを適切に終了できること |
| FR-OPS-04 | WSL2 上で headless 実行できること |
| FR-OPS-05 | runtime/state/log を systemd 管理ディレクトリ配下に配置できること |

### 6.8 観測性要件

| ID | 要件 |
|---|---|
| FR-OBS-01 | run ごとに `events.ndjson` を保存すること |
| FR-OBS-02 | subagent ごとに stdout/stderr ログを保存すること |
| FR-OBS-03 | journald に主要イベントを出力すること |
| FR-OBS-04 | `saga status` で queue, run, worker, GitHub 同期状態を確認できること |
| FR-OBS-05 | `saga logs` で run と worker のログを参照できること |

## 7. 非機能要件

### 7.1 信頼性

- daemon が異常終了しても SQLite と artifact から復旧できること
- 子プロセスのハングを timeout で検知し、強制終了できること
- GitHub API の一時失敗やネットワークエラー時に再試行できること

### 7.2 性能

- ポーリング周期は設定可能であること
- 複数 task を並列処理できること
- 並列度は CPU/メモリ/リポジトリ特性に応じて設定可能であること

### 7.3 セキュリティ

- デフォルトは最小権限で動作すること
- `full` 権限は明示 opt-in にすること
- GitHub 認証情報と OpenAI/Codex 認証情報は環境変数または systemd credential で扱うこと
- `/mnt/c` 上の実行を非推奨として警告すること

### 7.4 可搬性

- Linux amd64/arm64 でビルドできること
- CGO に依存しないビルドを優先すること

## 8. リポジトリ側の前提条件

- 自動マージ対象リポジトリは、`saga` がマージ可能な権限設定を持つこと
- required review がある場合は、GitHub App に bypass 権限を与えるか、自動化対象ブランチの保護ルールを調整すること
- CI は PR ベースで実行され、GitHub API から結果取得可能であること
- branch naming, PR template, merge strategy を設定ファイルまたはリポジトリルールとして定義できること

## 9. v1 受け入れ条件

以下を満たしたとき v1 完了とみなす。

1. `systemd` で `saga serve` が自動起動する
2. 対象 Issue を自動検出して task 化できる
3. planner, implementer, tester, reviewer, verifier のワークフローが動作する
4. PR が自動作成される
5. CI failure 時に自動修正と再 push が実行される
6. 成功時に PR が自動マージされ、Issue が完了状態に同期される
7. daemon 再起動後に in-progress run を再同期できる
8. `status`, `logs`, `retry`, `cancel`, `resume` が運用可能である
