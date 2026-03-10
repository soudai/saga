# Phase 9: Release Readiness

関連 Issue: [#11](https://github.com/soudai/saga/issues/11)

## 目的
- v1 を出荷可能な状態まで仕上げる。

## 作業
- smoke test の追加
- docs の最終更新
- sample config の追加
- release pipeline の構築
- versioning / changelog の整備

## 実装ステップ
1. smoke test の対象機能とクリティカルパスを洗い出し、`systemctl start sg` から Issue 検出、PR 作成までの最小シナリオを定義する。
2. v1 仕様との差分レビューを行い、更新が必要なユーザー向け / 開発者向けドキュメントを一覧化して反映する。
3. `/etc/sg/config.yaml` と `.sg/config.yaml` を前提に、最小構成と推奨構成の sample config を追加する。
4. build, test, package, release note 生成までを含む release pipeline を定義し、tag もしくは release branch で実行できるようにする。
5. versioning policy を文書化し、`CHANGELOG.md` とリリースノートの更新手順を確立する。

## 完了条件
- v1 acceptance criteria を満たす
- binary を配布できる

## 参照
- [Implementation Plan](../implementation-plan.md)
- [Requirements](../requirements.md)
- [Tech Stack](../tech-stack.md)
- [systemd / WSL2](../systemd-wsl2.md)
