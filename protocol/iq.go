package protocol

import (
	"bytes"
	"fmt"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// IQEvent は <iq/> を受信し、主要属性と子要素ペイロードを解析したことを表す。
type IQEvent struct {
	From    jid.JID
	To      jid.JID
	ID      string
	Type    string
	Lang    string
	Payload []byte
	Token   xmlstream.Token
}

func (*IQEvent) isEvent() {}

// ParseInboundIQ は iq トークンから属性と直下の子要素 XML を抽出する。
func ParseInboundIQ(tok xmlstream.Token) (IQEvent, error) {
	if tok.Name != "iq" {
		return IQEvent{}, fmt.Errorf("protocol: not an iq stanza")
	}
	from, err := parseStanzaJIDAttr(tok.AttrValue("from"))
	if err != nil {
		return IQEvent{}, err
	}
	to, err := parseStanzaJIDAttr(tok.AttrValue("to"))
	if err != nil {
		return IQEvent{}, err
	}
	return IQEvent{
		From:    from,
		To:      to,
		ID:      tok.AttrValue("id"),
		Type:    tok.AttrValue("type"),
		Lang:    tok.AttrValue("xml:lang"),
		Payload: iqPayload(tok.Raw),
		Token:   tok,
	}, nil
}

func iqPayload(iqRaw []byte) []byte {
	inner := elementInnerBytes(iqRaw)
	if len(inner) == 0 {
		return nil
	}
	return bytes.Clone(inner)
}
