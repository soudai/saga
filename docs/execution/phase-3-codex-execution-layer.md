# Phase 3: Codex Execution Layer

Issue: #5

## Goal
- Codex CLI を安全に実行できる層を作る。

## Scope
- Codex adapter の実装
- sandbox / network / model の設定適用
- stdout / stderr / result の保存
- timeout / cancel / kill の実装
- mock Codex による integration test の追加

## Done
- 単一 stage を Codex で実行できる
- timeout と cancel が正しく動作する
