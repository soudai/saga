# Phase 7: Validation & Recovery

関連 Issue: [#9](https://github.com/soudai/saga/issues/9)

## 目的
- 中断と再開に耐える validation / recovery を実装する。

## 作業
- startup reconciliation の実装
- stale task detection の実装
- open PR からの状態再構築
- failed CI からの fix loop 自動化
- resume / retry policy の実装

## 実装ステップ
1. daemon 起動時に DB, worktree, GitHub PR を突き合わせる reconciliation 処理を追加する。
2. stale task の判定条件と lease 期限切れ処理を定義し、宙に浮いた run を検知できるようにする。
3. open PR と head SHA から task / run 状態を再構築するロジックを実装する。
4. CI failure の詳細を artifact 化し、`tester` / `implementer` の再実行に渡す fix loop を自動化する。
5. `resume` と `retry` のポリシー差分を明確化し、どこから再開するかを store に記録する。

## 完了条件
- daemon 再起動後に in-progress task を再同期できる
- CI failure からの自動再試行が動く

## 参照
- [Implementation Plan](../implementation-plan.md)
- [GitHub Integration](../github-integration.md)
- [Architecture](../architecture.md)
- [Requirements](../requirements.md)
