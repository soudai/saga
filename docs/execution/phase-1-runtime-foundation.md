# Phase 1: Runtime Foundation

Issue: #3

## Goal
- daemon と CLI の最小実行基盤を作る。

## Scope
- `saga serve` と `saga doctor` の実装
- config loader の実装
- `slog` による構造化ログ初期化
- systemd readiness notify の導入

## Done
- `systemd` なしでも `saga serve` が起動する
- config の読み込みと validation が動作する
