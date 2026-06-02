# unavailable presence 送信

- Priority: Medium
- Created: 2026-06-02
- Completed:
- Model: Composer 2.5
- Branch: feature/add-unavailable-presence
- Polished:

## 目的

RFC 6121 Section 4.5 / Example 15 に沿い、ログアウト時の `type='unavailable'` presence を `status` 付きで送信できるようにする。`Close` の前にセッションを offline にする定番フロー。

## 優先度根拠

Medium — 0008 (available) の対。差分は小さく、graceful ログアウトに必要。

## 現状

- `PresenceStanzaBytes` で `type='unavailable'` は属性のみ (子要素なし)
- `SendAvailablePresence` はあるが unavailable 専用 API がない

## 設計方針

- `UnavailablePresenceOpts` (`Lang`, `Status`)。`to` は directed 用に引数で渡し、空なら broadcast
- `UnavailablePresenceStanzaBytes(from, to, opts)` は常に `type='unavailable'`
- `status` は `escapeXMLText` のみ (0008 と同様、引用符は許可)
- `Connection.SendUnavailablePresence(to, opts)`

## 完了条件

- RFC 6121 Example 15 相当の conformance テストが通る
- directed (`to` あり) も生成できる
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/stanza.go`: `UnavailablePresenceOpts`, `UnavailablePresenceStanzaBytes`
- `protocol/conn.go`: `SendUnavailablePresence`
- `protocol/stanza_test.go` にテスト追加
