# 受信 message の構造化 (MessageEvent)

- Priority: High
- Created: 2026-06-02
- Completed:
- Model: Composer 2.5
- Branch: feature/add-inbound-message-parse
- Polished:

## 目的

`StateReady` / `StateClosing` で受信した `<message/>` から `from` / `to` / `type` / `xml:lang` / `<body>` を抽出し、アプリが生 `Token` を触らずに本文を扱えるようにする。

## 優先度根拠

High — outbound stanza (0005) と対になり、IM クライアントの最小ループ (受信表示) に必要。README が想定する `MessageEvent` とも整合する。

## 現状

- 受信 message は `StanzaEvent` のみ (`Name` + `Token`)
- 属性・body のパースは利用者側

## 設計方針

- `MessageEvent` を新設し、`<message/>` 受信時は `StanzaEvent` ではなくこれを返す
- `ParseInboundMessage(tok)` で属性と最初の `<body/>` 子要素を解析 (RFC 6121 Section 5.2)
- `from` / `to` は `jid.Parse`。不正 JID はエラー
- presence / iq は引き続き `StanzaEvent` (別 issue で拡張可)

## 完了条件

- RFC 6121 Example 9 の message 1 件で body / from / to / type / lang が一致するテストが通る
- 不正 `from` でエラー
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/message.go`: `ParseInboundMessage`, body 抽出
- `protocol/event.go`: `MessageEvent`
- `protocol/conn.go`: message 分岐で `MessageEvent` を返す
- `protocol/message_test.go`, `conn_test.go` 更新
