# Resource binding を protocol に追加する

- Priority: High
- Created: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-resource-binding

## 目的

RFC 6120 Section 7 の resource binding を Sans I/O 状態機械に組み込む。SASL 成功後の `<iq type='set'>` / `<bind/>` 送信と、`<iq type='result'>` に含まれる full JID のイベント化を行う。

## 優先度根拠

resource binding 完了まで stanza を送れない。手動 `SetReady()` は本番フローと乖離するため High。

## 現状

`ParseStreamFeatures` は `BindOffered` を判定するのみ。`<bind/>` の IQ 送受信・`StateReady` への遷移は `SetReady()` の手動呼び出しに依存している。

## 設計方針

- `Bind()` / `BindResource(resource)` で bind IQ を送信キューへ載せる (resource 空ならサーバー生成、RFC 7.6.1)
- `StateAwaitBind` で `<iq type='result'>` / `<iq type='error'>` を処理
- 成功で `BindSuccessEvent` と `StateReady`、失敗で `BindFailureEvent` と `StateNegotiating` (リトライ可能)
- `SetReady()` は bind なしのテスト・後方互換用に残す

## 完了条件

- RFC 6120 Section 7.6.1 / 7.7.1 のサンプルに沿ったテストが通る
- `go test ./...` / `staticcheck` が通る
- `protocol` カバレッジ 90% 以上を維持

## 解決方法

(完了時に記載)
