# wez-mux — WezTerm Multi-Agent Pane Orchestrator CLI

WezTerm 上でマルチエージェント開発環境を構築・操作するための Go 製 CLI ツール。
`wezterm cli` コマンドをラップし、ラベルベースのペイン管理と通信を提供する。

## 背景

- Claude Code Agent Teams の split-pane モードは tmux/iTerm2 のみ対応
- WezTerm は `wezterm cli` で split-pane, send-text, list 等の API を持つ
- これらを統合し、ラベル指定で簡単にペイン操作できる CLI が必要

## コマンド体系

```
wez-mux init [--config <path>]       # ペイン作成 + エージェント起動
wez-mux send <label> "<message>"     # メッセージ送信（自動 Enter）
wez-mux read <label> [lines]         # ペイン出力の読み取り
wez-mux list                         # 管理中のペイン一覧
wez-mux status                       # 各ペインの稼働状態
wez-mux kill [<label> | --all]       # ペイン終了
```

---

## サブコマンド詳細

### `wez-mux init`

5ペインのレイアウトを作成し、各ペインでエージェントを自動起動する。

**動作フロー:**

1. `wezterm cli list --format json` で現在のペイン情報を取得
2. 起点ペイン（現在のペイン）を orchestrator として登録
3. `wezterm cli split-pane` を4回実行しペインを分割
4. 各ペイン ID をレジストリに保存
5. `wezterm cli send-text` で各ペインにエージェント起動コマンドを送信

**レイアウト構成:**

```
┌──────────────┬──────────────┬──────────────┐
│ orchestrator │   analyzer   │   designer   │
│   (Opus)     │  (Sonnet)    │   (Opus)     │
├──────────────┴──────┬───────┴──────────────┤
│       coder         │       tester         │
│  (Sonnet/Haiku)     │  (Sonnet/Haiku)      │
└─────────────────────┴──────────────────────┘
```

**分割手順（wezterm cli の実行順序）:**

```bash
# 1. 起点ペイン（orchestrator）を右に分割 → analyzer を生成
ANALYZER=$(wezterm cli split-pane --right --percent 66 --pane-id $ORCHESTRATOR)

# 2. analyzer を右に分割 → designer を生成
DESIGNER=$(wezterm cli split-pane --right --percent 50 --pane-id $ANALYZER)

# 3. orchestrator を下に分割（top-level で下段を作る）→ coder を生成
CODER=$(wezterm cli split-pane --bottom --percent 50 --top-level --pane-id $ORCHESTRATOR)

# 4. coder を右に分割 → tester を生成
TESTER=$(wezterm cli split-pane --right --percent 50 --pane-id $CODER)
```

**各ペインで実行するエージェント起動コマンド:**

```bash
# orchestrator（起点ペインなので起動不要、または手動で claude を起動済み）

# analyzer
wezterm cli send-text --pane-id $ANALYZER --no-paste \
  "claude --model sonnet --skill agent-analyzer\n"

# designer
wezterm cli send-text --pane-id $DESIGNER --no-paste \
  "claude --model opus --skill agent-designer\n"

# coder
wezterm cli send-text --pane-id $CODER --no-paste \
  "claude --model sonnet --skill agent-coder\n"

# tester
wezterm cli send-text --pane-id $TESTER --no-paste \
  "claude --model haiku --skill agent-tester\n"
```

**オプション:**

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `--config <path>` | `~/.config/wez-mux/default.yaml` | レイアウト定義ファイルのパス |
| `--cwd <path>` | カレントディレクトリ | 各ペインの作業ディレクトリ |
| `--no-start` | false | ペイン分割のみ行い、エージェント起動はスキップ |

---

### `wez-mux send`

指定ラベルのペインにメッセージを送信する。

```bash
wez-mux send <label> "<message>"
```

**動作:**
1. レジストリからラベルに対応する pane_id を解決
2. `wezterm cli send-text --pane-id <ID> --no-paste "<message>\n"` を実行
3. `--no-paste` でブラケットペーストを回避し、末尾に `\n` を付与して自動送信

**オプション:**

| フラグ | デフォルト | 説明 |
|--------|-----------|------|
| `--no-enter` | false | 末尾の `\n` を付けない（入力途中のテキストを送る場合） |

---

### `wez-mux read`

指定ラベルのペインの出力を取得する。

```bash
wez-mux read <label> [lines]
```

**動作:**
1. レジストリからラベルに対応する pane_id を解決
2. `wezterm cli get-text --pane-id <ID>` でペインのスクロールバッファを取得
3. 末尾 N 行を stdout に出力

**引数:**

| 引数 | デフォルト | 説明 |
|------|-----------|------|
| `lines` | 50 | 取得する末尾行数 |

---

### `wez-mux list`

管理中の全ペインの一覧を表示する。

```bash
wez-mux list
```

**出力例:**

```
LABEL          PANE_ID  MODEL         RUNTIME           STATUS
orchestrator   0        opus          claude-code       active
analyzer       1        sonnet        claude-code       active
designer       2        opus          claude-code       active
coder          3        sonnet        claude-code+codex active
tester         4        haiku         claude-code+codex active
```

---

### `wez-mux status`

各ペインの稼働状態を確認する。

```bash
wez-mux status
```

