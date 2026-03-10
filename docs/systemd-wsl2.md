# Saga Operations on WSL2 and systemd

## 1. 目的

本書は `sg` を WSL2 Ubuntu 上で `systemd` 管理の常駐サービスとして動かす際の運用要件をまとめる。

## 2. 前提

- Windows 側で WSL2 が有効
- Ubuntu が導入済み
- `/etc/wsl.conf` で `systemd=true`
- `Codex CLI`, `git`, `sg` binary が Ubuntu 側にインストール済み

## 3. 推奨配置

- binary: `/usr/local/bin/sg`
- config: `/etc/sg/config.yaml`
- secrets: `/etc/sg/sg.env`
- state: `/var/lib/sg`
- runtime: `/run/sg`
- logs: `/var/log/sg`

## 4. systemd unit 例

```ini
[Unit]
Description=Saga Orchestrator
After=network-online.target
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/local/bin/sg serve --config /etc/sg/config.yaml
Restart=on-failure
RestartSec=5
EnvironmentFile=/etc/sg/sg.env
StateDirectory=sg
RuntimeDirectory=sg
LogsDirectory=sg
KillMode=control-group
TimeoutStopSec=180

[Install]
WantedBy=multi-user.target
```

## 5. 運用チェックリスト

- `systemctl status sg`
- `journalctl -u sg -f`
- `sg doctor`
- `sg status`

## 6. WSL2 固有の注意点

- 対象 repository は `/home/...` など Linux 側に置く
- `/mnt/c/...` 上の repo は performance と permission の面で非推奨
- WSL 再起動後に `systemd` 起動と network ready を待つ
- 外部 webhook 公開を前提にしない

## 7. 障害対応

- 起動失敗時は `journalctl -u sg -n 200`
- stale worktree は `sg doctor` と startup cleanup で検出
- GitHub credential 不備は preflight で失敗させる
