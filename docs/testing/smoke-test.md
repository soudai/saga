# Smoke Test

## Goal
- `sg serve` が起動し、ローカル control plane が応答できることを最小シナリオで確認する。

## Scenario
1. `sg doctor` が成功する
2. 別のターミナル、バックグラウンド実行、または `systemctl start sg` で `sg serve` を継続起動する
3. `sg serve` が動作している状態で、別のターミナルから `sg status` を実行して daemon に接続できる
4. SQLite state store と Unix socket が作成される