**動作:**
1. レジストリから全ペインを取得
2. `wezterm cli list --format json` で WezTerm 側の現在のペイン一覧と突合
3. レジストリ上のペインが WezTerm に存在するかチェック
4. 各ペインの最終出力行を取得して状態を推定

---

### `wez-mux kill`

ペインを終了する。

```bash
wez-mux kill <label>      # 特定のペインを終了
wez-mux kill --all         # 全ペインを終了（orchestrator 以外）
```

**動作:**
1. `wezterm cli kill-pane --pane-id <ID>` でペインを終了
2. レジストリから削除

---

## レジストリ

ペインのラベルと pane_id のマッピングを管理する。

**保存場所:** `~/.config/wez-mux/registry.json`

**形式:**

```json
{
  "workspace": "multi-agent",
  "created_at": "2026-04-02T15:30:00Z",
  "cwd": "/home/user/project",
  "panes": {
    "orchestrator": {
      "pane_id": 0,
      "model": "opus",
      "runtime": "claude-code",
      "skill": "agent-orchestrator"
    },
    "analyzer": {
      "pane_id": 1,
      "model": "sonnet",
      "runtime": "claude-code",
      "skill": "agent-analyzer"
    },
    "designer": {
      "pane_id": 2,
      "model": "opus",
      "runtime": "claude-code",
      "skill": "agent-designer"
    },
    "coder": {
      "pane_id": 3,
      "model": "sonnet",
      "runtime": "claude-code+codex",
      "skill": "agent-coder"
    },
    "tester": {
      "pane_id": 4,
      "model": "haiku",
      "runtime": "claude-code+codex",
      "skill": "agent-tester"
    }
  }
}
```

---

## 設定ファイル

**保存場所:** `~/.config/wez-mux/default.yaml`

レイアウトとエージェント構成をカスタマイズする。

```yaml
workspace: multi-agent

layout:
  rows:
    - panes:
        - label: orchestrator
          model: opus
          skill: agent-orchestrator
          percent: 33
        - label: analyzer
          model: sonnet
          skill: agent-analyzer
          percent: 33
        - label: designer
          model: opus
          skill: agent-designer
          percent: 34
    - panes:
        - label: coder
          model: sonnet
          skill: agent-coder
          codex: true
          percent: 50
        - label: tester
          model: haiku
          skill: agent-tester
          codex: true
          percent: 50

# 行の高さ比率（上段:下段）
row_ratio: [50, 50]
```

---

## Go プロジェクト構成

```
project/wez-mux/
├── main.go                 # エントリポイント（cobra rootCmd）
├── go.mod
├── go.sum
├── cmd/
│   ├── root.go             # ルートコマンド定義
│   ├── init.go             # init サブコマンド
│   ├── send.go             # send サブコマンド
│   ├── read.go             # read サブコマンド
│   ├── list.go             # list サブコマンド
│   ├── status.go           # status サブコマンド
│   └── kill.go             # kill サブコマンド
├── internal/
│   ├── wezterm/            # wezterm cli ラッパー
│   │   ├── client.go       # WezTerm CLI 実行の共通処理
│   │   ├── pane.go         # split-pane, kill-pane, activate-pane
│   │   ├── text.go         # send-text, get-text
│   │   └── list.go         # list --format json のパース
│   ├── registry/           # ペインレジストリ管理
│   │   └── registry.go     # JSON 読み書き、ラベル→ID 解決
│   ├── config/             # 設定ファイル読み込み
│   │   └── config.go       # YAML パース、デフォルト値
│   └── layout/             # レイアウト計算・実行
│       └── layout.go       # 分割順序の決定、ペイン生成
└── configs/
    └── default.yaml        # デフォルト設定
```

## 依存ライブラリ

| ライブラリ | 用途 |
|-----------|------|
| `github.com/spf13/cobra` | CLI フレームワーク |
| `gopkg.in/yaml.v3` | YAML 設定パース |
| `encoding/json` (stdlib) | レジストリ JSON 操作 |
| `os/exec` (stdlib) | wezterm cli プロセス実行 |

## ビルドとインストール

```bash
cd project/wez-mux
go build -o wez-mux .
# PATH の通った場所にコピー
cp wez-mux ~/.local/bin/
```

---

## 使用例

### 環境の起動

```bash
# プロジェクトディレクトリで初期化
cd ~/project/my-app
wez-mux init

# → 5ペインが作成され、各エージェントが自動起動
```

### Orchestrator からの操作

```bash
# analyzer に調査依頼
wez-mux send analyzer "src/auth/ ディレクトリの認証フローを調査してください"

# 結果を確認
wez-mux read analyzer 100

# designer に設計依頼
wez-mux send designer "調査結果: JWT + httpOnly cookie 方式。リフレッシュトークン対応の設計をお願いします"

# coder に実装依頼
wez-mux send coder "設計に基づいて src/auth/refresh.ts を実装してください"

# tester にテストコード作成依頼
wez-mux send tester "src/auth/refresh.ts のテストコードを作成してください。正常系・異常系・トークン期限切れのケースを網羅"

# 全体の状態確認
wez-mux status

# 終了時
wez-mux kill --all
```

---

## 前提条件

- WezTerm がインストール済みで `wezterm cli` が PATH にあること
- Go 1.21 以上（ビルド時）
- Claude Code がインストール済みで `claude` コマンドが使えること
- 各 agent-* スキルが `~/.claude/skills/` または プロジェクトの `.claude/skills/` に配置されていること
