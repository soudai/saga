# Saga Documentation

## 目的

このディレクトリは、`saga` を Go 製の AI Agent Framework として実装するための設計ドキュメントをまとめたものです。  
`saga` は `Codex CLI` を実行エンジンとして利用し、WSL2 上の Ubuntu で `systemd` 管理の常駐サービスとして動作し、GitHub Issue/PR と完全自動連携しながらサブエージェントで実装・テスト・検証を進めることを前提にします。

## ドキュメント一覧

- [requirements.md](./requirements.md)
  - v1 の機能要件、非機能要件、受け入れ条件
- [architecture.md](./architecture.md)
  - システム構成、コンポーネント責務、状態遷移、配置方針
- [github-integration.md](./github-integration.md)
  - GitHub Issue/PR 完全自動連携の詳細仕様
- [tech-stack.md](./tech-stack.md)
  - 採用技術、依存関係、開発要件、運用要件
- [implementation-plan.md](./implementation-plan.md)
  - 実装フェーズ、マイルストーン、完了条件
- [systemd-wsl2.md](./systemd-wsl2.md)
  - WSL2/systemd 上での運用要件と unit 例

## v1 スコープ要約

- Go 単一バイナリで提供する
- WSL2 Ubuntu 上で `systemd` 常駐サービスとして稼働する
- `Codex CLI` をサブエージェント実行エンジンとして利用する
- YAML 定義のワークフローで `plan -> implement -> test/review -> verify -> merge` を実行する
- GitHub Issue の取得、進捗コメント、PR 作成、CI 監視、修正ループ、マージ、Issue 完了までを自動化する
- タスクごとに git worktree を分離し、並列実行と再試行に耐える
- 実行履歴、成果物、ログ、状態を永続化し、サービス再起動後も復旧可能にする

## ドキュメントの読み順

1. [requirements.md](./requirements.md)
2. [github-integration.md](./github-integration.md)
3. [architecture.md](./architecture.md)
4. [tech-stack.md](./tech-stack.md)
5. [implementation-plan.md](./implementation-plan.md)
6. [systemd-wsl2.md](./systemd-wsl2.md)
