# Phase 4: Git Worktree Layer

関連 Issue: [#6](https://github.com/soudai/saga/issues/6)

## 目的
- task ごとの隔離実行を支える worktree 管理を実装する。

## 作業
- primary worktree の作成
- branch 命名の実装
- retry / recreate ロジックの追加
- startup orphan cleanup の追加
- shadow worktree の生成

## 実装ステップ
1. 管理対象 repository の preflight を追加し、`git worktree` が利用可能な状態かを検証する。
2. task ID と issue 番号から一意な branch 名を生成するルールを定義し、primary worktree 作成処理を実装する。
3. `tester`, `reviewer`, `verifier` 向けに shadow worktree を派生させる処理を追加する。
4. retry 時の再利用・再作成条件を定義し、破損した worktree を安全に作り直せるようにする。
5. 起動時 orphan cleanup と cleanup の integration test を追加する。

## 完了条件
- task ごとに独立 worktree を作成できる
- cleanup の自動テストが通る

## 参照
- [Implementation Plan](../implementation-plan.md)
- [Requirements](../requirements.md)
- [Architecture](../architecture.md)
- [Tech Stack](../tech-stack.md)
