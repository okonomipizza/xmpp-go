package protocol

import (
	"fmt"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// MessageEvent は <message/> を受信し、主要属性と body を解析したことを表す。
type MessageEvent struct {
	From  jid.JID
	To    jid.JID
	ID    string
	Type  string
	Lang  string
	Body  string
	Token xmlstream.Token
}

func (*MessageEvent) isEvent() {}

// ParseInboundMessage は message トークンから RFC 6121 Section 5.2 の主要フィールドを抽出する。
func ParseInboundMessage(tok xmlstream.Token) (MessageEvent, error) {
	if tok.Name != "message" {
		return MessageEvent{}, fmt.Errorf("protocol: not a message stanza")
	}
	from, err := parseStanzaJIDAttr(tok.AttrValue("from"))
	if err != nil {
		return MessageEvent{}, err
	}
	to, err := parseStanzaJIDAttr(tok.AttrValue("to"))
	if err != nil {
		return MessageEvent{}, err
	}
	return MessageEvent{
		From:  from,
		To:    to,
		ID:    tok.AttrValue("id"),
		Type:  tok.AttrValue("type"),
		Lang:  tok.AttrValue("xml:lang"),
		Body:  messageBody(tok.Raw),
		Token: tok,
	}, nil
}

func parseStanzaJIDAttr(attr string) (jid.JID, error) {
	if attr == "" {
		return jid.JID{}, nil
	}
	return jid.Parse(attr)
}

func messageBody(messageRaw []byte) string {
	for _, ch := range parseDirectChildElements(messageRaw) {
		if ch.name == "body" {
			return unescapeXMLText(elementTextContent(ch.raw))
		}
	}
	return ""
}
