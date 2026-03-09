# Phase 5: Workflow Engine

関連 Issue: [#7](https://github.com/soudai/saga/issues/7)

## 目的
- 宣言的 workflow を実行できるエンジンを実装する。

## 作業
- YAML schema の定義
- stage transition の実装
- parallel stage の実装
- artifact handoff の実装
- fix loop の実装
- builtin workflow の追加

## 実装ステップ
1. stage 定義の schema を固め、`role`, `sandbox`, `network`, `timeout`, `retry`, `transition`, `worktree_mode` を読み込める parser を作る。
2. 単一路線の state machine と artifact handoff を先に実装し、`plan -> implement -> verify -> complete` を動かす。
3. `test` と `review` の parallel stage を追加し、収束条件を定義する。
4. validation failure 時に `implement` へ戻る fix loop を追加し、最大再試行回数も制御できるようにする。
5. v1 builtin workflow を用意し、GitHub 自動化フェーズへ接続しやすいインターフェースに整える。

## 完了条件
- `plan -> implement -> test/review -> verify -> complete` が動く
- ここでの `complete` は workflow engine 内の最終状態を指し、PR の `merge` は Phase 6 以降の GitHub 自動化で扱う
- validation failure で `implement` に戻れる

## 参照
- [Docs README](../README.md)
- [Implementation Plan](../implementation-plan.md)
- [Requirements](../requirements.md)
- [Architecture](../architecture.md)
