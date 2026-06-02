# roster get 送信と受信パース

- Priority: High
- Created: 2026-06-02
- Completed:
- Model: Composer 2.5
- Branch: feature/add-roster
- Polished:

## 目的

RFC 6121 Section 2 に沿い、bind 後に roster を取得する IQ を送り、result の `<query xmlns='jabber:iq:roster'>` から連絡先 item をパースできるようにする。

## 優先度根拠

High — IM クライアントの定番フロー (Example 1 / 2)。0010 の `IQEvent.Payload` を活用する。

## 現状

- roster get の XML 生成ヘルパがない
- `IQEvent.Payload` の item 抽出は利用者側

## 設計方針

- `RosterIQGetBytes(from, id)` / `Connection.SendRosterGet(id)`
- `RosterItem` と `ParseRosterItems(iqPayload)` — `jid`, `name`, `subscription`, `groups`
- `elementAttrValue` を `features.go` に追加し item / query 属性を読む
- roster set / push は別 issue

## 完了条件

- RFC 6121 Example 1 / 2 の conformance テストが通る
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/roster.go`: 送信バイト列と `ParseRosterItems`
- `protocol/features.go`: `elementAttrValue`
- `protocol/conn.go`: `SendRosterGet`
- `protocol/roster_test.go`
