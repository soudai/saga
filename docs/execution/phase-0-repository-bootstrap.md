# Phase 0: Repository Bootstrap

関連 Issue: [#2](https://github.com/soudai/saga/issues/2)

## 目的
- Go module と基本ディレクトリ構成を作り、実装を開始できる状態にする。

## 作業
- `go.mod` の初期化
- `cmd/sg`, `internal/`, `test/` の作成
- `Makefile` または `justfile` の追加
- CI ひな形の追加

## 実装ステップ
1. module path とバージョン埋め込み方針を決めて `go.mod` を初期化する。
2. `cmd/sg` のエントリポイントと `sg version` を追加し、最小バイナリを起動できる状態にする。
3. `internal/`, `pkg/`, `test/` の最低限のディレクトリ構成を作り、以降のフェーズで追加するパッケージの受け皿を用意する。
4. `Makefile` または `justfile` に `test`, `build`, `fmt` の基本ターゲットを定義する。
5. `go test ./...` を回す最小 CI を追加し、空の実装でも継続的に検証できる状態にする。

## 完了条件
- `go test ./...` が通る
- `sg version` が動作する

## 参照
- [Docs README](../README.md)
- [Implementation Plan](../implementation-plan.md)
- [Requirements](../requirements.md)
