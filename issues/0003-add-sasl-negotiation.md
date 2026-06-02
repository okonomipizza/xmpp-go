# SASL 交渉を protocol に追加する

- Priority: High
- Created: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-starttls-negotiation

## 目的

RFC 6120 Section 6 の SASL 交渉を Sans I/O 状態機械に組み込む。

## 設計方針

- PLAIN の初期応答を内蔵
- SCRAM 等は challenge/response をイベント化し呼び出し側が応答を生成
- SCRAM クライアント実装自体は含めない

## 完了条件

- PLAIN 成功フローのテスト
- challenge/response パスのテスト

## 解決方法

(完了時に記載)
