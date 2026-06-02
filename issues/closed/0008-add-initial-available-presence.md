# 初期 presence (available) 送信

- Priority: High
- Created: 2026-06-02
- Completed: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-initial-available-presence

## 目的

RFC 6121 Section 4.2.1 に沿い、bind 直後に送る broadcast 初期 presence (`to` なし、`type` なし) を `show` / `status` / `priority` / `xml:lang` 付きで送信できるようにする。

## 優先度根拠

High — ログイン完了後の定番フロー。既存 `SendPresence` は directed / subscribe 向けで子要素未対応。

## 現状

- `PresenceStanzaBytes` は属性のみの空要素または `type='unavailable'` 等
- RFC Example 3 (`<presence/>`) および Example 10 (`<show>`, `<status>`, `<priority>`) を生成できない

## 設計方針

- `AvailablePresenceOpts` で任意フィールドを指定
- `AvailablePresenceStanzaBytes(from, opts)` で子要素を XML エスケープして組み立て
- `Connection.SendAvailablePresence(opts)` は `boundJID` を `from` に使う
- `show` は非空時のみ `away` / `chat` / `dnd` / `xa` を許可
- `priority` は -128〜127、`nil` なら要素なし

## 完了条件

- RFC 6121 Example 3 相当 (属性 + 子なし) と Example 10 相当の conformance テストが通る
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/stanza.go`: `AvailablePresenceOpts`, `AvailablePresenceStanzaBytes`, `validatePresenceShow` / `validatePresencePriority`
- `protocol/conn.go`: `SendAvailablePresence`
- 属性は `validateStanzaAttr` (`&` 含む)、`<status>` は `escapeXMLText` のみ
- `protocol/stanza_test.go` (RFC 6121 Example 3/10 等)
