# Phase 1: Runtime Foundation

関連 Issue: [#3](https://github.com/soudai/saga/issues/3)

## 目的
- daemon と CLI の最小実行基盤を作る。

## 作業
- `saga serve` と `saga doctor` の実装
- config loader の実装
- `slog` による構造化ログ初期化
- systemd readiness notify の導入

## 実装ステップ
1. `cobra` ベースで `saga` ルートコマンドと `serve`, `doctor`, `version` の最小サブコマンドを定義する。
2. `/etc/saga/config.yaml` とプロジェクトローカル override を見据えた config schema を決め、validation を実装する。
3. `slog` を用いたロガー初期化と、runtime/state/log ディレクトリの解決処理を追加する。
4. daemon 起動シーケンスを `internal/app` または `internal/daemon` に切り出し、今後 Unix socket や store を差し込める構成にする。
5. systemd 環境では `sd_notify`、非 systemd 環境では no-op で動く起動処理を用意する。

## 完了条件
- `systemd` なしでも `saga serve` が起動する
- config の読み込みと validation が動作する

## 参照
- [Docs README](../README.md)
- [Implementation Plan](../implementation-plan.md)
- [Architecture](../architecture.md)
- [Tech Stack](../tech-stack.md)
