# c2s 向け outbound stanza 送信を protocol に追加する

- Priority: High
- Created: 2026-06-02
- Completed: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-outbound-stanzas

## 目的

`StateReady` 以降に `<message/>`, `<presence/>`, `<iq/>` を送信キューへ載せられるようにする。bind 完了後の通常 IM フローで必須。

## 優先度根拠

受信 (`StanzaEvent`) のみではクライアントとして使えない。RFC 6121 の基本メッセージ送信ができないため High。

## 現状

`protocol` は stanza 受信まで。送信は交渉コマンド (TLS, SASL, bind) のみ。

## 設計方針

- `MessageStanzaBytes`, `PresenceStanzaBytes`, `IQStanzaBytes` で XML 生成 (属性値の検証含む)
- `Connection.SendMessage` / `SendPresence` / `SendIQ` は `StateReady` のみ、`from` は `BoundJID` (full JID) を使用
- 本文・子要素は最小限 (message の `<body/>`, iq は呼び出し側が渡す子 XML)

## 完了条件

- RFC 6121 Section 5.2.1 の message 例に沿った生成テストが通る
- `go test ./...` / `staticcheck` が通る
- `protocol` カバレッジ 90% 以上を維持

## 解決方法

- `protocol/stanza.go`: `MessageStanzaBytes`, `PresenceStanzaBytes`, `IQStanzaBytes`
- `protocol/xmltext.go`: `escapeXMLText` (message body の XML エスケープ)
- `protocol/conn.go`: `SendMessage`, `SendPresence`, `SendIQ`, `sendStanza`
- `protocol/stanza_test.go`, `conn_test.go` の conformance / Connection 経由テスト
