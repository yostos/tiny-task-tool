# ttt - TODO

## 技術選定

### 決定済み

- [x] プログラミング言語 → Go
- [x] 設定ファイル処理（TOML） → pelletier/go-toml v2
- [x] Markdownパーサー → yuin/goldmark
- [x] CLI引数解析 → spf13/pflag
- [x] TUIフレームワーク → charmbracelet/bubbletea + lipgloss + bubbles
- [x] 静的解析 → golangci-lint
- [x] テスト → testing + testify（PR時にGitHub Actionsで自動化）

## 仕様書の完成

### 未定義の項目

- [x] 設定ファイル仕様（場所、全オプション）
- [x] キーバインド仕様（各キーの詳細な動作）
- [x] ファイル配置（メインファイル、archive.md の場所）
- [x] TUI画面レイアウト（画面構成）
- [x] エラー処理（ファイルが無い場合などの動作）

## 実装

### Phase 1: プロジェクト基盤

- [ ] プロジェクト初期化（go mod init, ディレクトリ構造）
- [ ] 設定ファイル読み込み（go-toml v2）
- [ ] CLIパーサー（pflag: -t, -h, -v オプション）

### Phase 2: TUI基本機能

- [ ] TUIの基本構造（bubbletea: Model/Update/View）
- [ ] ファイル読み込み・表示
- [ ] スクロール機能（↑↓, j/k, g/G, Ctrl+u/d）
- [ ] フッター表示（キーヒント、スクロール位置）
- [ ] キーバインド設定の反映

### Phase 3: コア機能

- [ ] 外部エディタ起動（`e`キー）
- [ ] 完了タスク検出（`- [x]`）
- [ ] `@done(日付)` 自動追加
- [ ] アーカイブ機能（`a`キー、auto設定）
- [ ] 再読み込み（`r`キー、エディタ終了後の自動再読み込み）

### Phase 4: CLI機能

- [ ] タスク追加（`-t`, `--task`オプション）
- [ ] ヘルプ表示（`--help`）
- [ ] バージョン表示（`--version`）

### Phase 5: Git連携

- [ ] 自動 `git init`
- [ ] 自動コミット（変更検出時）
- [ ] `git.auto_commit` 設定の反映

### Phase 6: 仕上げ

- [ ] ヘルプオーバーレイ（`?`/`h`キー）
- [ ] ステータスメッセージ表示（3秒タイムアウト）
- [ ] エラー処理の実装
- [ ] テスト作成（コア機能）
