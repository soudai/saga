# Phase 4: Git Worktree Layer

Issue: #6

## Goal
- task ごとの隔離実行を支える worktree 管理を実装する。

## Scope
- primary worktree の作成
- branch 命名の実装
- retry / recreate ロジックの追加
- startup orphan cleanup の追加
- shadow worktree の生成

## Done
- task ごとに独立 worktree を作成できる
- cleanup の自動テストが通る
