# presence 購読 (subscribe / subscribed 等)

- Priority: High
- Created: 2026-06-02
- Completed:
- Model: Composer 2.5
- Branch: feature/add-presence-subscription
- Polished:

## 目的

RFC 6121 Section 3.1 / 3.2 / 3.3 に沿い、presence 購読の送受信を扱いやすくする。受信は既存 `PresenceEvent.Type` で表現できるが、送信は `SendPresence` の生 `typ` 任せになっている。

## 優先度根拠

High — roster と並べて 6121 の次マイルストーン。IM クライアントで連絡先追加・承認に必要。

## 現状

- `SendPresence(to, id, typ)` で `type='subscribe'` は生成可能 (テストあり)
- 購読専用 API・`typ` 検証・RFC 例の conformance が未整理
- 受信 `subscribe` / `subscribed` / `unsubscribed` は `PresenceEvent` で足りる (追加イベント型は不要)

## 設計方針

- `SendPresenceSubscribe`, `SendPresenceSubscribed`, `SendPresenceUnsubscribe`, `SendPresenceUnsubscribed` を `SendPresence` の薄いラッパーとする
- 許可する `type` を定数化し、誤った typ を拒否
- `to` は bare JID 必須 (購読系 stanza)

## 完了条件

- RFC 6121 の subscribe / subscribed / unsubscribed 送信例に沿ったテストが通る
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/presence_sub.go` (または `stanza.go`): ラッパーと `validateSubscriptionPresenceType`
- `protocol/conn.go`: 上記 4 メソッド
- `protocol/presence_sub_test.go`
