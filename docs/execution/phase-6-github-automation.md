# Phase 6: GitHub Automation

関連 Issue: [#8](https://github.com/soudai/saga/issues/8)

## 目的
- GitHub Issue と PR の完全自動連携を実装する。

## 作業
- repository poller の実装
- issue selector と issue lease の実装
- plan / progress / result comment の実装
- PR create / update の実装
- check / status / review polling の実装
- merge と issue close / sync の実装

## 実装ステップ
1. GitHub 認証クライアントを初期化し、対象 repository の設定読み込みと API ラッパーを実装する。
2. issue poller と selector を追加し、`saga:ready`, assignee, `/sg run` コメントを基準に task 化する。
3. 二重実行を防ぐ issue lease と、plan / progress / failed / completed comment の同期処理を追加する。
4. branch と head SHA から PR を create / update する処理を実装する。
5. check run, status check, workflow run, review state の polling を追加し、merge 条件を満たしたら issue close まで同期する。

## 完了条件
- open Issue を自動検出して PR まで作成できる
- success 時に PR merge と Issue close まで進む

## 参照
- [Implementation Plan](../implementation-plan.md)
- [GitHub Integration](../github-integration.md)
- [Requirements](../requirements.md)
- [Architecture](../architecture.md)
