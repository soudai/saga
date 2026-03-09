# Phase 7: Validation & Recovery

Issue: #9

## Goal
- 中断と再開に耐える validation / recovery を実装する。

## Scope
- startup reconciliation の実装
- stale task detection の実装
- open PR からの状態再構築
- failed CI からの fix loop 自動化
- resume / retry policy の実装

## Done
- daemon 再起動後に in-progress task を再同期できる
- CI failure からの自動再試行が動く
