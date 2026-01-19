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

- [x] プロジェクト初期化（go mod init, ディレクトリ構造）
- [x] 設定ファイル読み込み（go-toml v2）
- [x] 設定ファイル自動作成（存在しない場合はデフォルト値で作成）
- [x] CLIパーサー（pflag: -t, -h, -v オプション）

### Phase 2: TUI基本機能

- [x] TUIの基本構造（bubbletea: Model/Update/View）
- [x] ファイル読み込み・表示
- [x] スクロール機能（↑↓, j/k, g/G, Ctrl+u/d）
- [x] フッター表示（キーヒント、スクロール位置）
- [x] キーバインド設定の反映

### Phase 3: コア機能

- [x] 外部エディタ起動（`e`キー）
- [x] 完了タスク検出（`- [x]`）
- [x] `@done(日付)` 自動追加
- [x] アーカイブ機能（`a`キー）
- [x] アーカイブ機能（auto設定）
- [x] 再読み込み（`r`キー、エディタ終了後の自動再読み込み）

### Phase 4: CLI機能

- [x] タスク追加（`-t`, `--task`オプション）
- [x] ヘルプ表示（`--help`）
- [x] バージョン表示（`--version`）

### Phase 5: Git連携

- [x] 自動 `git init`
- [x] 自動コミット（変更検出時）
- [x] `git.auto_commit` 設定の反映

### Phase 6: 仕上げ

- [x] ヘルプオーバーレイ（`?`/`h`キー）
- [x] ステータスメッセージ表示（3秒タイムアウト）
- [x] エラー処理の実装
- [x] テスト作成（コア機能）

### Phase 7: 階層タスク（v0.2.0）

- [x] インデント検出（2スペース = 1階層）
- [x] 親タスク完了 → 子タスク自動完了
- [x] 親タスクアーカイブ → 子タスクも一緒に移動
- [x] specification.md に階層タスク仕様を追記
- [x] Makefile追加（バージョン埋め込みビルド）
- [x] フッターにバージョン表示追加

#### ユーザーテスト項目

- [x] 階層タスクの完了カスケード確認
  - tasks.mdで親タスクを`[x]`にしてttt起動
  - 子タスクも自動で`[x]`と`@done(日付)`が付くか確認
- [x] 階層タスクのアーカイブ確認
  - 古い親タスク（delay_days経過）を`a`キーでアーカイブ
  - 子タスクも一緒にarchive.mdへ移動するか確認
- [x] バージョン表示確認
  - `make install`でビルド・インストール
  - フッター右端にバージョンが表示されるか確認

#### ドキュメンテーションの整備

- [x] development-guideline.md に記載すべきことがあれば追記する。なければ、削除する。
  - CLAUDE.md に開発ガイドラインが記載済みのため、空の development-guideline.md を削除
- [x] ユーザー向けガイドを英語で記載する。必要であれば、Plan Modeで計画を立てること。
  - README.md に英語でユーザーガイドを記載

#### テスト完了後（Claude作業）

- [x] 変更をコミット
- [x] PR作成（`gh pr create`）
- [x] v0.2.0 タグ作成・GitHub Release 公開
- [x] CI ワークフロー追加（test + golangci-lint）
- [x] README にバッジ追加（CI, Go, License, Release）
- [x] LICENSE ファイル追加（MIT）
- [x] CHANGELOG.md 追加

#### 長期テスト(ユーザー作業)

- [ ] 長期間使用して問題ないか検証する

#### 公開用β版の準備

- [ ] Go言語の慣習に基づいたリリースの準備
- [ ] Homebrewでのパッケージリリースの準備
