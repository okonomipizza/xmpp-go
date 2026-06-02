# Sans I/O 向け xmlstream と protocol の骨格を追加する

- Priority: High
- Created: 2026-06-02
- Completed: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-xmlstream-protocol

## 目的

XMPP クライアント (c2s) のコアを Sans I/O 状態機械として実装する。transport から独立し、増分 XML パースとストリーム開始フェーズを扱えるようにする。

## 優先度根拠

ライブラリの基盤であり、以降の TLS / SASL / stanza 処理の前提となるため High。

## 現状

`jid` のみ存在。README が示す `protocol` パッケージは未実装。

## 設計方針

- `xmlstream`: 自前の増分 XML パーサー (`Feed` / `Next` / `ErrNeedMore` / `NeedInput`)
- `protocol`: `Connection` 状態機械、`Receive` / `BytesToSend` / イベント型
- 標準ライブラリのみ (テストは `rapid`)
- 交渉中の TLS/SASL 子要素はイベント化せず、後続 issue で拡張

## 完了条件

- RFC 6120 のストリーム開始・features の conformance test が通る
- `go test ./...` と `staticcheck` が通る
- `protocol` のカバレッジが 90% 以上
- Fuzz / PBT でパニックしない

## 解決方法

- `xmlstream` パッケージ: 増分 XML パーサー、RFC 6120 conformance test、Fuzz / PBT
- `protocol` パッケージ: c2s `Connection` 状態機械、ストリーム開始・features・`ElementEvent` / `StanzaEvent`、`State.String()`、`xml:lang` 検証
- `Makefile`: 各パッケージの Fuzz ターゲットを追加
