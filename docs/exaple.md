# Saga Usage Examples

## 目的

このドキュメントは、`sg` を実際に使うときの代表的なユースケースを、手順ベースでまとめたサンプル集です。  
README よりも実行順と操作例に寄せてあります。

## 前提

- Linux または WSL2 上で作業している
- `go`, `git`, `gh` が利用可能
- `gh auth login` が必要なケースでは事前に認証済みである

---

## 1. 初回セットアップして daemon を起動する

### 1-1. binary を build する

```bash
make build
```

### 1-2. config を生成する

```bash
./bin/sg init ./sg.local.yaml
```

対話で確認される主な値:

- `runtime.state_dir`
- `runtime.run_dir`
- `runtime.log_dir`
- `server.socket_path`
- `log.level`

### 1-3. daemon を起動する

```bash
./bin/sg serve --config ./sg.local.yaml
```

別ターミナルで疎通確認します。

```bash
./bin/sg doctor --config ./sg.local.yaml
./bin/sg status --config ./sg.local.yaml
```

期待結果:

- `doctor` が path 周りのチェック結果を返す
- `status` が `tasks=0 active_runs=0` などの現在状態を返す

---

## 2. 既存の GitHub Issue を local task として登録する

対象 Issue が既に GitHub 上にある場合は、issue number を指定して queue に入れます。

```bash
./bin/sg enqueue soudai/saga 123 --config ./sg.local.yaml
```

確認:

```bash
./bin/sg status --config ./sg.local.yaml
```

期待結果:

- `task id=... repo=soudai/saga issue=123 state=queued` が見える

使いどころ:

- 既存 repository の Issue を `sg` に取り込む
- daemon/control plane までの接続を最小経路で確認する

---

## 3. markdown から「タスク指示書 Issue」を作成する

短いメモや要件整理を、GitHub 上の task-instruction Issue に変換したい場合の例です。

### 3-1. brief を markdown で用意する

```bash
cat > task.md <<'EOF'
Implement the enqueue flow for GitHub issues.
EOF
```

### 3-2. GitHub に送る前に body を確認する

```bash
./bin/sg issue draft soudai/saga --from-file task.md
```

このコマンドは次を自動で補います。

- 先頭見出しが無ければ title を推定して H1 を付ける
- `Background / Goal`, `Scope`, `Acceptance Criteria` を持つ最小テンプレートを作る
- `sg` 生成の marker comment を先頭に付与する

### 3-3. GitHub Issue を作成する

```bash
./bin/sg issue create soudai/saga --from-file task.md
```

期待結果:

- `issue #<number> https://github.com/...` が出力される

### 3-4. 作成した Issue をそのまま local task に登録する

```bash
./bin/sg issue create soudai/saga --from-file task.md --enqueue --config ./sg.local.yaml
```

期待結果:

- Issue 作成結果
- 続けて `task id=... repo=... issue=... state=queued`

使いどころ:

- 実装の指示書を先に GitHub に残したい
- `repository + issue number` ベースの既存 task model を崩さずに投入したい

---

## 4. task の状態を確認・操作する

現在の local task 一覧を確認します。

```bash
./bin/sg status --config ./sg.local.yaml
```

task id を指定して状態を操作できます。

```bash
./bin/sg cancel 1 --config ./sg.local.yaml
./bin/sg retry 1 --config ./sg.local.yaml
./bin/sg resume 1 --config ./sg.local.yaml
```

想定用途:

- `cancel`: 実行を止めたい
- `retry`: `queued` に戻したい
- `resume`: `running` として再開扱いにしたい

---

## 5. WSL2 + systemd で常駐させる

systemd 配下で運用したい場合は、次のドキュメントを参照してください。

- [`systemd-wsl2.md`](./systemd-wsl2.md)
- [`testing/smoke-test.md`](./testing/smoke-test.md)
- [`../contrib/systemd/sg.service`](../contrib/systemd/sg.service)

最小確認:

```bash
systemctl start sg
systemctl status sg
journalctl -u sg -f
```

---

## 現時点の注意

- `sg issue create` は GitHub CLI (`gh`) の install と認証が必要
- 現在の `sg` は local control plane と task 登録までは end-to-end だが、daemon が自律的に GitHub Issue を取り込み続ける loop はまだ完全接続されていない
- `sg issue create` の初版は `--from-file` 前提で、AI 対話による指示書生成はまだ含まない
