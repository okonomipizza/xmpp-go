# 変更履歴

## develop

- [ADD] 増分 XML パーサー `xmlstream` パッケージを追加する
  - @okonomipizza
- [ADD] c2s 向け Sans I/O 状態機械 `protocol` パッケージを追加する
  - @okonomipizza
- [ADD] STARTTLS 交渉 (`StartTLS`, `ResetAfterTLS`) を `protocol` に追加する
  - @okonomipizza
- [ADD] SASL PLAIN 認証と challenge/response パスを `protocol` に追加する
  - @okonomipizza
- [ADD] resource binding (`Bind`, `BindResource`) を `protocol` に追加する
  - @okonomipizza
- [ADD] outbound stanza 送信 (`SendMessage`, `SendPresence`, `SendIQ`) を `protocol` に追加する
  - @okonomipizza

### misc

- [ADD] 変更履歴ファイル `CHANGES.md` を追加する
  - @okonomipizza
- [UPDATE] `AGENTS.md` を xmpp-go 向けに整備し、レビューフローと issue 運用を追記する
  - @okonomipizza
- [UPDATE] `CLAUDE.md` に「変更をレビュー」手順への参照を追記する
  - @okonomipizza
- [ADD] ファイルベース issue 管理の雛形 (`issues/`) を追加する
  - @okonomipizza
