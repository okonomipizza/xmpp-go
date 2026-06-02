# Agents

- Premature Optimization is the Root of All Evil
- If it hurts, do it more often
- Don't live with broken windows
- 妥協しないこと
- 忖度しないこと
- 常に日本語を利用すること
- 全角と半角の間には半角スペースを入れること
- 絵文字を使わないこと
- コメントは全て日本語にすること (テストコードは除く)
- ログメッセージは全て英語にすること
- エラーメッセージは全て英語にすること
- テストコードは全て英語にすること (日本語テストケースを除く)


## プロジェクト概要

- Go で実装された **依存 0** (本番コードは標準ライブラリのみ) かつ **Sans I/O** な XMPP ライブラリ
- **クライアント実装のみ** — サーバー (S2S) は対象外
- c2s (client-to-server) 接続のみサポート
- ネットワーク・TLS・時間は呼び出し側の責務。本ライブラリはバイト列と状態機械のみ扱う

### パッケージ構成

| パッケージ | 役割 |
|-----------|------|
| `jid/` | RFC 7622 JID パース |
| `xmlstream/` | 増分 XML パーサー (Sans I/O 入力層) |
| `protocol/` | c2s プロトコル状態機械 |

### 開発環境

```bash
nix develop          # Go, staticcheck など
go test ./...        # または make test
make test-pbt        # Property-Based Test のみ
make test-fuzz       # Fuzzing (FUZZ_TIME=30s がデフォルト)
make lint            # staticcheck
```

- RFC 原文は `refs/` 以下 (`rfc6120.txt`, `rfc6121.txt`, `rfc7622.txt`)


## レビューについて

- 常に日本語を利用すること
- レビューはかなり厳しくすること
- 表現はシンプルに、指摘は明確かつ具体的にすること
- 指摘には優先順位をつけ、重要なものから順に記載すること
- ドキュメントは別に書いているので、ドキュメントについては考慮しないこと
- 変更点と `CHANGES.md` の整合性を確認すること


## レビューフロー

実装とレビューを分離する。issue の `Model` には実装担当 (例: Composer 2.5) を記載する。

```
[Composer 2.5 等] 実装 + テスト + CHANGES.md (develop)
        ↓
[プレコミット] 下記チェックリストを満たす (コミットはまだしない)
        ↓
[Claude Code / Codex] リポジトリで「変更をレビュー」→ 指摘対応 → 必要なら再レビュー
        ↓
[人間] コミット (AGENTS.md のコミット規約に従う)
```

### プレコミットチェックリスト

- [ ] `nix develop -c go test ./...` が通る
- [ ] `nix develop -c staticcheck ./...` (または `make lint`) が通る
- [ ] 触ったパッケージで PBT / fuzz が必要なら実行済み
- [ ] `CHANGES.md` の `## develop` に変更を追記済み
- [ ] `AGENTS.md` 準拠 (コメント日本語、エラー・ログ英語、モック・スタブなし)
- [ ] issue 駆動なら issue の完了条件を満たしている

### 「変更をレビュー」

リポジトリを開いた状態で、ユーザーが **「変更をレビュー」** (同義の短い依頼含む) と言ったら、レビューのみ行う。コミット・コード修正・issue の close はしない。

エージェントが自ら読む・実行するもの:

1. 本ファイル (`AGENTS.md`) の「レビューについて」「プレコミットチェックリスト」
2. `git status` と `git diff` (未コミット変更。マージ済みならユーザーに確認)
3. `CHANGES.md` の `## develop` — 差分との整合
4. 変更パスに対応する `*_test.go` / `pbt_test.go` / `fuzz_test.go`
5. `issues/` 直下のオープン issue のうち、変更と関連するもの (なければスキップ)
6. 必要なら `refs/rfc*.txt` の該当節のみ (全文を読まない)

レビュー観点 (重要度順):

1. `AGENTS.md` 違反 (Sans I/O, テスト方針, エラーメッセージ, モック禁止)
2. issue の完了条件・RFC とのズレ
3. プレコミット未達 (テスト・lint 未実行の疑いがあれば `go test` / `staticcheck` を実行して確認)
4. 設計・可読性

スコープ外: `README.md` 等のドキュメント、`issues/0000-template.md`、無関係パッケージ。

出力: 日本語、優先度付きの指摘。問題なければプレコミット可否を明示する。


## コミットについて

- 勝手にコミットしないこと
- 全てのテストが通らない限りコミットしないこと
- コミットメッセージは確認すること
- コミットメッセージは日本語で書くこと
- コミットメッセージは命令形で書くこと
- コミットメッセージは〜するという形で書くこと
- フックをスキップしないこと


## issues について

タスク・バグ・設計判断は `issues/` 以下の Markdown で管理する。

### ディレクトリ構成

```
issues/
  SEQUENCE              # 次に採番する番号 (改行のみ、例: 0001)
  0000-template.md      # 雛形 (採番しない)
  {seqnum}-....md       # オープンな issue
  closed/               # 完了した issue
  pending/              # 保留中の issue
```

### 作成・更新

- `0000-template.md` の構成・メタデータ形式に従うこと
- 番号が小さい issue から順に対応すること
- ファイル名: `{seqnum}-{category}-{short-description}.md`
  - seqnum は `issues/SEQUENCE` の値 (作成後に +1。9999 超えたら 5 桁)
  - 例: `0001-bug-xmlstream-nested-close.md`
