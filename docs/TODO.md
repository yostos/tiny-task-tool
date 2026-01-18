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

- [ ] 設定ファイル仕様（場所、全オプション）
- [ ] キーバインド仕様（各キーの詳細な動作）
- [ ] ファイル配置（メインファイル、archive.md の場所）
- [ ] TUI画面レイアウト（画面構成）
- [ ] エラー処理（ファイルが無い場合などの動作）

## 実装

（仕様書完成後に追加）
