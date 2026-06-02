# roster set / remove と push 受信

- Priority: High
- Created: 2026-06-02
- Completed: 2026-06-02
- Model: Composer 2.5
- Branch: feature/add-roster-set-and-push
- Polished:

## 目的

RFC 6121 Section 2.1.5 / 2.1.6 に沿い、roster の変更 (set / remove) を送り、サーバーからの roster push (`iq type='set'`) を `ParseRosterItems` で処理できるようにする。0012 の get だけでは連絡先追加・サーバー更新を扱えない。

## 優先度根拠

High — 0013 (presence 購読) とセットで 6121 の連絡先管理の芯。0012 の直後が自然。

## 現状

- roster **get** + `ParseRosterItems` のみ
- item の XML 生成・`iq type='set'` 送信ヘルパなし
- push 受信は `IQEvent` + 手動 `ParseRosterItems` で可能だが専用 API なし

## 設計方針

- `RosterSetItemBytes` / `RosterQuerySetBytes` で `<item/>` を組み立て (`jid`, `name`, `subscription`, `group`; outbound に `ask` は含めない)
- `RosterIQSetBytes(from, id, RosterSetItem)` / `Connection.SendRosterSet` (item は 1 件のみ、RFC 2.1.5)
- remove は `subscription='remove'` の item 1 件 (RFC 2.1.6)
- push 受信は `ParseRosterPush` (`from` 検証・item 1 件、RFC 2.1.6)

## 完了条件

- roster set / remove の conformance テストが通る
- push payload のパーステスト (Example 3.1.5 相当の roster push 断片) が通る
- `go test ./...` / staticcheck 通過、`protocol` カバレッジ 90% 以上
- `CHANGES.md` develop に追記

## 解決方法

- `protocol/roster.go` 拡張: `RosterSetItemBytes`, `RosterIQSetBytes`, `SendRosterSet`, `SendRosterRemove`, `ParseRosterPush`
- `protocol/roster_test.go` 追加

## 依存

- **0013 より先に 0014 でも可** — 連絡先追加フローは roster set → presence subscribe の順が多い
- 実装順の推奨: **0014 → 0013** または並行
