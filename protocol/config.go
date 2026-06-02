package protocol

import "github.com/okonomipizza/xmpp-go/jid"

// Config はクライアント接続 (c2s) の設定を表す。
type Config struct {
	// JID は接続するユーザーの bare JID (local@domain)。
	JID jid.JID

	// Lang はストリームのデフォルト言語 (xml:lang)。空なら "en"。
	Lang string

	// Resource は bind 時に送る resourcepart。空ならサーバー生成 (RFC 6120 Section 7.6.1)。
	Resource string

	// Password は SASL PLAIN 用パスワード。
	Password string

	// SASLMechanisms は優先する SASL メカニズム名の順序。空なら PLAIN のみ (内蔵初期応答があるもの)。
	SASLMechanisms []string
}
