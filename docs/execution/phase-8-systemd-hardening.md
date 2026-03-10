# Phase 8: systemd Hardening

関連 Issue: [#10](https://github.com/soudai/saga/issues/10)

## 目的
- WSL2 + systemd 前提の常駐運用を固める。

## 作業
- unit file の作成
- install 手順の作成
- `doctor` の systemd / WSL2 チェック追加
- graceful shutdown の実装
- journald 出力確認

## 実装ステップ
1. `Type=notify`, `Restart=on-failure`, `StateDirectory`, `RuntimeDirectory`, `LogsDirectory` を満たす unit file を追加する。
2. `/usr/local/bin/sg` と `/etc/sg/` を前提にした install / upgrade 手順を文書化する。
3. `doctor` に systemd 有効化確認、WSL2 判定、`/mnt/c` 警告、runtime directory 書き込み確認を追加する。
4. daemon 停止時に子プロセスを確実に終了する graceful shutdown を実装する。
5. journald と artifact log の両方に主要イベントが出ることを smoke test で確認する。

## 完了条件
- `systemctl enable --now sg` で常駐できる
- 再起動、停止、異常終了時の挙動を確認できる

## 参照
- [Implementation Plan](../implementation-plan.md)
- [systemd / WSL2](../systemd-wsl2.md)
- [Requirements](../requirements.md)
- [Tech Stack](../tech-stack.md)
