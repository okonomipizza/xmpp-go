package protocol

import (
	"errors"

	"github.com/okonomipizza/xmpp-go/jid"
)

const rosterNS = "jabber:iq:roster"

// RosterItem は roster の 1 連絡先を表す (RFC 6121 Section 2)。
type RosterItem struct {
	JID          jid.JID
	Name         string
	Subscription string
	Groups       []string
}

// RosterQueryGetBytes は roster get 用の <query/> 子要素を返す。
func RosterQueryGetBytes() []byte {
	return []byte("<query xmlns='" + rosterNS + "'/>")
}

// RosterIQGetBytes は RFC 6121 Example 1 の roster get IQ を返す。
func RosterIQGetBytes(from jid.JID, id string) ([]byte, error) {
	return IQStanzaBytes(from, jid.JID{}, id, "get", RosterQueryGetBytes())
}

// ParseRosterItems は IQ の Payload から roster item を抽出する。
// Payload は <query xmlns='jabber:iq:roster'> 要素、またはその親要素の inner でもよい。
func ParseRosterItems(iqPayload []byte) ([]RosterItem, error) {
	if isRosterQueryElement(iqPayload) {
		_, selfClosing, ok := findOpenTagEnd(iqPayload, 0)
		if !ok {
			return nil, errors.New("protocol: invalid roster query")
		}
		if selfClosing {
			return nil, nil
		}
		return parseRosterQueryItems(iqPayload)
	}
	for _, ch := range parseDirectChildElements(iqPayload) {
		if ch.name == "query" && elementAttrValue(ch.raw, "xmlns") == rosterNS {
			return parseRosterQueryItems(ch.raw)
		}
	}
	return nil, errors.New("protocol: roster query not found")
}

func isRosterQueryElement(raw []byte) bool {
	if len(raw) == 0 || raw[0] != '<' {
		return false
	}
	openEnd, _, ok := findOpenTagEnd(raw, 0)
	if !ok {
		return false
	}
	return openTagLocalName(raw[:openEnd+1]) == "query" && elementAttrValue(raw, "xmlns") == rosterNS
}

func parseRosterQueryItems(queryRaw []byte) ([]RosterItem, error) {
	var items []RosterItem
	for _, ch := range parseDirectChildElements(queryRaw) {
		if ch.name != "item" {
			continue
		}
		item, err := parseRosterItem(ch.raw)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func parseRosterItem(itemRaw []byte) (RosterItem, error) {
	jidStr := elementAttrValue(itemRaw, "jid")
	if jidStr == "" {
		return RosterItem{}, errors.New("protocol: roster item missing jid")
	}
	j, err := jid.Parse(jidStr)
	if err != nil {
		return RosterItem{}, err
	}
	var groups []string
	for _, ch := range parseDirectChildElements(itemRaw) {
		if ch.name == "group" {
			text := unescapeXMLText(elementTextContent(ch.raw))
			if text != "" {
				groups = append(groups, text)
			}
		}
	}
	return RosterItem{
		JID:          j,
		Name:         elementAttrValue(itemRaw, "name"),
		Subscription: elementAttrValue(itemRaw, "subscription"),
		Groups:       groups,
	}, nil
}
