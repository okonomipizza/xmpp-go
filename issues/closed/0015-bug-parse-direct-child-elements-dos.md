# parseDirectChildElements の不正 XML による DoS

- Priority: High
- Created: 2026-06-03
- Completed: 2026-06-03
- Model: Composer 2.5
- Branch: feature/add-roster-set-and-push
- Polished: 2026-06-03

## 目的

`protocol/features.go` の `parseDirectChildElements` が不正 XML で無限ループし、任意入力による DoS になる問題を修正する。

## 優先度根拠

High — `ParseRosterItems` / roster push など複数経路で共通利用され、Fuzz で再現済み。

## 現状

- 非 self-closing 要素で閉じタグが見つからないと外側の `i` が進まない
- 不完全な `</...>` (終端 `>` なし) で `i++` のみだと同じ `<` に戻り外側ループが止まらない
- 内側の深さ探索で `j` が進まないケースがある

## 再現

- `FuzzRosterParsing` で約 11s 後にハング
- 単体: `ParseRosterItems([]byte("<query ''><''></query>"))` が停止
- corpus: `protocol/testdata/fuzz/FuzzRosterParsing/af57aafd0ebe4fef`, `9af5fe10a8e08f60`

## 設計方針

- 閉じタグ未完了時は外側 `i` を必ず進める / 不完全 `</` は inner 末尾までスキップ
- 深さ > 0 のまま終了したら開始タグ直後へスキップ
- 閉じタグ名不一致時は return せず `j` を進めて継続
- 内側ループで `j` が進まなければ `j++`

## 完了条件

- 上記 corpus で 1 秒以内に終了する回帰テストが通る
- `FuzzRosterParsing` が 30s ハングしない
- `go test ./...` / staticcheck 通過

## 解決方法

- `protocol/features.go` (`parseDirectChildElements`): 不完全閉じタグで `i = len(inner)`、深さ未解消で `i = openEnd+1`、不一致閉じタグで `j` 進行、`j` 停滞時 `j++`
- `protocol/features_test.go`, `protocol/roster_test.go`: 回帰テスト
- `protocol/fuzz_test.go`: `FuzzRosterParsing`
- `protocol/testdata/fuzz/FuzzRosterParsing/`: fuzz corpus 2 件
- `Makefile`: `test-fuzz` に `FuzzRosterParsing` を追加
