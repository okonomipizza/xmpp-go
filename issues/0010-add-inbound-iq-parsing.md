# 受信 IQ の構造化 (IQEvent)

- Priority: High
- Created: 2026-06-02
- Completed:
- Model: Composer 2.5
- Branch: feature/add-inbound-iq-parse
- Polished:

## 目的

`StateReady` / `StateClosing` で受信した `<iq/>` から `from` / `to` / `id` / `type` と子要素ペイロードを抽出し、roster や ping 応答を扱いやすくする。

## 優先度根拠

High — message / presence と同様、stanza 受信の最後の未構造化部分。後続の roster 等の前提。

## 現状

- 受信 IQ は `StanzaEvent` + 生 `Token` のみ
- bind IQ は `BindSuccessEvent` / `BindFailureEvent` で別処理 (変更なし)

## 設計方針

- `IQEvent` を新設し、Ready / Closing の `<iq/>` はこれを返す
- `ParseInboundIQ(tok)` で属性と `elementInnerBytes` 由来の `Payload` を返す
- `StanzaEvent` は削除 (iq 専用だったため)

## 完了条件

- RFC 6121 Example 2 (roster result) 相当の conformance テストが通る
- 空の `<iq/>` もエラーなく解析できる
- bind パスは従来どおり
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/iq.go`: `IQEvent`, `ParseInboundIQ`
- `protocol/conn.go`: `handleIQ` の Ready / Closing 分岐
- `protocol/iq_test.go`, `conn_test.go` 更新、`StanzaEvent` 削除
