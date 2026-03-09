# Phase 2: State & Control Plane

Issue: #4

## Goal
- SQLite を中心にした state/control plane を作る。

## Scope
- SQLite schema の定義
- `repository`, `task`, `run`, `subagent`, `lease` テーブルの作成
- Unix socket API の実装
- `status`, `cancel`, `retry`, `resume` API の実装

## Done
- task を DB に登録できる
- CLI から daemon の状態を取得できる
