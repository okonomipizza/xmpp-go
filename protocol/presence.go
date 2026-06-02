package protocol

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/okonomipizza/xmpp-go/jid"
	"github.com/okonomipizza/xmpp-go/xmlstream"
)

// PresenceEvent は <presence/> を受信し、主要属性と子要素を解析したことを表す。
type PresenceEvent struct {
	From     jid.JID
	To       jid.JID
	ID       string
	Type     string
	Lang     string
	Show     string
	Status   string
	Priority *int
	Token    xmlstream.Token
}

func (*PresenceEvent) isEvent() {}

// ParseInboundPresence は presence トークンから RFC 6121 の主要フィールドを抽出する。
func ParseInboundPresence(tok xmlstream.Token) (PresenceEvent, error) {
	if tok.Name != "presence" {
		return PresenceEvent{}, fmt.Errorf("protocol: not a presence stanza")
	}
	from, err := parseStanzaJIDAttr(tok.AttrValue("from"))
	if err != nil {
		return PresenceEvent{}, err
	}
	to, err := parseStanzaJIDAttr(tok.AttrValue("to"))
	if err != nil {
		return PresenceEvent{}, err
	}
	priority, err := presencePriority(tok.Raw)
	if err != nil {
		return PresenceEvent{}, err
	}
	return PresenceEvent{
		From:     from,
		To:       to,
		ID:       tok.AttrValue("id"),
		Type:     tok.AttrValue("type"),
		Lang:     tok.AttrValue("xml:lang"),
		Show:     presenceChildText(tok.Raw, "show"),
		Status:   presenceChildText(tok.Raw, "status"),
		Priority: priority,
		Token:    tok,
	}, nil
}

func presenceChildText(presenceRaw []byte, name string) string {
	for _, ch := range parseDirectChildElements(presenceRaw) {
		if ch.name == name {
			return unescapeXMLText(elementTextContent(ch.raw))
		}
	}
	return ""
}

func presencePriority(presenceRaw []byte) (*int, error) {
	for _, ch := range parseDirectChildElements(presenceRaw) {
		if ch.name != "priority" {
			continue
		}
		text := strings.TrimSpace(elementTextContent(ch.raw))
		if text == "" {
			return nil, nil
		}
		p, err := strconv.Atoi(text)
		if err != nil {
			return nil, fmt.Errorf("protocol: invalid presence priority %q", text)
		}
		return &p, nil
	}
	return nil, nil
}
