# Phase 5: Workflow Engine

Issue: #7

## Goal
- 宣言的 workflow を実行できるエンジンを実装する。

## Scope
- YAML schema の定義
- stage transition の実装
- parallel stage の実装
- artifact handoff の実装
- fix loop の実装
- builtin workflow の追加

## Done
- `plan -> implement -> test/review -> verify -> complete` が動く
- validation failure で `implement` に戻れる
