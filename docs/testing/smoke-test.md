# Smoke Test

## Goal
- `saga serve` が起動し、ローカル control plane が応答できることを最小シナリオで確認する。

## Scenario
1. `saga doctor` が成功する
2. `saga serve` を起動する
3. `saga status` で daemon に接続できる
4. SQLite state store と Unix socket が作成される
