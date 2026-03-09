# Phase 2: State & Control Plane

関連 Issue: [#4](https://github.com/soudai/saga/issues/4)

## 目的
- SQLite を中心にした state/control plane を作る。

## 作業
- SQLite schema の定義
- `repository`, `task`, `run`, `subagent`, `lease` テーブルの作成
- Unix socket API の実装
- `status`, `cancel`, `retry`, `resume` API の実装

## 実装ステップ
1. `repository`, `task`, `run`, `subagent`, `lease` の責務と状態遷移を整理し、SQLite schema と migration 方針を決める。
2. store 層で `CreateTask`, `StartRun`, `AcquireLease`, `ListActiveRuns` などの基本 repository API を定義する。
3. daemon と CLI の接続に使う Unix socket API を `net/http` ベースで設計し、`/status`, `/tasks/{id}/cancel`, `/tasks/{id}/retry`, `/tasks/{id}/resume` などの最小エンドポイントを決める。
4. request/response の JSON schema を固め、CLI 側に `status`, `cancel`, `retry`, `resume` を実装する。
5. DB 再起動復旧と lease の排他を integration test で検証する。

## 完了条件
- task を DB に登録できる
- CLI から daemon の状態を取得できる

## 参照
- [Implementation Plan](../implementation-plan.md)
- [Requirements](../requirements.md)
- [Architecture](../architecture.md)
- [Tech Stack](../tech-stack.md)
