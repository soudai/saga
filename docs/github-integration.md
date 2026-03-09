# GitHub Integration Specification

## 1. 目的

本書は `saga` v1 における GitHub Issue/PR 完全自動連携の仕様を定義する。  
v1 では、Issue の検出、進捗通知、PR 作成、CI 監視、fix loop、マージ、Issue 完了同期までを自動化対象とする。

## 2. 基本方針

- v1 の主方式は GitHub API の polling + reconciliation
- inbound webhook は必須にしない
- 実行単位は `repository + issue number`
- 同一 Issue の同時実行は禁止
- PR は task に 1 本を基本とし、fix loop は同じ PR に積み増す

## 3. 認証モデル

### 3.1 推奨

- GitHub App

理由:

- repository 単位で権限を最小化しやすい
- Issue/PR/Checks/Contents/Metadata をまとめて扱える
- auto-merge や branch protection との整合を取りやすい

### 3.2 代替

- Personal Access Token

用途:

- 単一ユーザーの PoC
- GitHub App の準備前段階

制約:

- 権限過多になりやすい
- 組織ポリシーで制限されやすい

## 4. 必要権限

### 4.1 GitHub App 権限

- `Issues`: Read and write
- `Pull requests`: Read and write
- `Contents`: Read and write
- `Metadata`: Read-only
- `Checks`: Read-only
- `Commit statuses`: Read-only
- `Actions`: Read-only

必要に応じて:

- `Administration`: branch protection bypass を使う場合のみ

### 4.2 リポジトリ運用前提

- Saga が push できること
- Saga が PR を作成できること
- Saga がマージできること
- auto-merge を使う場合、対象ブランチのルールがそれを許容すること

## 5. 自動取り込みルール

### 5.1 対象 repository

設定ファイルに列挙された repository のみ対象にする。

### 5.2 対象 Issue 判定

v1 では以下のいずれかで task 化する。

1. `saga:ready` ラベルが付いている
2. `assignee == saga-bot` である
3. `/saga run` コメントが追加された

追加条件:

- Issue は open であること
- `saga:blocked` ラベルがないこと
- 既に active task が存在しないこと

## 6. Issue ライフサイクル

### 6.1 取り込み時

実施内容:

- Issue 本文、コメント、ラベル、assignee を取得
- 既存 PR の有無を確認
- ローカル DB に task を登録
- lease を取得して重複実行を防止

### 6.2 計画フェーズ

Planner 実行後に Issue コメントを投稿する。

投稿内容:

- 実装方針
- 想定ステップ
- リスク
- 追加確認が必要な点

### 6.3 進行中同期

Issue コメントまたは label で以下を反映する。

- `running`
- `blocked`
- `waiting-ci`
- `retrying`
- `completed`
- `failed`

### 6.4 完了同期

成功時:

- 結果コメントを投稿
- PR へのリンクを付与
- Issue を close

失敗時:

- 失敗理由をコメント
- `saga:failed` ラベル付与を可能にする

## 7. PR ライフサイクル

### 7.1 PR 作成

PR 作成タイミング:

- `implement`
- `test/review`
- `verify`

の基準を初回で満たしたタイミング

方針:

- 初回は Draft PR を既定にする
- repository 設定または workflow 設定で ready-for-review に切り替え可能にする

PR title/body に含める内容:

- Issue への close link
- Summary
- Test Plan
- Automated by Saga の記載

### 7.2 PR 更新

fix loop 発生時:

- 同一 branch に追加 commit/push
- PR body の summary/test plan を更新可能
- コメントで修正内容を追記可能

### 7.3 CI 監視

監視対象:

- check runs
- status checks
- workflow runs
- required review state

監視方式:

- poll interval ごとに API 取得
- state 変化時のみローカルイベントとして記録

### 7.4 CI failure 時の fix loop

流れ:

1. 失敗チェックを取得
2. 失敗内容を artifact 化
3. `tester` または `implementer` を再実行
4. commit/push
5. 同じ PR のチェックを再監視

再試行回数:

- workflow または repo 設定で制御
- 上限を超えたら `failed` として停止

### 7.5 マージ

マージ条件:

- required checks が成功
- required reviews が満たされている、または bot bypass が許可されている
- merge conflict がない
- workflow 上の `verifier` が success

マージ方式:

- repo 設定に追従
- 未設定時は `squash` を推奨値とする

### 7.6 マージ後

- PR merged 状態を記録
- Issue 完了コメントを投稿
- Issue を close する
- worktree cleanup を行う

## 8. コメントコマンド

v1 でサポートするコマンド案:

- `/saga run`
- `/saga retry`
- `/saga cancel`
- `/saga status`
- `/saga resume`

実装方針:

- ポーリング時に新規コメントを検出
- bot 自身のコメントは無視
- 許可ユーザーまたは repository collaborator のみ有効にする

## 9. リコンシリエーション

daemon 起動時および定期的に以下を照合する。

- DB 上 `running` の task に対応する open PR が GitHub に存在するか
- PR が merged 済みなら local state を補正する
- Issue が手動 close されていたら task を cancel または complete に補正する
- branch が削除済みなら worktree cleanup を促進する

## 10. Rate Limit と障害対応

- GitHub API は token/app ごとに rate limit を監視する
- 余裕が少ない場合は polling interval を動的に伸ばす
- 429/5xx は exponential backoff で再試行する
- merge 直前に最新状態を再取得し、TOCTOU を避ける

## 11. データモデル

### 11.1 Task と GitHub の関連

| フィールド | 内容 |
|---|---|
| `repo_owner` | repository owner |
| `repo_name` | repository name |
| `issue_number` | Issue number |
| `issue_node_id` | GitHub node id |
| `branch_name` | 作業 branch |
| `pull_number` | PR number |
| `pull_node_id` | GitHub node id |
| `head_sha` | 最新 commit SHA |
| `merge_commit_sha` | merged 時の SHA |

### 11.2 イベント記録

- `issue_discovered`
- `issue_locked`
- `plan_posted`
- `pr_created`
- `pr_updated`
- `checks_failed`
- `fix_loop_started`
- `checks_passed`
- `pr_merged`
- `issue_closed`

## 12. 設定例

```yaml
github:
  auth:
    mode: app
    app_id: 123456
    installation_id: 987654
    private_key_file: /etc/saga/github-app.pem

  repositories:
    - owner: example
      name: backend
      default_branch: main
      selectors:
        labels: ["saga:ready"]
        assignees: ["saga-bot"]
        commands: ["/saga run"]
      pr:
        draft: true
        merge_method: squash
        auto_merge: true
      sync:
        poll_interval: 30s
        comment_progress: true
```

## 13. v1 受け入れ条件

1. open Issue を自動検知して task 化できる
2. plan comment を自動投稿できる
3. PR を自動作成できる
4. CI failure 時に同一 PR 上で fix loop を回せる
5. CI success 後に自動マージできる
6. Issue を完了状態に同期できる
7. daemon 再起動後に open PR / active task を再同期できる
