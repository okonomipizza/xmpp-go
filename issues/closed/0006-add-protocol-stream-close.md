# クライアントからのストリーム終了送信

- Priority: High
- Created: 2026-06-02
- Completed: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-stream-close

## 目的

RFC 6120 Section 4.4 に沿い、クライアントが `</stream:stream>` を送信してセッションを graceful に終了できるようにする。受信側の `StreamClosedEvent` のみでは接続ライフサイクルが閉じない。

## 優先度根拠

High — 実クライアントではログアウト・切断時に必須。差分が小さく bind / stanza 送信と並べて c2s の終端を揃えられる。

## 現状

- `Receive` で `</stream:stream>` を受信すると `StreamClosedEvent` となり `StateClosed` になる
- クライアントからの close タグをキューに載せる API がない
- close 送信後も `SendMessage` 等が呼べる余地がある

## 設計方針

- `StreamCloseBytes()` で RFC 6120 の `</stream:stream>` バイト列を返す
- `Connection.Close()` で送信キューに載せ、`StateClosing` に遷移する
- `StateClosing` では受信 (`Receive`) は継続し、送信系 API は拒否する
- サーバーから `</stream:stream>` を受信したら `StateClosed` と `StreamClosedEvent` (既存)

## 完了条件

- `StateReady` から `Close()` 後、`BytesToSend()` で `</stream:stream>` が得られる
- close 送信後は stanza 送信がエラーになる
- サーバー close 受信後は `StateClosed` となり、以降の `Receive` はエラーになる
- `go test ./protocol/...` と staticcheck が通る
- `CHANGES.md` の develop に追記済み

## 解決方法

- `protocol/stream.go`: `StreamCloseBytes`
- `protocol/state.go`: `StateClosing`
- `protocol/conn.go`: `Close()`。`StateClosing` 中は inbound stanza を `StanzaEvent` として処理 (RFC 6120 Section 4.4)
- `protocol/conn_test.go`, `protocol/stream_test.go`, `protocol/state_test.go`, `protocol/pbt_test.go`
