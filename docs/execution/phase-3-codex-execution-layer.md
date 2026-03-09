# Phase 3: Codex Execution Layer

関連 Issue: [#5](https://github.com/soudai/saga/issues/5)

## 目的
- Codex CLI を安全に実行できる層を作る。

## 作業
- Codex adapter の実装
- sandbox / network / model の設定適用
- stdout / stderr / result の保存
- timeout / cancel / kill の実装
- mock Codex による integration test の追加

## 実装ステップ
1. stage 実行に必要な入力を `RunnerRequest` のような構造体にまとめ、`codex` パッケージの public interface を定義する。
2. `os/exec` ベースで Codex CLI を起動し、sandbox, network, model, timeout を引数と環境変数に反映する。
3. stdout, stderr, exit code を artifact として保存し、`result.json` に正規化する。
4. cancel と timeout の両系統でプロセス終了を制御し、子プロセスリークがないことを確認する。
5. mock Codex バイナリを使った integration test を追加し、正常終了・失敗・timeout・cancel を検証する。

## 完了条件
- 単一 stage を Codex で実行できる
- timeout と cancel が正しく動作する

## 参照
- [Implementation Plan](../implementation-plan.md)
- [Requirements](../requirements.md)
- [Architecture](../architecture.md)
- [Tech Stack](../tech-stack.md)
