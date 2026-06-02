# SASL 交渉を protocol に追加する

- Priority: High
- Created: 2026-06-02
- Completed: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-sasl-negotiation

## 目的

RFC 6120 Section 6 の SASL 交渉を Sans I/O 状態機械に組み込む。

## 優先度根拠

認証は c2s 接続の必須ステップであり、resource bind の前提のため High。

## 現状

STARTTLS 後の features で mechanisms が提供されるが、`<auth/>` / challenge / success は未処理。

## 設計方針

- PLAIN の初期応答を内蔵
- SCRAM 等は challenge/response をイベント化し呼び出し側が応答を生成
- SCRAM クライアント実装自体は含めない

## 完了条件

- PLAIN 成功フローのテスト
- challenge/response パスのテスト
- `go test ./...` / `staticcheck` が通る
- `protocol` カバレッジ 90% 以上を維持

## 解決方法

- `protocol/sasl.go`, `protocol/xmltext.go`: `<auth/>` / `<response/>` 生成、PLAIN 初期応答、mechanism 選択
- `protocol/features.go`: mechanisms / bind の解析
- `protocol/conn.go`: `Authenticate`, `SASLResponse`, `ResetAfterSASL()`, `StateAwaitSASLOutcome`, SASL イベント (failure は `StateNegotiating` に戻す)
- `protocol/config.go`: `Password`, `SASLMechanisms`
- `protocol/sasl_test.go`, `conn_test.go` の SASL conformance テスト
