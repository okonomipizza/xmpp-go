package protocol

import "github.com/okonomipizza/xmpp-go/jid"

// Config はクライアント接続 (c2s) の設定を表す。
type Config struct {
	// JID は接続するユーザーの bare JID (local@domain)。
	JID jid.JID

	// Lang はストリームのデフォルト言語 (xml:lang)。空なら "en"。
	Lang string
}
