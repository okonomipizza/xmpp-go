# STARTTLS 交渉を protocol に追加する

- Priority: High
- Created: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-starttls-negotiation

## 目的

RFC 6120 Section 5 の STARTTLS 交渉を Sans I/O 状態機械に組み込む。transport 層の TLS ハンドシェイは呼び出し側が行い、プロトコル層はコマンド送信と `<proceed/>` / `<failure/>` のイベント化を担う。

## 優先度根拠

TLS は SASL の前提であり、c2s 接続の必須ステップのため High。

## 現状

`protocol` はストリーム開始と `<stream:features/>` まで。交渉子要素は黙殺されている。

## 設計方針

- `ParseStreamFeatures` で STARTTLS 提供・必須を判定
- `StartTLS()` で `<starttls/>` を送信キューへ
- `<proceed/>` で `StartTLSProceedEvent`、transport で TLS 後に `ResetAfterTLS()` + 再度 `Start()`
- TLS 実装は含めない

## 完了条件

- RFC 6120 Section 5.4.2 のサンプルに沿ったテストが通る
- `go test ./...` / `staticcheck` が通る
- `protocol` カバレッジ 90% 以上を維持

## 解決方法

(完了時に記載)
