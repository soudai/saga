# Phase 8: systemd Hardening

Issue: #10

## Goal
- WSL2 + systemd 前提の常駐運用を固める。

## Scope
- unit file の作成
- install 手順の作成
- `doctor` の systemd / WSL2 チェック追加
- graceful shutdown の実装
- journald 出力確認

## Done
- `systemctl enable --now saga` で常駐できる
- 再起動、停止、異常終了時の挙動を確認できる
