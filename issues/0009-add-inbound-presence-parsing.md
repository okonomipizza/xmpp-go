# 受信 presence の構造化 (PresenceEvent)

- Priority: High
- Created: 2026-06-02
- Completed:
- Model: Composer 2.5
- Branch: feature/add-inbound-presence-parse
- Polished:

## 目的

`StateReady` / `StateClosing` で受信した `<presence/>` から `from` / `to` / `type` / `show` / `status` / `priority` 等を抽出し、アプリが生 `Token` を触らずに扱えるようにする。

## 優先度根拠

High — 0008 (初期 presence 送信) の対。連絡先の online / away / probe を扱うには受信側の構造化が必要。

## 現状

- 受信 presence は `StanzaEvent` のみ
- 0007 の `MessageEvent` と非対称

## 設計方針

- `PresenceEvent` を新設し、`<presence/>` 受信時は `StanzaEvent` ではなくこれを返す
- `ParseInboundPresence(tok)` で属性と子要素 (`show`, `status`, `priority`) を解析
- 子要素テキストは `unescapeXMLText` で復元 (0007 と同様)
- `iq` は引き続き `StanzaEvent`

## 完了条件

- RFC 6121 Example 11 相当の conformance テストが通る
- `type='probe'` 等の属性のみ presence も解析できる
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/presence.go`: `PresenceEvent`, `ParseInboundPresence`
- `protocol/conn.go`: presence 分岐を更新
- `protocol/presence_test.go`, `conn_test.go` 更新