- メタデータ (タイトル直下の箇条書き): `Priority`, `Created`, `Completed`, `Model`, `Branch`, `Polished`
  - `Priority` は High / Medium / Low
  - 未完了なら `Completed` は空欄または行ごと省略してよい
  - `Branch` は対応ブランチ名 (Git Flow の `feature/...`)
- 本文セクション: 目的、優先度根拠、現状、設計方針、完了条件、解決方法 (テンプレート参照)
- issue を作成したらコミットすること (メッセージに番号とタイトルを含める)
- 1 issue 完了ごとに 1 コミットすること
- 対応が難しい場合は `issues/pending/` へ `git mv` すること

### git ブランチの命名規則

- Git Flow を使うこと
- バグ修正: `feature/fix-`
- 機能追加: `feature/add-`
- 後方互換のない変更: `feature/change-`
- リファクタリング: `feature/refactor-`
- ブランチ名に issue の番号を含めないこと

### issue が実は解決してなかった場合

- reopen の理由を issue に書き、`issues/closed` から `issues/` へ `git mv` すること
- 何がどう解決していなかったのかを明確にすること

### バグが見つかった場合

- `issues/` 以下に markdown で登録すること
- 再現手順と、分かる範囲の環境・入力データを記載すること

### バグを修正した場合

- 修正内容を issue に記載し、`issues/closed` へ `git mv` すること
- 「## 解決方法」セクションに何をどう修正したかを明記すること

### 設計判断が必要な issue の場合

- 外部依存の追加や設計判断が必要なものは `issues/pending/` に置くこと
- pending にした理由を issue に明記すること
- pending の issue は修正せずそのまま残す (close しない)


## 変更履歴について

- 変更履歴は `CHANGES.md` に記載すること
- 変更の種別は以下の 4 つを使うこと
  - `[CHANGE]`: 後方互換のない変更
  - `[ADD]`: 後方互換がある追加
  - `[UPDATE]`: 後方互換がある変更
  - `[FIX]`: バグ修正
- エントリは種別の順番を守って記載すること (CHANGE → ADD → UPDATE → FIX)
- 機能に直接影響しない変更 (ドキュメント追加、リファクタリング等) は `### misc` サブセクションに記載すること
- 未リリースの変更は `## develop` セクションに追記すること
- 各エントリは `- [種別] 変更内容を〜するという形で書く` というフォーマットにすること
- 各エントリの担当者はエントリの次の行に、変更内容より 2 文字分インデントして `- @ユーザー名` の形式で記載すること
- 変更内容の説明は日本語で書くこと
- リリース時は `## develop` を `## バージョン` に変更し、`**リリース日**: YYYY-MM-DD` を記載すること
- 変更履歴は派生元ブランチとの最終的な差分のみを記載すること (開発ブランチ内の中間修正は載せない)


## ライブラリ・依存

- **本番コード (`*.go` の library)** は標準ライブラリのみ。外部パッケージを増やさないこと
- **テストのみ** `pgregory.net/rapid` を PBT に使用する (`go.mod` の `require`)
- TLS / SASL / TCP は本ライブラリに含めない (transport 層は利用者側)
- ログライブラリは使わない (将来必要ならその時点で issue で設計判断)

## テスト

### 方針

- **モックやスタブは使わないこと** (Sans I/O の入出力を直接渡す)
- Property-Based Testing を重視する
- 任意のバイト列で panic しないことを Fuzzing で保証する
- RFC 記載のサンプルを conformance test として使用する

### ツール

| 用途 | ツール |
|------|--------|
| PBT | `pgregory.net/rapid` (`rapid.Check`) |
| Fuzzing | Go 組み込み (`go test -fuzz=...`, `testing.F`) |
| 静的解析 | `staticcheck` (`make lint`) |
| カバレッジ | `go test -coverprofile=...` (`go tool cover`) |

### 命名規則

- `TestXxx` — 通常のユニットテスト
- `TestXxx_Property` — Property-Based Test
- `FuzzXxx` — Go 組み込み Fuzzing

### ファイル構成

各パッケージ内でテストファイルを分離する:

- `xxx_test.go` — ユニットテスト・conformance test
- `pbt_test.go` — Property-Based Test
- `fuzz_test.go` — Fuzzing

`pbt_test.go` に unittest だけを書かないこと。PBT で足りる性質は PBT に寄せる。

### 役割分担

- **PBT**: Strategy に基づく入力生成とプロパティ検証 (ラウンドトリップ等)
- **Fuzzing**: 任意入力に対するパニック安全性
- **単体テスト**: 意図的なエラーパス、RFC サンプル、境界値など PBT だけでは扱いにくいケース
- PBT に「任意入力でパニックしないだけ」を書かない (それは Fuzzing の役割)

### カバレッジ

- `protocol` パッケージは **90% 以上** を目標とする
- 未カバー行の扱い:
  - 正常系 → PBT または unittest を追加
  - エラーパス → unittest または fuzz
  - 到達不能 → デッドコードとして削除を検討

```bash
# パッケージ単位の例
go test -coverprofile=coverage.out ./protocol/...
go tool cover -func=coverage.out
```

### Fuzzing の実行例

```bash
make test-fuzz
# 個別: go test -fuzz=FuzzParserFeed -fuzztime=30s ./xmlstream/...
```
